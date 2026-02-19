package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineEntryPoint(t *testing.T) {
	// Helper to create a temp dir with specific files
	createDir := func(t *testing.T, files []string) string {
		dir := t.TempDir()
		for _, f := range files {
			err := os.WriteFile(filepath.Join(dir, f), []byte("content"), 0644)
			require.NoError(t, err)
		}
		return dir
	}

	t.Run("Default to start if exists", func(t *testing.T) {
		dir := createDir(t, []string{"start.md", "main.md"})
		assert.Equal(t, "start", determineEntryPoint(dir))
	})

	t.Run("Fallback to main", func(t *testing.T) {
		dir := createDir(t, []string{"main.md", "index.md"})
		assert.Equal(t, "main", determineEntryPoint(dir))
	})

	t.Run("Fallback to index", func(t *testing.T) {
		dir := createDir(t, []string{"index.md", "other.md"})
		assert.Equal(t, "index", determineEntryPoint(dir))
	})

	t.Run("Fallback to DirectoryName", func(t *testing.T) {
		// We need to create a directory structure where the leaf dir name matches a file
		tmpRoot := t.TempDir()
		moduleDir := filepath.Join(tmpRoot, "checkout")
		err := os.Mkdir(moduleDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(moduleDir, "checkout.md"), []byte("content"), 0644)
		require.NoError(t, err)

		assert.Equal(t, "checkout", determineEntryPoint(moduleDir))
	})

	t.Run("Default to start if nothing matches", func(t *testing.T) {
		dir := createDir(t, []string{"other.md"})
		assert.Equal(t, "start", determineEntryPoint(dir))
	})
}
