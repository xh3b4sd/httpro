package transport

import (
	"net"
	"net/http"
	"time"

	"github.com/op/go-logging"
	"github.com/zyndiecate/httpro/breaker"
	"github.com/zyndiecate/httpro/logger"
)

var (
	DefaultRequestTimeout = 30 * time.Second
	DefaultConnectTimeout = 30 * time.Second

	DefaultConnectRetryDelay = 200 * time.Millisecond
	DefaultRequestRetryDelay = 200 * time.Millisecond

	DefaultConnectRetry = uint(2)
	DefaultRequestRetry = uint(2)
)

// Config to configure the HTTP transport.
type Config struct {
	// ConnectRetryDelay describes the time the client should block before
	// attempting to connect again. This configuration is used when the
	// preceeding connection was refused.
	ConnectRetryDelay time.Duration

	// RequestRetryDelay describes the time the client should block before
	// attempting to try a request again. This configuration is used when the
	// preceeding connection was failed in terms of status code 5XX or timed out.
	RequestRetryDelay time.Duration

	// ConnectTimeout is the value used to configure the call to net.DialTimeout.
	ConnectTimeout time.Duration

	// RequestTimeout describes the time within a request needs to be processed
	// before it is canceled.
	RequestTimeout time.Duration

	// ConnectRetry describes the number of retries the client will do in case a
	// connection was refused.
	ConnectRetry uint

	// RequestRetry describes the number of retries the client will do in case a
	// request timed out or received a 5XX status code.
	RequestRetry uint

	// BreakerConfig is used to configure the breaker internally used as circuit
	// breaker.
	BreakerConfig breaker.Config

	// LogLevel defines the log level used to log process information. If none is
	// given, logging is disabled. See
	// https://godoc.org/github.com/op/go-logging#Level.
	LogLevel string
}

type Transport struct {
	Config           Config
	Breaker          *breaker.Breaker
	defaultTransport *http.Transport
	logger           *logging.Logger
}

// NewTransport creates a new *http.Transport that implements http.RoundTripper.
func NewTransport(c Config) http.RoundTripper {
	if c.ConnectRetryDelay == 0 {
		c.ConnectRetryDelay = DefaultConnectRetryDelay
	}

	if c.RequestRetryDelay == 0 {
		c.RequestRetryDelay = DefaultRequestRetryDelay
	}

	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = DefaultConnectTimeout
	}

	if c.RequestTimeout == 0 {
		c.RequestTimeout = DefaultRequestTimeout
	}

	if c.ConnectRetry == 0 {
		c.ConnectRetry = DefaultConnectRetry
	}

	if c.RequestRetry == 0 {
		c.RequestRetry = DefaultRequestRetry
	}

	if c.BreakerConfig.LogLevel == "" {
		c.BreakerConfig.LogLevel = c.LogLevel
	}

	t := &Transport{
		Config:  c,
		Breaker: breaker.NewBreaker(c.BreakerConfig),
		logger:  logger.NewLogger(logger.Config{Name: "transport", Level: c.LogLevel}),
	}

	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		t.defaultTransport = defaultTransport
		// TODO configure
		t.defaultTransport.DisableKeepAlives = true
		t.defaultTransport.Dial = t.DialFunc
	}

	t.logger.Debug("created http transport with config: %#v", c)

	return t
}

func (t *Transport) CancelRequest(req *http.Request) {
	t.defaultTransport.CancelRequest(req)
}

func (t *Transport) CloseIdleConnections() {
	t.defaultTransport.CloseIdleConnections()
}

func (t *Transport) DialFunc(network, addr string) (net.Conn, error) {
	var err error
	var conn net.Conn

	for i := 0; i < int(t.Config.ConnectRetry); i++ {
		conn, err = net.DialTimeout(network, addr, t.Config.ConnectTimeout)

		if IsErrConnectRefused(err) {
			time.Sleep(t.Config.ConnectRetryDelay)
			t.logger.Info("retry connection to %s", addr)
			continue
		} else if err != nil {
			return nil, Mask(err)
		}
	}

	return conn, Mask(err)
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error
	var res *http.Response

	err = t.Breaker.Run(func() error {
		for i := 0; i <= int(t.Config.RequestRetry); i++ {
			if i > 0 {
				time.Sleep(t.Config.RequestRetryDelay)
				t.logger.Info("retry request to %s", req.URL.String())
			}

			res, err = t.roundTrip(req)
			if err != nil {
				continue
			}

			if isStatusCode5XX(res.StatusCode) {
				// We just want to track 5XX errors for the breaker. Later we must
				// reset this error.
				err = ErrStatusCode5XX
				continue
			}
		}

		return Mask(err)
	})

	// Because we just want to track 5XX errors for the breaker, we reset those
	// errors here. Otherwise the roundtripper would not be able to handle the
	// response properly. In case the roundtripper replies both, a response and
	// an error, the response will be ignored.
	if IsErrStatusCode5XX(err) {
		err = nil
	}

	return res, Mask(err)
}

//------------------------------------------------------------------------------
// private

type foo struct {
	res *http.Response
	err error
}

func (t *Transport) roundTrip(req *http.Request) (*http.Response, error) {
	resChan := make(chan foo, 1)

	go func() {
		res, err := t.defaultTransport.RoundTrip(req)
		resChan <- foo{res: res, err: err}
	}()

	select {
	case <-time.After(t.Config.RequestTimeout):
		t.defaultTransport.CancelRequest(req)
		t.logger.Info("request timed out: %s", req.URL.String())

		return nil, Mask(ErrRequestTimeout)
	case x := <-resChan:
		if x.err != nil {
			return nil, Mask(x.err)
		}

		return x.res, nil
	}
}

func (t *Transport) preResHandler(ctx *context) error {
	return nil
}

func (t *Transport) postResHandler(ctx *context) error {
	return nil
}
