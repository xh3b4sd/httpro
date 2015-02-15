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
		IsErrMaxConcurrentReached,
		IsErrMaxErrorRateReached,
		IsErrMaxPerformanceLossReached,
	)
)

func IsErrMaxConcurrentReached(err error) bool {
	return errgo.Cause(err) == ErrMaxConcurrentActionsReached
}

func IsErrMaxErrorRateReached(err error) bool {
	return errgo.Cause(err) == ErrMaxErrorRateReached
}

func IsErrMaxPerformanceLossReached(err error) bool {
	return errgo.Cause(err) == ErrMaxPerformanceLossReached
}
