package httpro

import (
	"net/http"
	"time"
)

type Config struct {
	ReconnectDelay time.Duration
	ConnectTimeout time.Duration
	RequestTimeout time.Duration

	RequestRetry uint
}

type Client struct {
	Config     Config
	HTTPClient *http.Client
}

func NewHTTPClient(c Config) *http.Client {
	httpClient := &http.Client{
		Transport: NewTransport(TransportConfig{
			ReconnectDelay: c.ReconnectDelay,
			ConnectTimeout: c.ConnectTimeout,
			RequestTimeout: c.RequestTimeout,

			RequestRetry: c.RequestRetry,
		}),
	}

	return httpClient
}

//------------------------------------------------------------------------------
// private
