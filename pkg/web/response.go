package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/otel"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"go.opentelemetry.io/otel/attribute"
)

type Respond struct {
	log logger.Logger
}

func NewRespond(log logger.Logger) *Respond {
	return &Respond{
		log: log,
	}
}

func (respond *Respond) Respond(handlerFunc HandlerFunc) http.HandlerFunc {
	// This is the decorator/middleware pattern in golang
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Executes the handlerFunc for the specific route
		resp := handlerFunc(ctx, w, r)

		// Get status code
		statusCode := respond.statusCode(resp)

		// Record errors with status code 500 and above
		err := isError(resp)
		if err != nil {
			resp = respond.errorRecorder(ctx, statusCode, err)
		}

		// Send an encoded response back to a client
		if err := respond.encode(ctx, w, statusCode, resp); err != nil {
			respond.log.Error(ctx, "web-respond", "ERROR", err)
		}
	}
}

func (respond *Respond) encode(
	ctx context.Context,
	w http.ResponseWriter,
	statusCode int,
	dataModel Encoder,
) error {
	// If the context has been canceled, it means the client is no longer waiting for a encode.
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send encode")
		}
	}

	_, span := otel.AddSpan(ctx, "web.encode", attribute.Int("status", statusCode))
	defer span.End()

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)

		return nil
	}

	data, contentType, err := dataModel.Encode()
	if err != nil {
		return fmt.Errorf("respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}

	return nil
}

func (respond *Respond) errorRecorder(ctx context.Context, statusCode int, err error) Encoder {
	var appErr *errs.Error

	ok := errors.As(err, &appErr)
	if !ok {
		appErr = errs.Errorf(errs.Internal, "Internal Server Error")
	}

	// If not, the critical error does not record it
	if statusCode < http.StatusInternalServerError {
		return appErr
	}

	_, span := otel.AddSpan(ctx, "web.encode.error")

	span.RecordError(err)
	defer span.End()

	respond.log.Error(ctx, "error during request",
		"err", err,
		"source_err_file", path.Base(appErr.FileName),
		"source_err_func", path.Base(appErr.FuncName),
	)

	if appErr.Code == errs.InternalOnlyLog {
		appErr = errs.Errorf(errs.Internal, "Internal Server Error")
	}

	// Send the error back so it can be used as the encode.
	return appErr
}

// isError checks if the Encoder has an error inside it.
func isError(e Encoder) error {
	err, isError := e.(error)
	if isError {
		return err
	}

	return nil
}

func (respond *Respond) statusCode(dataModel Encoder) int {
	statusCode := http.StatusOK

	switch v := dataModel.(type) {
	case httpStatus:
		statusCode = v.HTTPStatus()
	case error:
		statusCode = http.StatusInternalServerError
	default:
		if dataModel == nil {
			statusCode = http.StatusNoContent
		}
	}

	return statusCode
}
