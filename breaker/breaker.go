package breaker

import (
	"sync"
	"time"

	"github.com/op/go-logging"
	"github.com/zyndiecate/httpro/logger"
)

var (
	DefaultMaxConcurrencyLimit uint = 20
	DefaultBreakTTL                 = 10 * time.Second
	DefaultMaxErrorRate        uint = 10
	DefaultMaxPerformanceLoss  uint = 25

	defaultSampleTTL    = 2 * time.Second
	defaultMinSampleVol = 10
)

type Config struct {
	// MaxConcurrencyLimit describes the maximum number of concurrent actions
	// that are allowed to happen. In case there are more actions than this value
	// allows, the breaker stops to accept new actions.
	MaxConcurrencyLimit uint

	// BreakTTL describes the amount of time no new actions will be accepted
	// after the breaker stops to accept new actions.
	BreakTTL time.Duration

	// MaxErrorRate describes the number of errors in percentage allowed to happen
	// before the breaker stops to accept new actions.
	MaxErrorRate uint

	// MaxPerformanceLoss describes the percentage of action performance allowed to
	// loose before the breaker stops to accept new actions.
	MaxPerformanceLoss uint

	// LogLevel defines the log level used to log process information. If none is
	// given, logging is disabled. See
	// https://godoc.org/github.com/op/go-logging#Level.
	LogLevel string
}

type state struct {
	concurrencyLimit int64
	errorRate        int64
	performanceLoss  int64
}

type Breaker struct {
	Config  Config
	state   state
	mutex   *sync.Mutex
	samples []*sample
	logger  *logging.Logger
}

func NewBreaker(c Config) *Breaker {
	if c.MaxConcurrencyLimit == 0 {
		c.MaxConcurrencyLimit = DefaultMaxConcurrencyLimit
	}

	if c.BreakTTL == 0 {
		c.BreakTTL = DefaultBreakTTL
	}

	if c.MaxErrorRate == 0 {
		c.MaxErrorRate = DefaultMaxErrorRate
	}

	if c.MaxPerformanceLoss == 0 {
		c.MaxPerformanceLoss = DefaultMaxPerformanceLoss
	}

	b := &Breaker{
		Config:  c,
		mutex:   &sync.Mutex{},
		samples: []*sample{},
		logger:  logger.NewLogger(logger.Config{Name: "breaker", Level: c.LogLevel}),
	}

	go b.trackState()

	b.logger.Debug("created breaker with config: %#v", c)

	return b
}

// Run executes action in case the breaker still accept new actions.
//
//   b.Run(func() error {
//     // do whatever you want and return an error in case bad things happened
//     return err
//   })
//
func (b *Breaker) Run(action func() error) error {
	if action == nil {
		return Mask(ErrNilAction)
	}

	if err := b.accept(); err != nil {
		return Mask(err)
	}

	cs := b.currentSample()

	var actionEnd int64
	actionStart := cs.actionStart()
	if err := action(); err != nil {
		actionEnd = cs.actionFailure()
		return Mask(err)
	} else {
		actionEnd = cs.actionSuccess()
	}

	cs.actionPerformance(actionStart, actionEnd)

	return nil
}

//------------------------------------------------------------------------------
// private

func (b *Breaker) newSample() *sample {
	return newSample(sampleConfig{mutex: b.mutex})
}

func (b *Breaker) trackState() {
	for {
		time.Sleep(defaultSampleTTL)

		// add new sample
		b.samples = append(b.samples, b.newSample())

		// remove old sample
		if len(b.samples) > defaultMinSampleVol {
			b.samples = b.samples[1:len(b.samples)]
		}

		if len(b.samples) < defaultMinSampleVol {
			continue
		}

		if err := b.accept(); err != nil {
			b.logger.Error("no new action accepted for %s: %s", b.Config.BreakTTL.String(), err.Error())
			time.Sleep(b.Config.BreakTTL)
		}

		cs := b.currentSample()

		// calculate concurrent actions
		b.state.concurrencyLimit = cs.concurrencyLimit

		// calculate error rate
		b.state.errorRate = b.calculateErrorRate()

		// calculate performance loss
		b.state.performanceLoss = b.calculatePerformanceLoss()

		b.logger.Debug("breaker state: %+v\n", b.state)
	}
}

func (b *Breaker) accept() error {
	if b.state.concurrencyLimit > int64(b.Config.MaxConcurrencyLimit) {
		return Mask(ErrMaxConcurrencyLimitReached)
	}

	if b.state.errorRate > int64(b.Config.MaxErrorRate) {
		return Mask(ErrMaxErrorRateReached)
	}

	if b.state.performanceLoss > int64(b.Config.MaxPerformanceLoss) {
		return Mask(ErrMaxPerformanceLossReached)
	}

	return nil
}

func (b *Breaker) currentSample() *sample {
	if len(b.samples) == 0 {
		return b.newSample()
	}

	return b.samples[len(b.samples)-1]
}

func (b *Breaker) calculatePerformanceLoss() int64 {
	histPerfAvg := calculatePerformanceAvg(b.samples[:len(b.samples)-1])
	currPerfAvg := calculatePerformanceAvg(b.samples[len(b.samples)-2 : len(b.samples)-1])

	if histPerfAvg == 0 || currPerfAvg == 0 {
		return 0
	}

	currPerfLost := (currPerfAvg * 100 / histPerfAvg) - 100

	return currPerfLost
}

func (b *Breaker) calculateErrorRate() int64 {
	var totalActions int64
	var totalFailures int64

	for _, s := range b.samples {
		totalActions += s.totalActions
		totalFailures += s.totalFailures
	}

	if totalActions == 0 || totalFailures == 0 {
		return 0
	}

	currErrorRate := totalFailures * 100 / totalActions

	return currErrorRate
}

func calculatePerformanceAvg(ss []*sample) int64 {
	var performanceCount int64
	var performanceSum int64

	for _, s := range ss {
		performanceCount += int64(len(s.performances))

		for _, p := range s.performances {
			performanceSum += p
		}
	}

	if performanceCount == 0 || performanceSum == 0 {
		return 0
	}

	performanceAvg := performanceSum / performanceCount

	return performanceAvg
}
