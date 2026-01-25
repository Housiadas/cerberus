package worker

import "errors"

var (
	ErrMaxRunningJobsInvalid = errors.New("max running jobs must be greater than 0")
	ErrShuttingDown          = errors.New("shutting down")
	ErrWorkNotRunning        = errors.New("work is not running")
)
