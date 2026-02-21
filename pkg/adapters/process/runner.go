package process

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/lifecycle/pkg/core/worker"
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

// resolveProcess looks up the command in the registry or tries inline execution if allowed.
func (r *Runner) resolveProcess(name string, metadata map[string]string) (RegisteredProcess, bool) {
	// 1. Try Registry
	proc, ok := r.registry[name]
	if ok {
		return proc, true
	}

	// 2. Try Inline (Ad-Hoc) if allowed
	if !r.allowInline || metadata == nil {
		return RegisteredProcess{}, false
	}

	// Support x-exec-command and x-exec-args (set by Loam parser)
	cmd, exists := metadata["x-exec-command"]
	if !exists {
		return RegisteredProcess{}, false
	}

	argsStr := metadata["x-exec-args"]
	var args []string
	if argsStr != "" {
		args = strings.Fields(argsStr)
	}

	return RegisteredProcess{
		Command: cmd,
		Args:    args,
	}, true
}

// Execute triggers the execution of a registered or inline process tool.
func (r *Runner) Execute(ctx context.Context, toolCall domain.ToolCall) (domain.ToolResult, error) {
	proc, ok := r.resolveProcess(toolCall.Name, toolCall.Metadata)
	if !ok {
		// Not found in this adapter.
		return domain.ToolResult{
			ID:      toolCall.ID,
			IsError: true,
			Error:   fmt.Sprintf("process tool not registered: %s (and inline execution not enabled/found)", toolCall.Name),
		}, nil
	}

	// Prepare Environment
	argsJSON, err := json.Marshal(toolCall.Args)
	if err != nil {
		return domain.ToolResult{
			ID:      toolCall.ID,
			IsError: true,
			Error:   fmt.Sprintf("failed to marshal tool arguments: %v", err),
		}, nil
	}

	// Use lifecycle ProcessWorker for graceful shutdown
	w := lifecycle.NewProcessWorker(toolCall.ID, proc.Command, proc.Args...)
	w.SetEnv("TRELLIS_ARGS", string(argsJSON))

	// Capture Output
	var stdout, stderr bytes.Buffer
	w.SetOutput(&stdout, &stderr)

	execErr := r.runWorker(ctx, w)

	result := domain.ToolResult{
		ID: toolCall.ID,
	}

	if execErr != nil {
		result.IsError = true
		// Combine error message with stderr for context
		result.Error = fmt.Sprintf("execution failed: %v. Stderr: %s", execErr, stderr.String())
		return result, nil
	}

	result.Result = r.parseResult(stdout.String())
	return result, nil
}

// TeardownTimeout is the failsafe duration to wait for a zombie process to die
// after a Stop/Kill signal before giving up to avoid deep deadlocks.
// This is especially important for slow environments like WSL/CI.
const TeardownTimeout = 3 * time.Second

// runWorker starts the worker and waits for completion or context cancellation.
func (r *Runner) runWorker(ctx context.Context, w *worker.ProcessWorker) error {
	if err := w.Start(ctx); err != nil {
		return err
	}

	done := w.Wait()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Timeout or Interrupt -> Trigger Graceful Shutdown
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try to stop, but don't return immediately on error.
		// We must wait for the process to actually die/cleanup to avoid data races.
		stopErr := w.Stop(stopCtx)

		// Wait for the worker to fully cleanup (ensure IO flush)
		select {
		case <-done:
		case <-time.After(TeardownTimeout):
			// Failsafe: prevent deadlock
		}

		if stopErr != nil {
			return fmt.Errorf("stopped with error: %w (original context: %v)", stopErr, ctx.Err())
		}

		return ctx.Err()
	}
}

// parseResult attempts to parse the tool output as JSON, falling back to a trimmed string.
func (r *Runner) parseResult(output string) any {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return ""
	}

	// Try to parse as JSON (Auto-Detection)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		var jsonResult any
		if err := json.Unmarshal([]byte(trimmed), &jsonResult); err == nil {
			return jsonResult
		}
	}

	return trimmed
}
