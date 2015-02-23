package breaker

func newTestAction(executed *bool) func() error {
	return func() error {
		*executed = true
		return nil
	}
}

func newTestBreaker(c Config) *Breaker {
	b := NewBreaker(c)
	b.samples = newTestSamples(b)
	return b
}

// newTestSamples returns a collection of 10 samples, that is needed to
// calculate breaker metrics used to descide to break or not.
func newTestSamples(b *Breaker) []*sample {
	s := []*sample{}

	for i := 0; i < 10; i++ {
		s = append(s, b.newSample())
	}

	return s
}

type testSampleConfig struct {
	actionStart     int
	actionFailure   int
	performanceLoss int
}

// mockTestSample manipulates the forelast sample of the given breaker for to
// the upcoming trackState call.
func mockTestSample(b *Breaker, c testSampleConfig) {
	newestSample := b.samples[len(b.samples)-1]

	for i := 0; i < c.actionStart; i++ {
		newestSample.actionStart()
	}

	if c.actionFailure > 0 {
		for i := 1; i <= 100; i++ {
			newestSample.actionStart()
		}
		for i := 1; i <= c.actionFailure; i++ {
			newestSample.actionFailure()
		}
	}

	if c.performanceLoss > 0 {
		// avg performance of 100
		oldestSample := b.samples[0]
		oldestSample.performances = []int64{100, 120, 80, 70, 130}

		// avg performance of 100
		foreNewestSample := b.samples[len(b.samples)-2]
		foreNewestSample.performances = []int64{100, 120, 80, 50, 150}

		// performance loss
		newestSample.performances = []int64{
			int64(100 + c.performanceLoss),
			int64(120 + c.performanceLoss),
			int64(80 + c.performanceLoss),
			int64(50 + c.performanceLoss),
			int64(150 + c.performanceLoss),
		}
	}
}
