package summarizer

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RetryConfig tunes the retry/fallback behaviour of RetryClient.
type RetryConfig struct {
	MaxAttempts int           // total attempts against the primary client (>=1)
	BaseDelay   time.Duration // first backoff delay; doubles each retry
	Fallback    LLMClient     // optional: tried once if primary exhausts retries
}

// RetryClient wraps an LLMClient with exponential-backoff retries
// against retryable errors, plus an optional fallback client tried
// once after the primary's retries are exhausted.
type RetryClient struct {
	primary LLMClient
	cfg     RetryConfig
}

// NewRetryClient builds a RetryClient.
func NewRetryClient(primary LLMClient, cfg RetryConfig) *RetryClient {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 1 * time.Second
	}
	return &RetryClient{primary: primary, cfg: cfg}
}

// Complete tries the primary with retries on retryable errors;
// if exhausted and a fallback is configured, it tries the fallback once.
func (r *RetryClient) Complete(ctx context.Context, prompt string) (string, error) {
	var lastErr error
	delay := r.cfg.BaseDelay

	for attempt := 1; attempt <= r.cfg.MaxAttempts; attempt++ {
		out, err := r.primary.Complete(ctx, prompt)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !IsRetryable(err) {
			return "", err
		}
		if attempt == r.cfg.MaxAttempts {
			break
		}
		log.Printf("llm retry attempt %d/%d after %v: %v", attempt, r.cfg.MaxAttempts, delay, err)
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("retry interrupted: %w", ctx.Err())
		case <-time.After(delay):
		}
		delay *= 2
	}

	if r.cfg.Fallback != nil {
		log.Printf("llm primary exhausted after %d attempts, trying fallback: %v", r.cfg.MaxAttempts, lastErr)
		out, err := r.cfg.Fallback.Complete(ctx, prompt)
		if err == nil {
			return out, nil
		}
		return "", fmt.Errorf("fallback also failed (primary err: %v): %w", lastErr, err)
	}
	return "", fmt.Errorf("retries exhausted: %w", lastErr)
}
