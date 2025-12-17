package testutils

import (
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/stretchr/testify/require"
)

// SetupTestRepo creates a temporary directory and initializes a Loam repository in it.
// It returns the absolute path to the temp dir and the initialized repository.
// It fails the test immediately on error.
func SetupTestRepo(t *testing.T, opts ...loam.Option) (string, core.Repository) {
	t.Helper()

	tmpDir := t.TempDir()

	// Loam sometimes prefers absolute paths, though t.TempDir usually returns one.
	// Ensuring it is absolute is safe.
	absPath, err := filepath.Abs(tmpDir)
	require.NoError(t, err, "Failed to get absolute path for temp dir")

	repo, err := loam.Init(absPath, opts...)
	require.NoError(t, err, "Failed to init loam repo")

	return absPath, repo
}
