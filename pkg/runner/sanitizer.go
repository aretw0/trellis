package runner

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// DefaultMaxInputSize is 4KB (conservative default)
	DefaultMaxInputSize = 4096
	// EnvMaxInputSize is the environment variable to override the default
	EnvMaxInputSize = "TRELLIS_MAX_INPUT_SIZE"
)

var (
	ErrInputTooLarge = errors.New("input exceeds maximum allowed size")
	ErrInvalidUTF8   = errors.New("input contains invalid UTF-8 sequences")
)

// SanitizeInput cleans user input by enforcing size limits,
// validating UTF-8, and stripping dangerous control characters.
func SanitizeInput(input string) (string, error) {
	// 1. Enforce Size Limit
	limit := getMaxInputSize()
	if len(input) > limit {
		// We explicitly reject rather than truncate to ensure deterministic state.
		return "", fmt.Errorf("%w: size=%d limit=%d", ErrInputTooLarge, len(input), limit)
	}

	// 2. Validate UTF-8
	if !utf8.ValidString(input) {
		return "", ErrInvalidUTF8
	}

	// 3. Strip Control Characters
	// We preserve:
	// - Newline (\n)
	// - Tab (\t)
	// - Carriage Return (\r) - treated as whitespace
	// We remove:
	// - ANSI codes (ESC), NULL, BEL, etc.
	// This prevents log poisoning and terminal corruption.

	// Fast path: if no control chars, return as is.
	clean := true
	for _, r := range input {
		if unicode.IsControl(r) && !isSafeControl(r) {
			clean = false
			break
		}
	}
	if clean {
		return input, nil
	}

	// Slow path: build clean string
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		if !unicode.IsControl(r) || isSafeControl(r) {
			b.WriteRune(r)
		}
	}
	return b.String(), nil
}

func isSafeControl(r rune) bool {
	return r == '\n' || r == '\t' || r == '\r'
}

func getMaxInputSize() int {
	if val := os.Getenv(EnvMaxInputSize); val != "" {
		if size, err := strconv.Atoi(val); err == nil && size > 0 {
			return size
		}
	}
	return DefaultMaxInputSize
}
