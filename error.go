package httpro

import (
	"net"
	"net/url"
	"syscall"

	"github.com/juju/errgo"
)

var (
	ErrConnectRefused = errgo.New("connect refused")
	ErrConnectTimeout = errgo.New("connect timeout")
	ErrRequestTimeout = errgo.New("request timeout")

	Mask = errgo.MaskFunc(IsErrConnectTimeout, IsErrRequestTimeout, IsErrConnectionRefused)
)

// TODO
func IsErrConnectTimeout(err error) bool {
	return false
}

func IsErr5XX(statusCode int) bool {
	return statusCode >= 500 && statusCode <= 599
}

func IsErrRequestTimeout(err error) bool {
	errCause := errgo.Cause(err)

	if urlErr, ok := errCause.(*url.Error); ok {
		return errgo.Cause(urlErr.Err) == ErrRequestTimeout
	}

	return errCause == ErrRequestTimeout
}

func IsErrConnectionRefused(err error) bool {
	errCause := errgo.Cause(err)

	if urlErr, ok := errCause.(*url.Error); ok {
		urlErrCause := errgo.Cause(urlErr.Err)

		if urlErrCause == ErrConnectRefused {
			return true
		}

		if opErr, ok := urlErrCause.(*net.OpError); ok {
			if errno, ok := opErr.Err.(syscall.Errno); ok {
				if errno == syscall.ECONNREFUSED {
					return true
				}
			}
		}
	}

	if opErr, ok := errCause.(*net.OpError); ok {
		opErrCause := errgo.Cause(opErr.Err)

		if opErrCause == ErrConnectRefused {
			return true
		}

		if errno, ok := opErrCause.(syscall.Errno); ok {
			if errno == syscall.ECONNREFUSED {
				return true
			}
		}
	}

	return errCause == ErrConnectRefused
}
