package breaker

import (
	"fmt"
	"time"
)

var (
	DefaultMaxConcurrentActions uint = 20
	DefaultBreakTTL                  = 10 * time.Second
	DefaultMaxErrorRate         uint = 10
	DefaultMaxPerformanceLoss   uint = 25

	defaultSampleTTL    = 2 * time.Second
	defaultMinSampleVol = 10
)

type Config struct {
	// MaxConcurrentActions describes the maximum number of concurrent actions
	// that are allowed to happen. In case there are more actions than this value
	// allows, the breaker stops to accept new actions.
	MaxConcurrentActions uint

	// BreakTTL describes the amount of time no new actions will be accepted
	// after the breaker stops to accept new actions.
	BreakTTL time.Duration

	// MaxErrorRate describes the number of errors in percentage allowed to happen
	// before the breaker stops to accept new actions.
	MaxErrorRate uint

	// MaxPerformanceLoss describes the percentage of action performance allowed to
	// loose before the breaker stops to accept new actions.
	MaxPerformanceLoss uint
}

type state struct {
	concurrentActions int64
	errorRate         int64
	performanceLoss   int64
}

type Breaker struct {
	Config  Config
	state   state
	samples []*sample
}

func NewBreaker(c Config) *Breaker {
	if c.MaxConcurrentActions == 0 {
		c.MaxConcurrentActions = DefaultMaxConcurrentActions
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
		Config: c,
		samples: []*sample{
			&sample{},
		},
	}

	go b.trackState()

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

func (b *Breaker) trackState() {
	for {
		time.Sleep(defaultSampleTTL)

		if len(b.samples) < defaultMinSampleVol {
			continue
		}

		if err := b.accept(); err != nil {
			time.Sleep(b.Config.BreakTTL)
		}

		// calculate concurrent actions
		b.state.concurrentActions = b.currentSample().concurrentActions

		// calculate error rate
		b.state.errorRate = b.calculateErrorRate()

		// calculate performance loss
		b.state.performanceLoss = b.calculatePerformanceLoss()

		// add new sample
		b.samples = append(b.samples, &sample{})

		// remove old sample
		if len(b.samples) > defaultMinSampleVol {
			b.samples = b.samples[1:len(b.samples)]
		}

		fmt.Printf("%#v\n", "breaker state")
		fmt.Printf("%#v\n", b.state)
		fmt.Printf("%#v\n", "")

		fmt.Printf("%#v\n", "breaker samples")
		fmt.Printf("%#v\n", b.samples)
		fmt.Printf("%#v\n", "")
	}
}

func (b *Breaker) accept() error {
	if b.state.concurrentActions > int64(b.Config.MaxConcurrentActions) {
		return Mask(ErrMaxConcurrentActionsReached)
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
	return b.samples[len(b.samples)-1]
}

func (b *Breaker) calculatePerformanceLoss() int64 {
	histPerfAvg := calculatePerformanceAvg(b.samples[:len(b.samples)-1])
	currPerfAvg := calculatePerformanceAvg(b.samples[len(b.samples)-1:])

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
	currErrorRate := (totalFailures * 100 / totalActions) - 100

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

	performanceAvg := performanceSum / performanceCount

	return performanceAvg
}
