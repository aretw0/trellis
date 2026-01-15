package runner

import (
	"strings"
	"testing"
)

func TestSanitizeInput_SizeLimit(t *testing.T) {
	// Default Limit is 4096
	limit := 4096

	tests := []struct {
		name      string
		inputSize int
		wantErr   bool
	}{
		{"Under Limit", limit - 1, false},
		{"Exact Limit", limit, false},
		{"Over Limit", limit + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.Repeat("a", tt.inputSize)
			_, err := SanitizeInput(input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SanitizeInput() expected error for size %d, got nil", tt.inputSize)
				}
			} else {
				if err != nil {
					t.Errorf("SanitizeInput() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSanitizeInput_ControlChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal Text", "Hello World", "Hello World"},
		{"Safe Controls", "Line1\nLine2\tTabbed", "Line1\nLine2\tTabbed"},
		{"ANSI Code", "\x1b[31mRed\x1b[0m", "[31mRed[0m"}, // ESC removed
		{"Null Byte", "Null\x00Byte", "NullByte"},         // NULL removed
		{"Bell", "Ding\x07", "Ding"},                      // BEL removed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeInput(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestSanitizeInput_EnvOverride(t *testing.T) {
	t.Setenv("TRELLIS_MAX_INPUT_SIZE", "10")

	// Input len 11 -> Should fail
	_, err := SanitizeInput("12345678901")
	if err == nil {
		t.Error("Expected error for input > 10 when env var is set")
	}

	// Input len 5 -> Should pass
	_, err = SanitizeInput("12345")
	if err != nil {
		t.Error("Unexpected error for valid input")
	}
}

func TestSanitizeInput_InvalidUTF8(t *testing.T) {
	// Invalid UTF-8 sequence
	input := "\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98"
	_, err := SanitizeInput(input)
	if err != ErrInvalidUTF8 {
		t.Errorf("Expected ErrInvalidUTF8, got %v", err)
	}
}
