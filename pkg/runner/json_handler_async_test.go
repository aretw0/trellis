package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aretw0/trellis/pkg/domain"
)

func TestJSONHandler_AsyncInput(t *testing.T) {
	// 1. Test Cancellation
	t.Run("ContextCancellation", func(t *testing.T) {
		handler := NewJSONHandler(nil) // Output to discard

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := handler.Input(ctx)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if ctx.Err() != context.DeadlineExceeded && ctx.Err() != context.Canceled {
			t.Errorf("Expected context error, got: %v", err)
		}
	})

	// 2. Test Success
	t.Run("Success", func(t *testing.T) {
		input := "Valid JSON String"
		handler := NewJSONHandler(nil)

		go func() {
			time.Sleep(10 * time.Millisecond)
			handler.FeedInput(input, nil)
		}()

		ctx := context.Background()
		val, err := handler.Input(ctx)
		if err != nil {
			t.Fatalf("Input failed: %v", err)
		}
		if val != "Valid JSON String" {
			t.Errorf("Expected 'Valid JSON String', got '%s'", val)
		}
	})

	// 3. Test EOF
	t.Run("EOF", func(t *testing.T) {
		handler := NewJSONHandler(nil)

		go func() {
			time.Sleep(10 * time.Millisecond)
			handler.FeedInput("", io.EOF)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := handler.Input(ctx)
		if err != io.EOF {
			t.Errorf("Expected EOF, got: %v", err)
		}
	})
}

func TestJSONHandler_AsyncHandleTool(t *testing.T) {
	// 1. Test Cancellation
	t.Run("Cancellation", func(t *testing.T) {
		// Capture output to avoid noise
		outBuf := &bytes.Buffer{}
		handler := NewJSONHandler(outBuf)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Call tool "test"
		call := domain.ToolCall{ID: "1", Name: "test"}
		_, err := handler.HandleTool(ctx, call)

		if err == nil {
			t.Error("Expected error from HandleTool, got nil")
		}
	})

	// 2. Test Success (JSONL Response)
	t.Run("Success", func(t *testing.T) {
		// Prepare a mocked response
		result := domain.ToolResult{
			ID:     "1",
			Result: "Success Result",
		}
		resultJSON, _ := json.Marshal(result)
		input := string(resultJSON)

		outBuf := &bytes.Buffer{}
		handler := NewJSONHandler(outBuf)

		// Mock the user response (which in JSON mode is just a JSON line)
		go func() {
			time.Sleep(20 * time.Millisecond)
			handler.FeedInput(input, nil)
		}()

		ctx := context.Background()
		call := domain.ToolCall{ID: "1", Name: "test"}

		res, err := handler.HandleTool(ctx, call)
		if err != nil {
			t.Fatalf("HandleTool failed: %v", err)
		}
		if res.Result != "Success Result" {
			t.Errorf("Expected 'Success Result', got '%s'", res.Result)
		}

		// Verify request was written
		encodedReq := strings.TrimSpace(outBuf.String())
		if !strings.Contains(encodedReq, `"CALL_TOOL"`) {
			t.Errorf("Output did not contain 'CALL_TOOL', got: %s", encodedReq)
		}
	})
}
