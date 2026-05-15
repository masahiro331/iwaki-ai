package summarizer

import (
	"errors"
	"fmt"
)

// RetryableError marks an error as worth retrying (e.g. transient 5xx/429
// responses from an LLM API). RetryClient inspects errors via IsRetryable.
type RetryableError struct {
	Cause error
}

// Error implements error.
func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable: %v", e.Cause)
}

// Unwrap lets errors.Is / errors.As reach the underlying cause.
func (e *RetryableError) Unwrap() error {
	return e.Cause
}

// IsRetryable reports whether err (or any error it wraps) is a RetryableError.
func IsRetryable(err error) bool {
	var r *RetryableError
	return errors.As(err, &r)
}

// isRetryableStatus tells whether an HTTP status code corresponds to a
// transient failure worth retrying (rate limit or server-side error).
func isRetryableStatus(code int) bool {
	if code == 429 {
		return true
	}
	return code >= 500 && code < 600
}
