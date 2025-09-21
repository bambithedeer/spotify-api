package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{"config error", NewConfigError("test"), IsConfigError, true},
		{"auth error", NewAuthError("test"), IsAuthError, true},
		{"api error", NewAPIError("test"), IsAPIError, true},
		{"network error", NewNetworkError("test"), IsNetworkError, true},
		{"validation error", NewValidationError("test"), IsValidationError, true},
		{"file error", NewFileError("test"), IsFileError, true},
		{"wrong type", NewConfigError("test"), IsAuthError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFn(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWrappedErrors(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := WrapConfigError(baseErr, "wrapped")

	if !IsConfigError(wrappedErr) {
		t.Error("Expected wrapped error to be config error")
	}

	// Check that the original error is contained in the error message
	if !strings.Contains(wrappedErr.Error(), baseErr.Error()) {
		t.Error("Expected wrapped error to contain base error message")
	}
}

func TestErrorMessages(t *testing.T) {
	err := NewConfigError("test message")
	expected := "configuration error: test message"

	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}