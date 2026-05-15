package summarizer

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type countingLLM struct {
	name     string
	calls    atomic.Int32
	failures int   // make Complete fail this many times before succeeding
	errKind  error // error to return while failing
	reply    string
}

func (c *countingLLM) Complete(_ context.Context, _ string) (string, error) {
	n := c.calls.Add(1)
	if int(n) <= c.failures {
		return "", c.errKind
	}
	return c.reply, nil
}

func TestRetryClient_NoErrorPassThrough(t *testing.T) {
	primary := &countingLLM{reply: "ok"}
	r := NewRetryClient(primary, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond})

	got, err := r.Complete(context.Background(), "x")
	if err != nil {
		t.Fatalf("Complete error = %v", err)
	}
	if got != "ok" {
		t.Errorf("got = %q, want ok", got)
	}
	if primary.calls.Load() != 1 {
		t.Errorf("primary calls = %d, want 1", primary.calls.Load())
	}
}

func TestRetryClient_RetriesUntilSuccess(t *testing.T) {
	primary := &countingLLM{
		failures: 2,
		errKind:  &RetryableError{Cause: errors.New("503")},
		reply:    "ok",
	}
	r := NewRetryClient(primary, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond})

	got, err := r.Complete(context.Background(), "x")
	if err != nil {
		t.Fatalf("Complete error = %v", err)
	}
	if got != "ok" {
		t.Errorf("got = %q, want ok", got)
	}
	if primary.calls.Load() != 3 {
		t.Errorf("primary calls = %d, want 3 (2 fails + 1 success)", primary.calls.Load())
	}
}

func TestRetryClient_NonRetryableFailsImmediately(t *testing.T) {
	primary := &countingLLM{
		failures: 10,
		errKind:  errors.New("400 bad request"),
	}
	r := NewRetryClient(primary, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond})

	_, err := r.Complete(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if primary.calls.Load() != 1 {
		t.Errorf("primary calls = %d, want 1 (no retries for non-retryable)", primary.calls.Load())
	}
}

func TestRetryClient_FallbackUsedAfterExhaustion(t *testing.T) {
	primary := &countingLLM{
		failures: 10,
		errKind:  &RetryableError{Cause: errors.New("503")},
	}
	fallback := &countingLLM{reply: "fallback-result"}
	r := NewRetryClient(primary, RetryConfig{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond, Fallback: fallback})

	got, err := r.Complete(context.Background(), "x")
	if err != nil {
		t.Fatalf("Complete error = %v", err)
	}
	if got != "fallback-result" {
		t.Errorf("got = %q, want fallback-result", got)
	}
	if primary.calls.Load() != 3 {
		t.Errorf("primary calls = %d, want 3", primary.calls.Load())
	}
	if fallback.calls.Load() != 1 {
		t.Errorf("fallback calls = %d, want 1", fallback.calls.Load())
	}
}

func TestRetryClient_NoFallbackReturnsLastError(t *testing.T) {
	lastErr := errors.New("final 503")
	primary := &countingLLM{
		failures: 10,
		errKind:  &RetryableError{Cause: lastErr},
	}
	r := NewRetryClient(primary, RetryConfig{MaxAttempts: 2, BaseDelay: 1 * time.Millisecond})

	_, err := r.Complete(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, lastErr) {
		t.Errorf("error should wrap last cause, got %v", err)
	}
}

func TestRetryClient_ContextCanceled(t *testing.T) {
	primary := &countingLLM{
		failures: 10,
		errKind:  &RetryableError{Cause: errors.New("503")},
	}
	r := NewRetryClient(primary, RetryConfig{MaxAttempts: 5, BaseDelay: 100 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.Complete(ctx, "x")
	if err == nil {
		t.Fatal("expected error after context cancellation")
	}
}
