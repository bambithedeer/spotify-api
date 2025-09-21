package logger

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		format      string
		output      string
		expectError bool
	}{
		{"valid config", "info", "text", "stdout", false},
		{"json format", "debug", "json", "stderr", false},
		{"invalid level", "invalid", "text", "stdout", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.level, tt.format, tt.output)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	logger, err := New("warn", "text", "stdout")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// This should not panic or cause issues
	logger.Info("This should not be logged")
	logger.Warn("This should be logged")
	logger.Error("This should be logged")
}