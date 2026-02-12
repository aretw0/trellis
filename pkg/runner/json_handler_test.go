package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
)

func TestJSONHandler_Output(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewJSONHandler(buf)

	actions := []domain.ActionRequest{
		{Type: domain.ActionRenderContent, Payload: "Hello Intent"},
		{Type: domain.ActionRequestInput, Payload: domain.InputRequest{Type: domain.InputText}},
	}

	needsInput, err := handler.Output(context.Background(), actions)
	if err != nil {
		t.Fatalf("Output failed: %v", err)
	}

	if !needsInput {
		t.Error("Expected needsInput to be true")
	}

	output := buf.String()
	// Should be a single line of JSON
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line of output, got %d", len(lines))
	}

	var decoded []domain.ActionRequest
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(decoded) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(decoded))
	}
	if decoded[0].Payload.(string) != "Hello Intent" {
		t.Errorf("Payload mismatch")
	}
}

func TestJSONHandler_Input(t *testing.T) {
	// Test Case 1: JSON String
	input := "\"Hello World\""
	handler := NewJSONHandler(nil)

	go func() {
		handler.FeedInput(input, nil)
	}()

	val, err := handler.Input(context.Background())
	if err != nil {
		t.Fatalf("Input failed: %v", err)
	}
	if val != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", val)
	}

	// Test Case 2: Plain Text
	input2 := "just plain text"
	handler2 := NewJSONHandler(nil)

	go func() {
		handler2.FeedInput(input2, nil)
	}()

	val2, err := handler2.Input(context.Background())
	if err != nil {
		t.Fatalf("Input2 failed: %v", err)
	}
	if val2 != "just plain text" {
		t.Errorf("Expected 'just plain text', got '%s'", val2)
	}
}

func TestJSONHandler_SystemOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewJSONHandler(buf)

	err := handler.SystemOutput(context.Background(), "System Status")
	if err != nil {
		t.Fatalf("SystemOutput failed: %v", err)
	}

	output := buf.String()
	// Should be a single line of JSON
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line of output, got %d", len(lines))
	}

	var decoded []domain.ActionRequest
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(decoded) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(decoded))
	}

	if decoded[0].Type != domain.ActionSystemMessage {
		t.Errorf("Expected type SYSTEM_MESSAGE, got %s", decoded[0].Type)
	}
	if decoded[0].Payload.(string) != "System Status" {
		t.Errorf("Payload mismatch")
	}
}
