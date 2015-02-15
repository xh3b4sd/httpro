package transport

import (
	"io"
	"net/http"
	"time"
)

type context struct {
	requestTimedOut bool

	req   *http.Request
	res   *http.Response
	timer *time.Timer
}

func (ctx *context) Close() {
	if ctx.timer != nil {
		ctx.timer.Stop()
	}
}

type bodyCloser struct {
	io.ReadCloser
	timer *time.Timer
}

func (bc *bodyCloser) Close() error {
	bc.timer.Stop()
	return bc.ReadCloser.Close()
}

func isStatusCode5XX(statusCode int) bool {
	return statusCode >= 500 && statusCode <= 599
}
