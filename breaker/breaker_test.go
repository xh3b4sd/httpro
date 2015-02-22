package breaker

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBreaker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "breaker")
}

var _ = Describe("breaker", func() {
	var (
		err         error
		executed    bool
		testSamples []*sample
	)

	BeforeEach(func() {
		err = nil
		executed = false
		testSamples = nil
	})

	Describe("simple Run()", func() {
		BeforeEach(func() {
			b := newTestBreaker(Config{SampleTTL: 1})
			err = b.Run(newTestAction(&executed))
		})

		It("should not return error", func() {
			Expect(err).To(BeNil())
		})

		It("should execute action", func() {
			Expect(executed).To(BeTrue())
		})
	})

	Describe("MaxConcurrencyLimit", func() {
		Describe("limit not reached", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{SampleTTL: 1})
				mockTestSample(b, testSampleConfig{actionStart: int(DefaultMaxConcurrencyLimit) - 1})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should not return error", func() {
				Expect(err).To(BeNil())
			})

			It("should execute action", func() {
				Expect(executed).To(BeTrue())
			})
		})

		Describe("limit reached", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{SampleTTL: 1})
				mockTestSample(b, testSampleConfig{actionStart: int(DefaultMaxConcurrencyLimit)})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should not return error", func() {
				Expect(err).To(BeNil())
			})

			It("should execute action", func() {
				Expect(executed).To(BeTrue())
			})
		})

		Describe("limit exceeded", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{SampleTTL: 1})
				mockTestSample(b, testSampleConfig{actionStart: int(DefaultMaxConcurrencyLimit) + 1})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should return error", func() {
				Expect(IsErrBreakerError(err)).To(BeTrue())
				Expect(IsErrMaxConcurrencyLimitExceeded(err)).To(BeTrue())
			})

			It("should not execute action", func() {
				Expect(executed).To(BeFalse())
			})
		})
	})

	Describe("MaxErrorRate", func() {
		Describe("rate not reached", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{MaxConcurrencyLimit: 100, SampleTTL: 1})
				mockTestSample(b, testSampleConfig{actionFailure: int(DefaultMaxErrorRate) - 1})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should not return error", func() {
				Expect(err).To(BeNil())
			})

			It("should execute action", func() {
				Expect(executed).To(BeTrue())
			})
		})

		Describe("rate reached", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{MaxConcurrencyLimit: 100, SampleTTL: 1})
				mockTestSample(b, testSampleConfig{actionFailure: int(DefaultMaxErrorRate)})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should not return error", func() {
				Expect(err).To(BeNil())
			})

			It("should execute action", func() {
				Expect(executed).To(BeTrue())
			})
		})

		Describe("rate exceeded", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{MaxConcurrencyLimit: 100, SampleTTL: 1})
				mockTestSample(b, testSampleConfig{actionFailure: int(DefaultMaxErrorRate) + 1})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should return error", func() {
				Expect(IsErrBreakerError(err)).To(BeTrue())
				Expect(IsErrMaxErrorRateExceeded(err)).To(BeTrue())
			})

			It("should not execute action", func() {
				Expect(executed).To(BeFalse())
			})
		})
	})

	Describe("MaxPerformanceLoss", func() {
		Describe("loss not reached", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{SampleTTL: 1})
				mockTestSample(b, testSampleConfig{performanceLoss: int(DefaultMaxPerformanceLoss) - 1})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should not return error", func() {
				Expect(err).To(BeNil())
			})

			It("should execute action", func() {
				Expect(executed).To(BeTrue())
			})
		})

		Describe("loss reached", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{SampleTTL: 1})
				mockTestSample(b, testSampleConfig{performanceLoss: int(DefaultMaxPerformanceLoss)})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should not return error", func() {
				Expect(err).To(BeNil())
			})

			It("should execute action", func() {
				Expect(executed).To(BeTrue())
			})
		})

		Describe("loss exceeded", func() {
			BeforeEach(func() {
				b := newTestBreaker(Config{SampleTTL: 1})
				mockTestSample(b, testSampleConfig{performanceLoss: int(DefaultMaxPerformanceLoss) + 1})
				b.trackState()
				err = b.Run(newTestAction(&executed))
			})

			It("should return error", func() {
				Expect(IsErrBreakerError(err)).To(BeTrue())
				Expect(IsErrMaxPerformanceLossExceeded(err)).To(BeTrue())
			})

			It("should not execute action", func() {
				Expect(executed).To(BeFalse())
			})
		})
	})
})
