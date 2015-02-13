package httpro

import (
	"net/http"
	"time"
)

type Config struct {
	RequestTimeout time.Duration
	ConnectTimeout time.Duration
	RequestRetry   uint
}

type Client struct {
	Config     Config
	HTTPClient *http.Client
}

func NewHTTPClient(c Config) *http.Client {
	httpClient := &http.Client{
		Transport: NewTransport(TransportConfig{
			RequestTimeout: c.RequestTimeout,
			ConnectTimeout: c.ConnectTimeout,
			RequestRetry:   c.RequestRetry,
		}),
	}

	return httpClient
}

//------------------------------------------------------------------------------
// private
