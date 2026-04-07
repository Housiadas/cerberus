package errs

import "net/http"

const (
	// None indicates the operation was successful.
	None StatusCode = iota
	// NoContent indicates the operation was successful with no content.
	NoContent
	// Canceled indicates the operation was canceled (typically by the caller).
	Canceled
	// Unknown error.
	Unknown
	// InvalidArgument indicates a client specified an invalid argument.
	InvalidArgument
	// DeadlineExceeded means the operation expired before completion.
	DeadlineExceeded
	// NotFound means some requested entity was not found.
	NotFound
	// AlreadyExists means an attempt to create an entity failed because one already exists.
	AlreadyExists
	// PermissionDenied indicates the caller does not have permission to execute the operation.
	PermissionDenied
	// ResourceExhausted indicates some resource has been exhausted.
	ResourceExhausted
	// FailedPrecondition indicates the operation was rejected because the system is not in
	// a state required for the operation's execution.
	FailedPrecondition
	// Aborted indicates the operation was aborted, typically due to a concurrency issue.
	Aborted
	// OutOfRange means the operation was attempted past the valid range.
	OutOfRange
	// Unimplemented indicates the operation is not implemented or not supported/enabled.
	Unimplemented
	// Internal errors. Means some invariants expected by the underlying system have been broken.
	Internal
	// Unavailable indicates the usecase is currently unavailable.
	Unavailable
	// DataLoss indicates unrecoverable data loss or corruption.
	DataLoss
	// Unauthenticated indicates the request does not have valid authentication credentials.
	Unauthenticated
	// TooManyRequests indicates the client has exceeded their rate limit and/or quota.
	TooManyRequests
	// InternalOnlyLog like Internal, but the error message is not sent to the client.
	InternalOnlyLog
)

type statusInfo struct {
	name string
	http int
}

var statusByName = func() map[string]StatusCode {
	m := make(map[string]StatusCode, len(statusTable))
	for k, v := range statusTable {
		m[v.name] = k
	}

	return m
}()

var statusTable = map[StatusCode]statusInfo{
	None:               {"ok", http.StatusOK},
	NoContent:          {"ok_no_content", http.StatusNoContent},
	Canceled:           {"canceled", http.StatusGatewayTimeout},
	Unknown:            {"unknown", http.StatusInternalServerError},
	InvalidArgument:    {"invalid_argument", http.StatusBadRequest},
	DeadlineExceeded:   {"deadline_exceeded", http.StatusGatewayTimeout},
	NotFound:           {"not_found", http.StatusNotFound},
	AlreadyExists:      {"already_exists", http.StatusConflict},
	PermissionDenied:   {"permission_denied", http.StatusForbidden},
	ResourceExhausted:  {"resource_exhausted", http.StatusTooManyRequests},
	FailedPrecondition: {"failed_precondition", http.StatusBadRequest},
	Aborted:            {"aborted", http.StatusConflict},
	OutOfRange:         {"out_of_range", http.StatusBadRequest},
	Unimplemented:      {"unimplemented", http.StatusNotImplemented},
	Internal:           {"internal", http.StatusInternalServerError},
	Unavailable:        {"unavailable", http.StatusServiceUnavailable},
	DataLoss:           {"data_loss", http.StatusInternalServerError},
	Unauthenticated:    {"unauthenticated", http.StatusUnauthorized},
	TooManyRequests:    {"too_many_requests", http.StatusTooManyRequests},
	InternalOnlyLog:    {"internal_only_log", http.StatusInternalServerError},
}
