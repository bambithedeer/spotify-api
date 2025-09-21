package ratelimit

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter()

	if rl.maxTokens != 100 {
		t.Errorf("Expected maxTokens 100, got %d", rl.maxTokens)
	}

	if rl.tokens != 100 {
		t.Errorf("Expected initial tokens 100, got %d", rl.tokens)
	}

	if rl.refillRate != 600*time.Millisecond {
		t.Errorf("Expected refillRate 600ms, got %v", rl.refillRate)
	}
}

func TestNewCustomRateLimiter(t *testing.T) {
	maxTokens := 50
	refillRate := 1 * time.Second
	maxRetries := 5

	rl := NewCustomRateLimiter(maxTokens, refillRate, maxRetries)

	if rl.maxTokens != maxTokens {
		t.Errorf("Expected maxTokens %d, got %d", maxTokens, rl.maxTokens)
	}

	if rl.refillRate != refillRate {
		t.Errorf("Expected refillRate %v, got %v", refillRate, rl.refillRate)
	}

	if rl.maxRetries != maxRetries {
		t.Errorf("Expected maxRetries %d, got %d", maxRetries, rl.maxRetries)
	}
}

func TestTryWait(t *testing.T) {
	rl := NewRateLimiter()

	// Should succeed when tokens are available
	if !rl.TryWait() {
		t.Error("Expected TryWait to succeed when tokens are available")
	}

	// Verify token was consumed
	available, _, _ := rl.GetStatus()
	if available != 99 {
		t.Errorf("Expected 99 tokens after one use, got %d", available)
	}
}

func TestTryWaitExhausted(t *testing.T) {
	// Create rate limiter with only 1 token
	rl := NewCustomRateLimiter(1, 1*time.Hour, 3) // Very slow refill

	// Use the only token
	if !rl.TryWait() {
		t.Error("Expected first TryWait to succeed")
	}

	// Should fail when no tokens available
	if rl.TryWait() {
		t.Error("Expected TryWait to fail when no tokens available")
	}
}

func TestWaitWithContext(t *testing.T) {
	rl := NewCustomRateLimiter(1, 100*time.Millisecond, 3)
	ctx := context.Background()

	// First wait should succeed immediately
	err := rl.Wait(ctx)
	if err != nil {
		t.Errorf("Expected first Wait to succeed, got error: %v", err)
	}

	// Second wait should succeed after refill
	start := time.Now()
	err = rl.Wait(ctx)
	if err != nil {
		t.Errorf("Expected second Wait to succeed, got error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 90*time.Millisecond { // Allow some tolerance
		t.Errorf("Expected wait time >= 90ms, got %v", elapsed)
	}
}

func TestWaitWithCancelledContext(t *testing.T) {
	rl := NewCustomRateLimiter(1, 1*time.Hour, 3) // Very slow refill
	ctx, cancel := context.WithCancel(context.Background())

	// Use the only token
	_ = rl.Wait(ctx)

	// Cancel context immediately
	cancel()

	// Wait should return context error
	err := rl.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestHandleRateLimitResponse(t *testing.T) {
	rl := NewRateLimiter()

	// Create a mock 429 response with Retry-After header
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     make(http.Header),
	}
	resp.Header.Set("Retry-After", "60")

	err := rl.HandleRateLimitResponse(resp)
	if err == nil {
		t.Error("Expected error from HandleRateLimitResponse")
	}

	// Check that tokens were reset and retry after was set
	available, _, retryAfter := rl.GetStatus()
	if available != 0 {
		t.Errorf("Expected 0 tokens after rate limit response, got %d", available)
	}

	if time.Until(retryAfter) < 59*time.Second {
		t.Errorf("Expected retry after to be set ~60 seconds in future")
	}
}

func TestHandleRateLimitResponseNoHeader(t *testing.T) {
	rl := NewRateLimiter()

	// Create a mock 429 response without Retry-After header
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     make(http.Header),
	}

	err := rl.HandleRateLimitResponse(resp)
	if err == nil {
		t.Error("Expected error from HandleRateLimitResponse")
	}

	// Should use default 60 second backoff
	_, _, retryAfter := rl.GetStatus()
	if time.Until(retryAfter) < 59*time.Second {
		t.Errorf("Expected default 60 second retry after period")
	}
}

