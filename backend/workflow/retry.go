package workflow

import (
	"context"
	"math"
	"time"
)

// RetryPolicy determines whether another attempt should be made and how long to wait.
type RetryPolicy interface {
	ShouldRetry(attempt int) bool
	Delay(attempt int) time.Duration
}

// FixedRetryPolicy waits a constant duration between every attempt.
type FixedRetryPolicy struct {
	maxAttempts int
	baseDelay   time.Duration
}

// NewFixedRetry constructs a FixedRetryPolicy from a ParsedRetry config.
func NewFixedRetry(p ParsedRetry) RetryPolicy {
	return &FixedRetryPolicy{maxAttempts: p.MaxAttempts, baseDelay: p.BaseDelay}
}

func (r *FixedRetryPolicy) ShouldRetry(attempt int) bool { return attempt < r.maxAttempts }
func (r *FixedRetryPolicy) Delay(_ int) time.Duration     { return r.baseDelay }

// ExponentialBackoffPolicy applies baseDelay * 2^(attempt-1), capped at maxDelay.
type ExponentialBackoffPolicy struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

// NewExponentialBackoff constructs an ExponentialBackoffPolicy from a ParsedRetry config.
func NewExponentialBackoff(p ParsedRetry) RetryPolicy {
	maxDelay := p.MaxDelay
	if maxDelay == 0 {
		maxDelay = 60 * time.Second
	}
	return &ExponentialBackoffPolicy{
		maxAttempts: p.MaxAttempts,
		baseDelay:   p.BaseDelay,
		maxDelay:    maxDelay,
	}
}

func (r *ExponentialBackoffPolicy) ShouldRetry(attempt int) bool { return attempt < r.maxAttempts }

func (r *ExponentialBackoffPolicy) Delay(attempt int) time.Duration {
	if attempt <= 1 {
		return r.baseDelay
	}
	delay := time.Duration(float64(r.baseDelay) * math.Pow(2, float64(attempt-1)))
	if delay > r.maxDelay {
		return r.maxDelay
	}
	return delay
}

// BuildRetryPolicy selects and constructs the correct RetryPolicy for a step.
func BuildRetryPolicy(p ParsedRetry) RetryPolicy {
	switch p.Strategy {
	case "exponential_backoff":
		return NewExponentialBackoff(p)
	default:
		return NewFixedRetry(p)
	}
}

// sleep waits for duration d while honouring context cancellation.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
