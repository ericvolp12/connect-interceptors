package retry

import (
	"context"
	"errors"
	"log/slog"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/cenkalti/backoff/v4"
)

// RetryConfig contains the configuration for the retry interceptor
type RetryConfig struct {
	// A BackoffGenerator returns an instance of a backoff function for use with a single RPC call.
	// By default, it returns a fixed backoff of 1 second between attempts.
	// When the backoff function returns backoff.Stop (-1ns), the RPC call will not be retried.
	BackoffGenerator func() func() time.Duration
	// IsRetryable should return true if the error is retryable.
	// By default, it returns true for connect.CodeUnavailable and common network errors.
	IsRetryable func(error) bool
	// Set max attempts to <= 0 to retry indefinitely.
	MaxAttempts int
}

// DefaultIsRetryable returns true if the error is caused by a network or connection error
func DefaultIsRetryable(err error) bool {
	isRetryable := false
	errCode := connect.CodeOf(err)
	if errCode == connect.CodeUnavailable ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNABORTED) {
		isRetryable = true
	}
	return isRetryable
}

// DefaultRetryConfig returns a default retry configuration with a fixed backoff and retries on common network errors.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		BackoffGenerator: FixedBackoffGenerator,
		IsRetryable:      DefaultIsRetryable,
		MaxAttempts:      10,
	}
}

// FixedBackoffGenerator returns a fixed backoff duration function.
func FixedBackoffGenerator() func() time.Duration {
	return FixedBackoff
}

// FixedBackoff returns a fixed backoff duration of 1 second.
func FixedBackoff() time.Duration {
	return time.Second
}

// ExpBackoffGenerator returns an exponential backoff duration function
func ExpBackoffGenerator() func() time.Duration {
	// Allow for indefinite EBO with a maximum delay of 15 seconds.
	return ExpBackoff(backoff.WithMaxElapsedTime(0), backoff.WithMaxInterval(15*time.Second))
}

// ExpBackoff returns an exponential backoff duration based on cenkalti/backoff.
func ExpBackoff(opts ...backoff.ExponentialBackOffOpts) func() time.Duration {
	ebo := backoff.NewExponentialBackOff(opts...)
	return func() time.Duration {
		return ebo.NextBackOff()
	}
}

// NewRetryInterceptor returns a new retry interceptor.
func NewRetryInterceptor(logger *slog.Logger, conf *RetryConfig) connect.UnaryInterceptorFunc {
	if logger != nil {
		logger = logger.With("interceptor", "retry-interceptor")
	}

	if conf == nil {
		conf = DefaultRetryConfig()
	}

	nextBackoff := conf.BackoffGenerator()

	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			name := req.Spec().Procedure
			attempt := 0
			for {
				attempt++
				start := time.Now()
				res, err := next(ctx, req)
				elapsedMicros := time.Since(start).Microseconds()
				if err == nil {
					return res, nil
				}
				if !conf.IsRetryable(err) {
					return nil, err
				}
				if logger != nil {
					logger.Warn("RPC request failed", "name", name, "attempt", attempt, "elapsed_us", elapsedMicros, "err", err)
				}

				// Set max attempts to <= 0 to retry indefinitely.
				if conf.MaxAttempts > 0 && attempt >= conf.MaxAttempts {
					if logger != nil {
						logger.Warn("RPC request failed after max attempts", "name", name, "max_attempts", conf.MaxAttempts, "elapsed_us", elapsedMicros, "err", err)
					}
					return nil, err
				}

				sleepTime := nextBackoff()
				if sleepTime == backoff.Stop {
					if logger != nil {
						logger.Warn("RPC request failed after max backoff", "name", name, "elapsed_us", elapsedMicros, "err", err)
					}
					return nil, err
				}
				time.Sleep(sleepTime)
			}
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
