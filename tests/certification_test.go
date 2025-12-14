package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/adapters"
)

// TestCertificationSuite runs the certification specs defined in tests/specs.
func TestCertificationSuite(t *testing.T) {
	specsDir := "specs" // Relative to this test file

	entries, err := filepath.Glob(filepath.Join(specsDir, "*"))
	if err != nil {
		t.Fatalf("Failed to list specs: %v", err)
	}

	for _, specPath := range entries {
		specName := filepath.Base(specPath)
		t.Run(specName, func(t *testing.T) {
			runSpec(t, specPath)
		})
	}
}

func runSpec(t *testing.T, sourcePath string) {
	// 1. Setup Isolated Temp Dir for this test
	tempDir := t.TempDir()

	// 2. Copy Spec Data to Temp Dir
	if err := copyDir(sourcePath, tempDir); err != nil {
		t.Fatalf("Failed to copy spec data: %v", err)
	}

	// 3. Init Loam on the Temp Dir
	// Even if IsDevRun is true, pointing to a new TempDir creates a fresh environment.
	// We use absolute path of tempDir.
	absPath, _ := filepath.Abs(tempDir)

	// Note: We intentionally allow Loam to use its default behavior or force temp.
	// Since we are already in a temp dir, it should be fine.
	repo, err := loam.Init(absPath, loam.WithVersioning(false), loam.WithForceTemp(false))
	if err != nil {
		t.Fatalf("Loam init failed: %v", err)
	}
	typedRepo := loam.NewTypedRepository[adapters.NodeMetadata](repo)

	// 4. Crawler / BFS Validation
	ctx := context.Background()
	visited := make(map[string]bool)
	queue := []string{"start"}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		// Load Node
		doc, err := typedRepo.Get(ctx, currentID)
		if err != nil {
			t.Fatalf("Failed to load node '%s': %v", currentID, err)
		}

		// Verify transitions loaded correctly
		if len(doc.Data.Transitions) > 0 {
			to := doc.Data.Transitions[0].To
			if to == "" {
				to = doc.Data.Transitions[0].ToFull
			}
			t.Logf("Transition 0: To=%s", to)
			// Verify that the transition system loaded *something* parsable.
			// (We rely on the crawler loop below to validate that the link target actually exists)
		}

		// Enqueue transitions
		for _, tr := range doc.Data.Transitions {
			to := tr.To
			if to == "" {
				to = tr.ToFull
			}
			if !visited[to] {
				queue = append(queue, to)
			}
		}
	}
}

// copyDir recursively copies a directory tree, attempting to preserve permissions.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}
