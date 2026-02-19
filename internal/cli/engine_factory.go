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
	if hasNode(opts.RepoPath, domain.DefaultErrorNodeID) {
		engineOpts = append(engineOpts, trellis.WithDefaultErrorNode(domain.DefaultErrorNodeID))
	}

	// 3. Smart Convention: Entrypoint Fallback
	entryPoint := determineEntryPoint(opts.RepoPath)

	// Only override if different from default "start" to avoid unnecessary config
	if entryPoint != "start" {
		engineOpts = append(engineOpts, trellis.WithEntryNode(entryPoint))
	}

	// 4. Initialize
	engine, err := trellis.New(opts.RepoPath, engineOpts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing engine: %w", err)
	}

	return engine, nil
}

// hasNode checks if a node exists as a file in the directory.
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

// determineEntryPoint implements the fallback logic for finding the initial node.
// Priority: start > main > index > DirectoryName
func determineEntryPoint(repoPath string) string {
	if hasNode(repoPath, "start") {
		return "start"
	}
	if hasNode(repoPath, "main") {
		return "main"
	}
	if hasNode(repoPath, "index") {
		return "index"
	}

	dirName := filepath.Base(repoPath)
	if hasNode(repoPath, dirName) {
		return dirName
	}

	return "start" // Default
}
