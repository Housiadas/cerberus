package elasticsearch

import "errors"

var (
	ErrConnectionFailed = errors.New("elasticsearch connection failed")
	ErrIndexNotFound    = errors.New("elasticsearch index not found")
	ErrDocumentNotFound = errors.New("elasticsearch document not found")
	ErrBulkIndexFailed  = errors.New("elasticsearch bulk index failed")
)
