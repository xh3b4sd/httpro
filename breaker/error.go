package breaker

import (
	"github.com/juju/errgo"
)

var (
	ErrMaxConcurrentActionsReached = errgo.New("max concurrent actions reached")
	ErrMaxErrorRateReached         = errgo.New("max error rate reached")
	ErrMaxPerformanceLossReached   = errgo.New("max performance loss reached")
	ErrNilAction                   = errgo.New("cannot call nil action")

	Mask = errgo.MaskFunc(
		errgo.Any,
		IsErrBreakerError,
	)
)

// IsErrBreakerError checks whether the given error is caused by the breaker or
// not.
func IsErrBreakerError(err error) bool {
	return IsErrMaxConcurrentActionsReached(err) ||
		IsErrMaxErrorRateReached(err) ||
		IsErrMaxPerformanceLossReached(err)
}

// IsErrMaxConcurrentActionsReached checks whether the given error is caused by
// the MaxConcurrentActions configuration or not.
func IsErrMaxConcurrentActionsReached(err error) bool {
	return errgo.Cause(err) == ErrMaxConcurrentActionsReached
}

// IsErrMaxErrorRateReached checks whether the given error is caused by the
// MaxErrorRate configuration or not.
func IsErrMaxErrorRateReached(err error) bool {
	return errgo.Cause(err) == ErrMaxErrorRateReached
}

// IsErrMaxPerformanceLossReached checks whether the given error is caused by
// the MaxPerformanceLoss configuration or not.
func IsErrMaxPerformanceLossReached(err error) bool {
	return errgo.Cause(err) == ErrMaxPerformanceLossReached
}
