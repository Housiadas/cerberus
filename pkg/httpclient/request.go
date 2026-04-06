package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/Housiadas/cerberus/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

func (cln *Client) Request(
	ctx context.Context,
	method string,
	endpoint string,
	headers map[string]string,
	r io.Reader,
) (*http.Response, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint: %w", err)
	}

	ctx, span := telemetry.AddSpan(
		ctx,
		"pkg.httpclient."+path.Base(u.Path),
		attribute.String("endpoint", endpoint),
	)

	var statusCode int

	defer func() {
		cln.log.Info(ctx, "http request: completed", "status", statusCode)
		span.SetAttributes(attribute.Int("status", statusCode))
		span.End()
	}()

	req, err := http.NewRequestWithContext(ctx, method, endpoint, r)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	setHeaders(req, headers)

	resp, err := cln.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %w", err)
	}

	statusCode = resp.StatusCode

	return resp, nil
}

func setHeaders(req *http.Request, headers map[string]string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}
