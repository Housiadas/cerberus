// Package httpclient provides support for external http requests
package httpclient

import (
	"net"
	"net/http"
	"time"

	"github.com/Housiadas/cerberus/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Config struct {
	Log     *logger.Service
	Timeout time.Duration
}

// Client represents an http client.
type Client struct {
	log  *logger.Service
	http *http.Client
}

// New constructs an http client.
func New(cfg Config) *Client {
	cln := Client{
		log: cfg.Log,
		http: &http.Client{
			Transport: otelhttp.NewTransport(&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 2 * time.Second, // specifies the amount of time to wait for a grpc's first response
			}),
			Timeout: cfg.Timeout, // specifies a time limit for requests made by this Client.
		},
	}

	return &cln
}
