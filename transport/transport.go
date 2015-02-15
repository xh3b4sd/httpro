package transport

import (
	"net"
	"net/http"
	"time"

	"github.com/zyndiecate/httpro/breaker"
)

var (
	defaultRequestTimeout = 30 * time.Second
	defaultConnectTimeout = 30 * time.Second
	defaultReconnectDelay = 200 * time.Millisecond
	defaultConnectRetry   = uint(2)
	defaultRequestRetry   = uint(1)
)

type Config struct {
	// ReconnectDelay describes the time the client should block before
	// attempting to connect again. This configuration is used when the
	// preceeding connection was refused.
	ReconnectDelay time.Duration

	// ConnectTimeout is the value used to configure the call to net.DialTimeout.
	ConnectTimeout time.Duration

	// RequestTimeout describes the time within a request needs to be processed
	// before it is canceled.
	RequestTimeout time.Duration

	// RequestRetry describes the number of retries the client will do in case a
	// request timed out or received a 5XX status code.
	RequestRetry uint

	// BreakerConfig is used to configure the breaker internally used as circuit
	// breaker.
	BreakerConfig breaker.Config
}

type Transport struct {
	Config           Config
	Breaker          *breaker.Breaker
	defaultTransport *http.Transport
}

// NewTransport creates a new *http.Transport that implements http.RoundTripper.
func NewTransport(c Config) http.RoundTripper {
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

	t := &Transport{
		Config:  c,
		Breaker: breaker.NewBreaker(c.BreakerConfig),
	}

	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		t.defaultTransport = defaultTransport

		t.defaultTransport.Dial = func(network, addr string) (net.Conn, error) {
			var err error
			var conn net.Conn

			for i := 0; i < int(defaultConnectRetry); i++ {
				conn, err = net.DialTimeout(network, addr, t.Config.ConnectTimeout)

				if IsErrConnectRefused(err) {
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

func (t *Transport) CancelRequest(req *http.Request) {
	t.defaultTransport.CancelRequest(req)
}

func (t *Transport) CloseIdleConnections() {
	t.defaultTransport.CloseIdleConnections()
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error
	var res *http.Response

	err = t.Breaker.Run(func() error {
		for i := 0; i < int(t.Config.RequestRetry); i++ {
			err = nil

			res, err = t.roundTrip(req)
			if err != nil {
				continue
			}

			if isStatusCode5XX(res.StatusCode) {
				// We just want to track 5XX errors for the breaker. Later we must
				// reset this error.
				err = Err5XX
				continue
			}
		}

		return Mask(err)
	})

	// Because we just want to track 5XX errors for the breaker, we reset those
	// errors here. Otherwise the roundtripper would not be able to handle the
	// response properly. In case the roundtripper replies both, a response and
	// an error, the response will be ignored.
	if IsErr5XX(err) {
		err = nil
	}

	return res, Mask(err)
}

//------------------------------------------------------------------------------
// private

func (t *Transport) roundTrip(req *http.Request) (*http.Response, error) {
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

func (t *Transport) preResHandler(ctx *context) error {
	ctx.timer = time.AfterFunc(t.Config.RequestTimeout, func() {
		ctx.requestTimedOut = true
		t.defaultTransport.CancelRequest(ctx.req)
	})

	return nil
}

func (t *Transport) postResHandler(ctx *context) error {
	if ctx.requestTimedOut {
		ctx.res.Body = &bodyCloser{ReadCloser: ctx.res.Body, timer: ctx.timer}
	}

	return nil
}
