package runner

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
)

func TestTextHandler_Output(t *testing.T) {
	outBuf := &bytes.Buffer{}
	handler := NewTextHandler(outBuf)

	// Mock Renderer (optional)
	handler.Renderer = func(s string) (string, error) {
		return "Rendered: " + s, nil
	}

	actions := []domain.ActionRequest{
		{Type: domain.ActionRenderContent, Payload: "Hello World"},
		{Type: domain.ActionRequestInput},
	}

	needsInput, err := handler.Output(context.Background(), actions)
	if err != nil {
		t.Fatalf("Output failed: %v", err)
	}

	if !needsInput {
		t.Error("Expected output to return true for needsInput")
	}

	// Verify Output
	output := outBuf.String()
	expected := "Rendered: Hello World"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain '%s', got '%s'", expected, output)
	}
}

func TestTextHandler_Input(t *testing.T) {
	inputStr := "my user input"
	outBuf := &bytes.Buffer{}

	handler := NewTextHandler(outBuf)

	// Feed input asynchronously to simulate bridge
	go func() {
		handler.FeedInput(inputStr, nil)
	}()

	val, err := handler.Input(context.Background())
	if err != nil {
		t.Fatalf("Input failed: %v", err)
	}

	if val != "my user input" {
		t.Errorf("Expected 'my user input', got '%s'", val)
	}

	// Verify Prompt was written
	prompt := outBuf.String()
	if prompt != "> " {
		t.Errorf("Expected prompt '> ', got '%s'", prompt)
	}
}
