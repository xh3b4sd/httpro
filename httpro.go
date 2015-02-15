package httpro

import (
	"net/http"
	"time"

	"github.com/zyndiecate/httpro/breaker"
	"github.com/zyndiecate/httpro/transport"
)

// Config to configure the HTTP client.
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

// NewHTTPClient creates a new *http.Client.
//
//   c := httpro.NewHTTPClient(httpro.Config{
//     ReconnectDelay: 500 * time.Millisecond,
//     ConnectTimeout: 500 * time.Millisecond,
//     RequestTimeout: 500 * time.Millisecond,
//
//     RequestRetry: 5,
//   })
//
//   res, err = c.Get("https://google.com/")
//
func NewHTTPClient(c Config) *http.Client {
	httpClient := &http.Client{
		Transport: transport.NewTransport(transport.Config{
			ReconnectDelay: c.ReconnectDelay,
			ConnectTimeout: c.ConnectTimeout,
			RequestTimeout: c.RequestTimeout,

			RequestRetry: c.RequestRetry,

			BreakerConfig: c.BreakerConfig,
		}),
	}

	return httpClient
}
