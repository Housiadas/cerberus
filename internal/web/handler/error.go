package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Housiadas/cerberus/internal/sdk/errs"
)

// RequestErrorHandler handles errors during request parsing (e.g. invalid JSON).
func requestErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Set(ContentTypeKey, ContentTypeJSON)
	w.WriteHeader(http.StatusBadRequest)

	//nolint:errchkjson // best-effort error response encoding
	_ = json.NewEncoder(w).Encode(errs.New(errs.InvalidArgument, errs.CodeRequestInvalid, err))
}

// ResponseErrorHandler converts handler errors into the correct HTTP error response.
func responseErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	var appErr *errs.Error

	ok := errors.As(err, &appErr)
	if !ok {
		appErr = errs.Errorf(errs.Internal, errs.CodeInternal, "Internal Server Error")
	}

	statusCode := appErr.HTTPStatus()

	w.Header().Set(ContentTypeKey, ContentTypeJSON)
	w.WriteHeader(statusCode)

	//nolint:errchkjson // best-effort error response encoding
	_ = json.NewEncoder(w).Encode(appErr)
}
