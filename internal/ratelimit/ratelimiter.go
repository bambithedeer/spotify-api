package ratelimit

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bambithedeer/spotify-api/internal/errors"
)

// RateLimiter manages rate limiting for Spotify API requests
type RateLimiter struct {
	mu                sync.RWMutex
	tokens            int           // Available tokens
	maxTokens         int           // Maximum tokens
	refillRate        time.Duration // Rate at which tokens are refilled
	lastRefill        time.Time     // Last time tokens were refilled
	retryAfter        time.Time     // Time until rate limit resets
	maxRetries        int           // Maximum number of retries
	baseRetryDelay    time.Duration // Base delay for exponential backoff
	maxRetryDelay     time.Duration // Maximum retry delay
}

// NewRateLimiter creates a new rate limiter with Spotify API defaults
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		tokens:         100,           // Start with full bucket
		maxTokens:      100,           // Spotify allows ~100 requests per minute in bursts
		refillRate:     600 * time.Millisecond, // Refill 1 token every 600ms (100 per minute)
		lastRefill:     time.Now(),
		maxRetries:     3,
		baseRetryDelay: 1 * time.Second,
		maxRetryDelay:  30 * time.Second,
	}
}

// NewCustomRateLimiter creates a rate limiter with custom settings
func NewCustomRateLimiter(maxTokens int, refillRate time.Duration, maxRetries int) *RateLimiter {
	return &RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefill:     time.Now(),
		maxRetries:     maxRetries,
		baseRetryDelay: 1 * time.Second,
		maxRetryDelay:  30 * time.Second,
	}
}

// Wait blocks until a token is available or the context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()

		// Check if we're in a rate limit cooldown period
		if time.Now().Before(rl.retryAfter) {
			waitTime := time.Until(rl.retryAfter)
			rl.mu.Unlock()

			select {
			case <-time.After(waitTime):
				continue // Try again after cooldown
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Refill tokens based on elapsed time
		rl.refillTokens()

		// If we have tokens available, use one
		if rl.tokens > 0 {
			rl.tokens--
			rl.mu.Unlock()
			return nil
		}

		// Need to wait for next token
		waitTime := rl.refillRate
		rl.mu.Unlock()

		select {
		case <-time.After(waitTime):
			continue // Try again after waiting
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// TryWait attempts to acquire a token without blocking
func (rl *RateLimiter) TryWait() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if we're in a rate limit cooldown period
	if time.Now().Before(rl.retryAfter) {
		return false
	}

	// Refill tokens based on elapsed time
	rl.refillTokens()

	// If we have tokens available, use one
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// HandleRateLimitResponse processes a 429 response and updates rate limiter state
func (rl *RateLimiter) HandleRateLimitResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusTooManyRequests {
		return nil
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Parse Retry-After header
	retryAfterStr := resp.Header.Get("Retry-After")
	if retryAfterStr != "" {
		if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
			rl.retryAfter = time.Now().Add(time.Duration(seconds) * time.Second)
		}
	} else {
		// Default backoff if no Retry-After header
		rl.retryAfter = time.Now().Add(60 * time.Second)
	}

	// Reset token bucket
	rl.tokens = 0

	return errors.NewAPIError(fmt.Sprintf("rate limited until %v", rl.retryAfter))
}

// refillTokens adds tokens based on elapsed time (must be called with lock held)
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	if elapsed >= rl.refillRate {
		tokensToAdd := int(elapsed / rl.refillRate)
		rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}
}

// GetStatus returns current rate limiter status
func (rl *RateLimiter) GetStatus() (availableTokens int, maxTokens int, retryAfter time.Time) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Refill tokens to get current state
	rl.mu.RUnlock()
	rl.mu.Lock()
	rl.refillTokens()
	rl.mu.Unlock()
	rl.mu.RLock()

	return rl.tokens, rl.maxTokens, rl.retryAfter
}

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries     int
	BaseDelay      time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	RetryableErrors map[int]bool // HTTP status codes that should trigger retries
}

// DefaultRetryConfig returns default retry configuration for Spotify API
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		BaseDelay:     1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: map[int]bool{
			http.StatusTooManyRequests:     true, // 429
			http.StatusInternalServerError: true, // 500
			http.StatusBadGateway:          true, // 502
			http.StatusServiceUnavailable:  true, // 503
			http.StatusGatewayTimeout:      true, // 504
		},
	}
}

// ShouldRetry determines if a request should be retried based on the response
func (rc *RetryConfig) ShouldRetry(resp *http.Response, attempt int) bool {
	if attempt >= rc.MaxRetries {
		return false
	}

	if resp == nil {
		return true // Network error, should retry
	}

	return rc.RetryableErrors[resp.StatusCode]
}

// GetRetryDelay calculates the delay before next retry attempt
func (rc *RetryConfig) GetRetryDelay(attempt int, resp *http.Response) time.Duration {
	// Check for Retry-After header first
	if resp != nil && resp.Header.Get("Retry-After") != "" {
		if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil {
			delay := time.Duration(seconds) * time.Second
			if delay <= rc.MaxDelay {
				return delay
			}
		}
	}

	// Calculate exponential backoff
	delay := time.Duration(float64(rc.BaseDelay) * math.Pow(rc.BackoffFactor, float64(attempt)))
	if delay > rc.MaxDelay {
		delay = rc.MaxDelay
	}

	return delay
}

// Helper function for min (Go 1.21+ has this built-in)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}