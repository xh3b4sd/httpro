package breaker

import (
	"github.com/juju/errgo"
)

var (
	ErrMaxConcurrencyLimitExceeded = errgo.New("max concurrent actions exceeded")
	ErrMaxErrorRateExceeded        = errgo.New("max error rate exceeded")
	ErrMaxPerformanceLossExceeded  = errgo.New("max performance loss exceeded")
	ErrNilAction                   = errgo.New("cannot call nil action")

	Mask = errgo.MaskFunc(
		errgo.Any,
		IsErrBreakerError,
	)
)

// IsErrBreakerError checks whether the given error is caused by the breaker or
// not.
func IsErrBreakerError(err error) bool {
	return IsErrMaxConcurrencyLimitExceeded(err) ||
		IsErrMaxErrorRateExceeded(err) ||
		IsErrMaxPerformanceLossExceeded(err)
}

// IsErrMaxConcurrencyLimitExceeded checks whether the given error is caused by
// the MaxConcurrencyLimit configuration or not.
func IsErrMaxConcurrencyLimitExceeded(err error) bool {
	return errgo.Cause(err) == ErrMaxConcurrencyLimitExceeded
}

// IsErrMaxErrorRateExceeded checks whether the given error is caused by the
// MaxErrorRate configuration or not.
func IsErrMaxErrorRateExceeded(err error) bool {
	return errgo.Cause(err) == ErrMaxErrorRateExceeded
}

// IsErrMaxPerformanceLossExceeded checks whether the given error is caused by
// the MaxPerformanceLoss configuration or not.
func IsErrMaxPerformanceLossExceeded(err error) bool {
	return errgo.Cause(err) == ErrMaxPerformanceLossExceeded
}
