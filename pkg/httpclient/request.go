package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/Housiadas/cerberus/pkg/otel"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"go.opentelemetry.io/otel/attribute"
)

func (cln *Client) Request(
	ctx context.Context,
	method string,
	endpoint string,
	headers map[string]string,
	r io.Reader,
	result any,
) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parsing endpoint: %w", err)
	}

	base := path.Base(u.Path)

	var statusCode int

	cln.log.Info(ctx, "http request: started", "method", method, "call", base, "endpoint", endpoint)

	defer func() {
		cln.log.Info(ctx, "http request: completed", "status", statusCode)
	}()

	ctx, span := otel.AddSpan(ctx, "pkg.httpclient."+base, attribute.String("endpoint", endpoint))

	defer func() {
		span.SetAttributes(attribute.Int("status", statusCode))
		span.End()
	}()

	req, err := http.NewRequestWithContext(ctx, method, endpoint, r)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	setHeaders(req, headers)

	resp, err := cln.http.Do(req)
	if err != nil {
		return fmt.Errorf("do: error: %w", err)
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode
	if statusCode == http.StatusNoContent {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}

	return parseResponse(statusCode, data, result, err)
}

func setHeaders(req *http.Request, headers map[string]string) {
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func parseResponse(statusCode int, data []byte, result any, err error) error {
	switch statusCode {
	case http.StatusNoContent:
		return nil

	case http.StatusOK:
		err := json.Unmarshal(data, result)
		if err != nil {
			return err
		}

		return nil

	case http.StatusUnauthorized:
		var errResult *errs.Error

		err = json.Unmarshal(data, &errResult)
		if err != nil {
			return err
		}

		return errResult

	default:
		return fmt.Errorf("%w: %s", ErrParseResponse, string(data))
	}
}
