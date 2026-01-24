package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
)

// createEngine initializes a Trellis engine with standard CLI conventions.
func createEngine(opts RunOptions, logger *slog.Logger) (*trellis.Engine, error) {
	engineOpts := []trellis.Option{}

	// 1. Logger & Hooks
	if opts.Debug {
		engineOpts = append(engineOpts, trellis.WithLogger(logger))
		engineOpts = append(engineOpts, trellis.WithLifecycleHooks(createDebugHooks(logger)))
	} else {
		// Even in non-debug, use the provided logger (standardized)
		engineOpts = append(engineOpts, trellis.WithLogger(logger))
	}

	// 2. Smart Convention: Default Error Node
	// If an 'error.md' (or .json/.yaml) exists in the repo, automatically use it as fallback.
	// This avoids "magic" behavior if the project doesn't want an error node.
	if hasNode(opts.RepoPath, domain.DefaultErrorNodeID) {
		engineOpts = append(engineOpts, trellis.WithDefaultErrorNode(domain.DefaultErrorNodeID))
	}

	// 3. Initialize
	engine, err := trellis.New(opts.RepoPath, engineOpts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing engine: %w", err)
	}

	return engine, nil
}

// hasNode checks if a node exists as a file in the directory.
// Note: This is an optimization for the factory to avoid over-configuring
// the engine with fallbacks that don't exist.
func hasNode(repoPath, nodeID string) bool {
	extensions := []string{".md", ".yaml", ".json"}
	for _, ext := range extensions {
		path := filepath.Join(repoPath, nodeID+ext)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}
