package httpro

import (
	"net"
	"net/http"
	"time"
)

var (
	defaultRequestTimeout = 30 * time.Second
	defaultConnectTimeout = 30 * time.Second
	defaultReconnectDelay = 200 * time.Millisecond
	defaultConnectRetry   = uint(2)
	defaultRequestRetry   = uint(1)
)

type TransportConfig struct {
	ReconnectDelay time.Duration
	ConnectTimeout time.Duration
	RequestTimeout time.Duration

	RequestRetry uint
}

type transport struct {
	Config           TransportConfig
	defaultTransport *http.Transport
}

func NewTransport(c TransportConfig) http.RoundTripper {
	if c.ReconnectDelay == 0 {
		c.ReconnectDelay = defaultReconnectDelay
	}

	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}

	if c.RequestTimeout == 0 {
		c.RequestTimeout = defaultRequestTimeout
	}

	if c.RequestRetry == 0 {
		c.RequestRetry = defaultRequestRetry
	}

	t := &transport{
		Config: c,
	}

	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		t.defaultTransport = defaultTransport

		t.defaultTransport.Dial = func(network, addr string) (net.Conn, error) {
			var err error
			var conn net.Conn

			for i := 0; i < int(defaultConnectRetry); i++ {
				conn, err = net.DialTimeout(network, addr, t.Config.ConnectTimeout)

				if IsErrConnectionRefused(err) {
					time.Sleep(t.Config.ReconnectDelay)
					continue
				} else if err != nil {
					return nil, Mask(err)
				}
			}

			return conn, Mask(err)
		}
	}

	return t
}

func (t *transport) CancelRequest(req *http.Request) {
	t.defaultTransport.CancelRequest(req)
}

func (t *transport) CloseIdleConnections() {
	t.defaultTransport.CloseIdleConnections()
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error
	var res *http.Response

	for i := 0; i < int(t.Config.RequestRetry); i++ {
		res, err = t.roundTrip(req)
		if IsErrRequestTimeout(err) {
			continue
		} else if err != nil {
			return nil, err
		}
	}

	return res, err
}

//------------------------------------------------------------------------------
// private

func (t *transport) roundTrip(req *http.Request) (*http.Response, error) {
	ctx := &context{
		req: req,
	}

	// pre response
	if err := t.preResHandler(ctx); err != nil {
		ctx.Close()
		return nil, Mask(err)
	}

	var err error
	ctx.res, err = t.defaultTransport.RoundTrip(ctx.req)
	if err != nil {
		ctx.Close()

		if ctx.requestTimedOut {
			return nil, Mask(ErrRequestTimeout)
		}

		return nil, Mask(err)
	}

	// post response
	if err := t.postResHandler(ctx); err != nil {
		ctx.Close()
		return nil, Mask(err)
	}

	return ctx.res, nil
}

func (t *transport) preResHandler(ctx *context) error {
	ctx.timer = time.AfterFunc(t.Config.RequestTimeout, func() {
		ctx.requestTimedOut = true
		t.defaultTransport.CancelRequest(ctx.req)
	})

	return nil
}

func (t *transport) postResHandler(ctx *context) error {
	if ctx.requestTimedOut {
		ctx.res.Body = &bodyCloser{ReadCloser: ctx.res.Body, timer: ctx.timer}
	}

	return nil
}
