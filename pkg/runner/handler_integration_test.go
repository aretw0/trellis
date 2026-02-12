package runner

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestTextHandler_Input_Sanitization_Retry(t *testing.T) {
	// Setup: Input contains 1 bad line (too long) and 1 good line.
	// We force the Limit to be small via Env is tricky in parallel tests,
	// so we'll rely on generating a huge string > 4096 default,
	// OR we assume the default 4096. Let's use > 4096.

	// Setup
	badInput := strings.Repeat("A", 5000)
	goodInput := "Small valid input"

	outBuf := &bytes.Buffer{}
	handler := NewTextHandler(outBuf)

	// Feed inputs in background
	go func() {
		// Wait for Input() to be called and ready
		time.Sleep(50 * time.Millisecond)
		handler.FeedInput(badInput, nil)

		// Wait for the retry prompt to be printed and Input() to run again
		time.Sleep(50 * time.Millisecond)
		handler.FeedInput(goodInput, nil)
	}()

	// Execute
	val, err := handler.Input(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify Behavior
	// 1. Valid output eventually returned
	if val != "Small valid input" {
		t.Errorf("Expected eventual valid input, got '%s'", val)
	}

	// 2. Output buffer should contain error message
	output := outBuf.String()
	if !strings.Contains(output, "input exceeds maximum allowed size") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
	if !strings.Contains(output, "Please try again") {
		t.Errorf("Expected retry prompt, got: %s", output)
	}
}

func TestJSONHandler_Input_Sanitization_Error(t *testing.T) {
	// Setup: Input contains just bad line. JSONHandler should fail immediately.
	// Setup
	badInput := strings.Repeat("A", 5000)

	// outBuf := &bytes.Buffer{} // Optional for JSON handler?
	handler := NewJSONHandler(nil) // Discard output

	go func() {
		time.Sleep(20 * time.Millisecond)
		handler.FeedInput(badInput, nil)
	}()

	// Execute
	val, err := handler.Input(context.Background())

	// Verify Behavior
	// Should fail
	if err == nil {
		t.Errorf("Expected error for large input, got success: '%s'", val)
	}
	if !strings.Contains(err.Error(), "input exceeds maximum allowed size") {
		t.Errorf("Expected size limit error, got: %v", err)
	}
}
