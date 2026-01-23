package process

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// Runner implements a ToolRunner that executes local processes.
// It follows a Strict Registry pattern for security (Allow-Listing).
type Runner struct {
	registry map[string]RegisteredProcess
}

// RegisteredProcess defines a allowed command execution.
type RegisteredProcess struct {
	Command string
	Args    []string // Default/Template args
}

// NewRunner creates a new Process Runner.
func NewRunner() *Runner {
	return &Runner{
		registry: make(map[string]RegisteredProcess),
	}
}

// Register adds a trusted script/command to the allow-list.
func (r *Runner) Register(name string, command string, args ...string) {
	r.registry[name] = RegisteredProcess{
		Command: command,
		Args:    args,
	}
}

// Execute satisfies the hypothetical ToolRunner interface (or is called by IOHandler).
// For now, checks the toolCall.Name against the registry.
func (r *Runner) Execute(ctx context.Context, toolCall domain.ToolCall) (domain.ToolResult, error) {
	proc, ok := r.registry[toolCall.Name]
	if !ok {
		// Not found in this adapter.
		// In a multi-adapter setup, we might return a specific error or nil to let others try.
		// Here we assume if it's routed here, it must exist.
		return domain.ToolResult{
			ID:      toolCall.ID,
			IsError: true,
			Error:   fmt.Sprintf("process tool not registered: %s", toolCall.Name),
		}, nil
	}

	// Prepare Command
	// Security: We do NOT pass toolCall.Args as direct command flags blindly.
	// Implementation Decision: Pass args as Environment Variables.
	// This prevents flag injection attacks (e.g. passing "; rm -rf /").

	cmd := exec.CommandContext(ctx, proc.Command, proc.Args...)

	// Prepare Environment
	env := []string{}
	// Copy useful fields from toolCall.Args to Key=Value strings
	for k, v := range toolCall.Args {
		// Basic sanitization: keys must be alphanumeric
		// Values are treated as strings.
		val := fmt.Sprintf("%v", v)
		env = append(env, fmt.Sprintf("TRELLIS_ARG_%s=%s", strings.ToUpper(k), val))
	}
	cmd.Env = append(cmd.Environ(), env...)

	// Capture Output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run
	err := cmd.Run()

	result := domain.ToolResult{
		ID: toolCall.ID,
	}

	if err != nil {
		result.IsError = true
		// Combine error message with stderr for context
		result.Error = fmt.Sprintf("execution failed: %v. Stderr: %s", err, stderr.String())

		// If it was an ExitError, we might want to include the exit code?
		// keeping it simple for now.
		return result, nil
	}

	// Success
	// We return stdout as the result.
	// The Host/Engine will handle parsing if it's JSON (via SaveTo logic).
	output := stdout.String()
	// Trim whitespace for cleaner string results?
	// Generally safer to trim.
	result.Result = strings.TrimSpace(output)

	return result, nil
}
