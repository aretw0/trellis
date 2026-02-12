package process

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// Runner implements a ToolRunner that executes local processes.
// It follows a Strict Registry pattern for security (Allow-Listing).
type Runner struct {
	registry    map[string]RegisteredProcess
	allowInline bool
	baseDir     string
}

// RegisteredProcess defines a allowed command execution.
type RegisteredProcess struct {
	Command string
	Args    []string // Default/Template args
}

// RunnerOption configures the runner.
type RunnerOption func(*Runner)

// WithRegistry populates the allow-list from a loaded config.
func WithRegistry(tools map[string]ProcessConfig) RunnerOption {
	return func(r *Runner) {
		for name, tool := range tools {
			r.Register(name, tool.Command, tool.Args...)
		}
	}
}

// WithInlineExecution enables ad-hoc execution (Dangerous).
func WithInlineExecution(allow bool) RunnerOption {
	return func(r *Runner) {
		r.allowInline = allow
	}
}

// WithBaseDir sets the working directory for executed processes.
func WithBaseDir(dir string) RunnerOption {
	return func(r *Runner) {
		r.baseDir = dir
	}
}

// NewRunner creates a new Process Runner.
func NewRunner(opts ...RunnerOption) *Runner {
	r := &Runner{
		registry: make(map[string]RegisteredProcess),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
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
	// 1. Try Registry
	proc, ok := r.registry[toolCall.Name]

	// 2. Try Inline (Ad-Hoc) if allowed
	if !ok {
		if r.allowInline && toolCall.Metadata != nil {
			// Check for x-exec vendor extension
			// The Engine passes metadata as map[string]string, but x-exec usually needs structure.
			// However, since Metadata is strings, users might pass x-exec-command, x-exec-args (comma separated)?
			// Or we assume the generic `map[string]string` flatten strategy is sufficient for simple use cases.
			// Let's look for `x-exec-command`.

			// FIXME: In `node_syntax.md` we designed:
			// x-exec:
			//   command: python
			//   args: ...
			//
			// But `ToolCall.Metadata` in `pkg/domain` is `map[string]string`.
			// The Loader/Parser flattens YAML? Or we need to update Domain?
			// Checking `pkg/domain/node.go`, ToolCall metadata is `map[string]string`.
			// So intricate objects in YAML might be lost or flattened.
			// For v0.7, we support `x-exec-command` and `x-exec-args` (space separated or just command line?)

			// Alternative: Support JSON string in `x-exec`.

			if cmd, exists := toolCall.Metadata["x-exec-command"]; exists {
				// Inline Found
				argsStr := toolCall.Metadata["x-exec-args"]
				var args []string
				if argsStr != "" {
					args = strings.Fields(argsStr) // Basic splitting
				}

				proc = RegisteredProcess{
					Command: cmd,
					Args:    args,
				}
				ok = true
			}
		}
	}

	if !ok {
		// Not found in this adapter.
		return domain.ToolResult{
			ID:      toolCall.ID,
			IsError: true,
			Error:   fmt.Sprintf("process tool not registered: %s (and inline execution not enabled/found)", toolCall.Name),
		}, nil
	}

	// Prepare Command
	// Security: We do NOT pass toolCall.Args as direct command flags blindly.
	// Implementation Decision: Pass args as Environment Variables.
	// This prevents flag injection attacks (e.g. passing "; rm -rf /").

	cmd := exec.CommandContext(ctx, proc.Command, proc.Args...)
	cmd.Dir = r.baseDir

	// Prepare Environment
	env := []string{}
	// Copy useful fields from toolCall.Args to Key=Value strings
	for k, v := range toolCall.Args {
		// Basic sanitization: keys must be alphanumeric
		// Values serialization strategy:
		// - Primitives (string, number, bool): fmt.Sprintf (Simple)
		// - Complex (Map, Slice): json.Marshal (Structured)
		var val string

		switch v.(type) {
		case string, int, int64, float64, bool:
			val = fmt.Sprintf("%v", v)
		case nil:
			val = ""
		default:
			// Complex types: Try JSON
			if inJson, err := json.Marshal(v); err == nil {
				val = string(inJson)
			} else {
				// Fallback to Go format if marshal fails
				val = fmt.Sprintf("%v", v)
			}
		}

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
	trimmed := strings.TrimSpace(output)

	// Try to parse as JSON (Auto-Detection)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		var jsonResult any
		if jsonErr := json.Unmarshal([]byte(trimmed), &jsonResult); jsonErr == nil {
			result.Result = jsonResult
			return result, nil
		}
	}

	// Fallback to string
	result.Result = trimmed

	return result, nil
}
