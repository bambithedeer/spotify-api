package errors

import (
	"errors"
	"fmt"
)

// Error types as sentinel errors
var (
	ErrConfig     = errors.New("configuration error")
	ErrAuth       = errors.New("authentication error")
	ErrAPI        = errors.New("API error")
	ErrNetwork    = errors.New("network error")
	ErrValidation = errors.New("validation error")
	ErrFile       = errors.New("file error")
)

// Wrap wraps an error with additional context and type
func Wrap(err error, errorType error, message string) error {
	return fmt.Errorf("%w: %s: %v", errorType, message, err)
}

// New creates a new error with type and message
func New(errorType error, message string) error {
	return fmt.Errorf("%w: %s", errorType, message)
}

// Convenience functions for common error types

func NewConfigError(message string) error {
	return New(ErrConfig, message)
}

func NewAuthError(message string) error {
	return New(ErrAuth, message)
}

func NewAPIError(message string) error {
	return New(ErrAPI, message)
}

func NewNetworkError(message string) error {
	return New(ErrNetwork, message)
}

func NewValidationError(message string) error {
	return New(ErrValidation, message)
}

func NewFileError(message string) error {
	return New(ErrFile, message)
}

func WrapConfigError(err error, message string) error {
	return Wrap(err, ErrConfig, message)
}

func WrapAuthError(err error, message string) error {
	return Wrap(err, ErrAuth, message)
}

func WrapAPIError(err error, message string) error {
	return Wrap(err, ErrAPI, message)
}

func WrapNetworkError(err error, message string) error {
	return Wrap(err, ErrNetwork, message)
}

func WrapValidationError(err error, message string) error {
	return Wrap(err, ErrValidation, message)
}

func WrapFileError(err error, message string) error {
	return Wrap(err, ErrFile, message)
}

// IsType checks if an error is of a specific type using errors.Is
func IsConfigError(err error) bool {
	return errors.Is(err, ErrConfig)
}

func IsAuthError(err error) bool {
	return errors.Is(err, ErrAuth)
}

func IsAPIError(err error) bool {
	return errors.Is(err, ErrAPI)
}

func IsNetworkError(err error) bool {
	return errors.Is(err, ErrNetwork)
}

func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}

func IsFileError(err error) bool {
	return errors.Is(err, ErrFile)
}