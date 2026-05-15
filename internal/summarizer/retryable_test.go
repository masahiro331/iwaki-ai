package summarizer

import (
	"errors"
	"fmt"
	"testing"
)

func TestRetryable_IsRetryable(t *testing.T) {
	plain := errors.New("plain")
	if IsRetryable(plain) {
		t.Errorf("plain error should not be retryable")
	}

	wrapped := fmt.Errorf("outer: %w", &RetryableError{Cause: plain})
	if !IsRetryable(wrapped) {
		t.Errorf("error wrapping RetryableError should be retryable")
	}
}

func TestRetryableError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	r := &RetryableError{Cause: inner}
	if !errors.Is(r, inner) {
		t.Errorf("errors.Is should find inner cause")
	}
}

func TestRetryableError_Error(t *testing.T) {
	inner := errors.New("503 overloaded")
	r := &RetryableError{Cause: inner}
	if r.Error() == "" {
		t.Errorf("Error() should not be empty")
	}
}
