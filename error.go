package httpro

import (
	"net"
	"net/url"
	"syscall"

	"github.com/juju/errgo"
)

var (
	ErrConnectTimeout = errgo.New("connect timeout")
	ErrRequestTimeout = errgo.New("request timeout")

	Mask = errgo.MaskFunc(IsErrConnectTimeout, IsErrRequestTimeout, IsErrConnectionRefused)
)

func IsErrConnectTimeout(err error) bool {
	return false
}

func IsErrRequestTimeout(err error) bool {
	errCause := errgo.Cause(err)

	if urlErr, ok := errCause.(*url.Error); ok {
		return errgo.Cause(urlErr.Err) == ErrRequestTimeout
	}

	return errgo.Cause(errCause) == ErrRequestTimeout
}

func IsErrConnectionRefused(err error) bool {
	errCause := errgo.Cause(err)

	if urlErr, ok := errCause.(*url.Error); ok {
		if opErr, ok := urlErr.Err.(*net.OpError); ok {
			if errno, ok := opErr.Err.(syscall.Errno); ok {
				if errno == syscall.ECONNREFUSED {
					return true
				}
			}
		}
	}

	return false
}