func TestHandleNonRateLimitResponse(t *testing.T) {
	rl := NewRateLimiter()

	// Create a mock 200 response
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
	}

	err := rl.HandleRateLimitResponse(resp)
	if err != nil {
		t.Errorf("Expected no error for non-429 response, got %v", err)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}

	if config.BaseDelay != 1*time.Second {
		t.Errorf("Expected BaseDelay 1s, got %v", config.BaseDelay)
	}

	// Test retryable status codes
	retryableCodes := []int{429, 500, 502, 503, 504}
	for _, code := range retryableCodes {
		if !config.RetryableErrors[code] {
			t.Errorf("Expected status code %d to be retryable", code)
		}
	}

	// Test non-retryable status codes
	nonRetryableCodes := []int{400, 401, 403, 404}
	for _, code := range nonRetryableCodes {
		if config.RetryableErrors[code] {
			t.Errorf("Expected status code %d to not be retryable", code)
		}
	}
}

func TestShouldRetry(t *testing.T) {
	config := DefaultRetryConfig()

	// Test max retries exceeded
	resp := &http.Response{StatusCode: 429}
	if config.ShouldRetry(resp, 3) {
		t.Error("Should not retry when max retries exceeded")
	}

	// Test retryable status code
	if !config.ShouldRetry(resp, 1) {
		t.Error("Should retry for 429 status code")
	}

	// Test non-retryable status code
	resp.StatusCode = 404
	if config.ShouldRetry(resp, 1) {
		t.Error("Should not retry for 404 status code")
	}

	// Test nil response (network error)
	if !config.ShouldRetry(nil, 1) {
		t.Error("Should retry for network errors (nil response)")
	}
}

func TestGetRetryDelay(t *testing.T) {
	config := DefaultRetryConfig()

	// Test with Retry-After header
	resp := &http.Response{
		StatusCode: 429,
		Header:     make(http.Header),
	}
	resp.Header.Set("Retry-After", "5")

	delay := config.GetRetryDelay(1, resp)
	if delay != 5*time.Second {
		t.Errorf("Expected 5s delay from Retry-After header, got %v", delay)
	}

	// Test exponential backoff without header
	resp.Header.Del("Retry-After")

	delay1 := config.GetRetryDelay(0, resp)
	delay2 := config.GetRetryDelay(1, resp)
	delay3 := config.GetRetryDelay(2, resp)

	if delay1 != 1*time.Second {
		t.Errorf("Expected 1s delay for attempt 0, got %v", delay1)
	}

	if delay2 != 2*time.Second {
		t.Errorf("Expected 2s delay for attempt 1, got %v", delay2)
	}

	if delay3 != 4*time.Second {
		t.Errorf("Expected 4s delay for attempt 2, got %v", delay3)
	}

	// Test max delay cap
	longDelay := config.GetRetryDelay(10, resp)
	if longDelay != config.MaxDelay {
		t.Errorf("Expected max delay %v for large attempt, got %v", config.MaxDelay, longDelay)
	}
}

func TestRefillTokens(t *testing.T) {
	// Create rate limiter with fast refill for testing
	rl := NewCustomRateLimiter(10, 10*time.Millisecond, 3)

	// Use all tokens
	for i := 0; i < 10; i++ {
		if !rl.TryWait() {
			t.Fatalf("Expected to consume token %d", i)
		}
	}

	// Verify no tokens left
	if rl.TryWait() {
		t.Error("Expected no tokens to be available")
	}

	// Wait for refill
	time.Sleep(50 * time.Millisecond) // Should refill ~5 tokens

	available, _, _ := rl.GetStatus()
	if available < 4 { // Allow some tolerance
		t.Errorf("Expected at least 4 tokens after refill, got %d", available)
	}
}