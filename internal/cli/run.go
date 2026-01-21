package cli

import (
	"encoding/json"
	"fmt"
)

// RunOptions contains all the configuration for the Run command.
type RunOptions struct {
	RepoPath  string
	Headless  bool
	Watch     bool
	JSON      bool
	Debug     bool
	Context   string // Raw JSON string
	SessionID string
	Fresh     bool
}

// Execute handles the 'run' command logic, dispatching to Session or Watch mode.
func Execute(opts RunOptions) error {
	// Parse initial context if provided
	var initialContext map[string]any
	if opts.Context != "" {
		if err := json.Unmarshal([]byte(opts.Context), &initialContext); err != nil {
			return fmt.Errorf("error parsing --context JSON: %w", err)
		}
	}

	if opts.Watch {
		if opts.Headless {
			return fmt.Errorf("--watch and --headless cannot be used together")
		}
		RunWatch(opts.RepoPath, opts.SessionID, opts.Debug, opts.Fresh)
		return nil
	}

	// Session Mode
	// Handle Fresh reset here for Session mode to mirror Watch mode behavior
	if opts.Fresh {
		ResetSession(opts.SessionID)
	}

	return RunSession(opts.RepoPath, opts.Headless, opts.JSON, opts.Debug, initialContext, opts.SessionID)
}
