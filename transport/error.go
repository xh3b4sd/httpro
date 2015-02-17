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
	if urlErr, ok := err.(*url.Error); ok {
		if isCustomErrErrRequestTimeout(urlErr.Err) {
			return true
		}
	}

	if isCustomErrErrRequestTimeout(err) {
		return true
	}

	return false
}

func IsErrConnectRefused(err error) bool {
	if isUrlErrErrConnectRefused(err) {
		return true
	}

	if isOpErrErrConnectRefused(err) {
		return true
	}

	if isCustomErrErrConnectRefused(err) {
		return true
	}

	return false
}

func IsErrStatusCode5XX(err error) bool {
	return errgo.Cause(err) == ErrStatusCode5XX
}

//------------------------------------------------------------------------------
// private

func isUrlErrErrConnectRefused(err error) bool {
	errCause := errgo.Cause(err)

	if urlErr, ok := errCause.(*url.Error); ok {
		if isOpErrErrConnectRefused(urlErr.Err) {
			return true
		}

		if isCustomErrErrConnectRefused(urlErr.Err) {
			return true
		}
	}

	return false
}

func isOpErrErrConnectRefused(err error) bool {
	errCause := errgo.Cause(err)

	if opErr, ok := errCause.(*net.OpError); ok {
		if isCustomErrErrConnectRefused(opErr.Err) {
			return true
		}

		if isSyscallErrErrConnectRefused(opErr.Err) {
			return true
		}
	}

	return false
}

func isSyscallErrErrConnectRefused(err error) bool {
	errCause := errgo.Cause(err)

	if errno, ok := errCause.(syscall.Errno); ok {
		if errno == syscall.ECONNREFUSED {
			return true
		}
	}

	return false
}

func isCustomErrErrConnectRefused(err error) bool {
	return errgo.Cause(err) == ErrStatusCode5XX
}

func isCustomErrErrRequestTimeout(err error) bool {
	return errgo.Cause(err) == ErrRequestTimeout
}
