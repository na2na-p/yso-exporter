// Package errors provides custom error types and error handling utilities
// for the YSO-Exporter application.
package errors

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"
)

// ErrorCode represents the type of error
type ErrorCode int

const (
	// CodeUnknown represents an unknown error
	CodeUnknown ErrorCode = iota
	// CodeNotFound represents a resource not found error
	CodeNotFound
	// CodePermission represents a permission denied error
	CodePermission
	// CodeConfig represents a configuration error
	CodeConfig
	// CodeIO represents an I/O error
	CodeIO
	// CodeInternal represents an internal error
	CodeInternal
	// CodeProcessNotFound represents a process not found error
	CodeProcessNotFound
	// CodeMemoryRead represents a memory read error
	CodeMemoryRead
)

// String returns the string representation of the error code
func (c ErrorCode) String() string {
	switch c {
	case CodeNotFound:
		return "NOT_FOUND"
	case CodePermission:
		return "PERMISSION"
	case CodeConfig:
		return "CONFIG"
	case CodeIO:
		return "IO"
	case CodeInternal:
		return "INTERNAL"
	case CodeProcessNotFound:
		return "PROCESS_NOT_FOUND"
	case CodeMemoryRead:
		return "MEMORY_READ"
	default:
		return "UNKNOWN"
	}
}

// YSOError represents a custom error with additional context
type YSOError struct {
	code       ErrorCode
	message    string
	wrapped    error
	stackTrace []uintptr
	retryable  bool
}

// Error implements the error interface
func (e *YSOError) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("%s: %s: %v", e.code.String(), e.message, e.wrapped)
	}
	return fmt.Sprintf("%s: %s", e.code.String(), e.message)
}

// Unwrap implements the errors.Unwrap interface
func (e *YSOError) Unwrap() error {
	return e.wrapped
}

// Code returns the error code
func (e *YSOError) Code() ErrorCode {
	return e.code
}

// IsRetryable returns whether the error is retryable
func (e *YSOError) IsRetryable() bool {
	return e.retryable
}

// StackTrace returns the stack trace
func (e *YSOError) StackTrace() []uintptr {
	return e.stackTrace
}

// NewError creates a new YSOError
func NewError(code ErrorCode, message string) *YSOError {
	return &YSOError{
		code:       code,
		message:    message,
		stackTrace: captureStackTrace(),
		retryable:  isRetryableCode(code),
	}
}

// WrapError wraps an existing error with additional context
func WrapError(code ErrorCode, message string, err error) *YSOError {
	return &YSOError{
		code:       code,
		message:    message,
		wrapped:    err,
		stackTrace: captureStackTrace(),
		retryable:  isRetryableCode(code),
	}
}

// NewProcessNotFoundError creates a process not found error
func NewProcessNotFoundError(processName string) *YSOError {
	return NewError(CodeProcessNotFound, fmt.Sprintf("process not found: %s", processName))
}

// NewMemoryReadError creates a memory read error
func NewMemoryReadError(address uintptr, err error) *YSOError {
	return WrapError(CodeMemoryRead, fmt.Sprintf("failed to read memory at address 0x%x", address), err)
}

// NewConfigLoadError creates a config load error
func NewConfigLoadError(path string, err error) *YSOError {
	return WrapError(CodeConfig, fmt.Sprintf("failed to load config from %s", path), err)
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
	RetryableBy func(error) bool
}

// DefaultRetryPolicy returns the default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:  3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
		RetryableBy: IsRetryable,
	}
}

// RetryFunc is a function that can be retried
type RetryFunc func(context.Context) error

// Retry executes a function with retry logic
func Retry(ctx context.Context, policy *RetryPolicy, fn RetryFunc) error {
	// 最初にコンテキストをチェック
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	var lastErr error
	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := calculateDelay(policy, attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		if err := fn(ctx); err != nil {
			lastErr = err
			if !policy.RetryableBy(err) {
				return err
			}
			continue
		}
		return nil
	}
	return lastErr
}

// RecoverFunc handles panic recovery
type RecoverFunc func(interface{}) error

// WithRecover wraps a function with panic recovery
func WithRecover(fn func() error, recoverfn RecoverFunc) error {
	defer func() {
		if r := recover(); r != nil {
			if recoverfn != nil {
				recoverfn(r)
			}
		}
	}()
	return fn()
}

// IsRetryable returns whether an error is retryable
func IsRetryable(err error) bool {
	var ysoErr *YSOError
	if errors.As(err, &ysoErr) {
		return ysoErr.IsRetryable()
	}
	return false
}

// As is a convenience function for errors.As
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is is a convenience function for errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// captureStackTrace captures the current stack trace
func captureStackTrace() []uintptr {
	const maxStackDepth = 32
	pcs := make([]uintptr, maxStackDepth)
	n := runtime.Callers(3, pcs) // Skip captureStackTrace, NewError/WrapError, and the calling function
	return pcs[:n]
}

// isRetryableCode determines if an error code is retryable by default
func isRetryableCode(code ErrorCode) bool {
	switch code {
	case CodeMemoryRead, CodeProcessNotFound:
		return true
	default:
		return false
	}
}

// calculateDelay calculates the delay for the next retry attempt
func calculateDelay(policy *RetryPolicy, attempt int) time.Duration {
	delay := time.Duration(float64(policy.BaseDelay) * pow(policy.Multiplier, float64(attempt-1)))
	if delay > policy.MaxDelay {
		delay = policy.MaxDelay
	}
	
	if policy.Jitter {
		// Add up to 10% jitter
		jitter := time.Duration(float64(delay) * 0.1 * (2.0*rand() - 1.0))
		delay += jitter
	}
	
	return delay
}

// pow is a simple power function for integers
func pow(base float64, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// rand returns a random number between 0 and 1
// This is a simple implementation to avoid external dependencies
func rand() float64 {
	return float64(time.Now().UnixNano()%1000) / 1000.0
}