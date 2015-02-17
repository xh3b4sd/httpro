package transport

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
	ErrStatusCode5XX  = errgo.New("error status code 5XX")

	Mask = errgo.MaskFunc(
		IsErrConnectTimeout,
		IsErrRequestTimeout,
		IsErrConnectRefused,
		IsErrStatusCode5XX,
	)
)

// TODO
func IsErrConnectTimeout(err error) bool {
	return false
}

func IsErrRequestTimeout(err error) bool {
	errCause := errgo.Cause(err)

	if urlErr, ok := errCause.(*url.Error); ok {
		return errgo.Cause(urlErr.Err) == ErrRequestTimeout
	}

	return errCause == ErrRequestTimeout
}

func IsErrConnectRefused(err error) bool {
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

func IsErrStatusCode5XX(err error) bool {
	return errgo.Cause(err) == ErrStatusCode5XX
}
