package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUniqueTestFiles creates sample files with unique basenames to avoid ID conflicts.
func setupUniqueTestFiles(t *testing.T, dir string) {
	t.Helper()

	jsonContent := `{"val": 9007199254740991}`
	os.WriteFile(filepath.Join(dir, "file_json.json"), []byte(jsonContent), 0644)

	yamlContent := "val: 9007199254740991"
	os.WriteFile(filepath.Join(dir, "file_yaml.yaml"), []byte(yamlContent), 0644)

	mdContent := "---\nval: 9007199254740991\n---\n"
	os.WriteFile(filepath.Join(dir, "file_md.md"), []byte(mdContent), 0644)
}

// TestNormalization verifies that when WithStrict(true) is enabled,
// ALL formats (JSON, YAML, Markdown) return consistent numeric types (json.Number).
func TestNormalization(t *testing.T) {
	tmpDir := t.TempDir()
	setupUniqueTestFiles(t, tmpDir)

	// Init Loam with Global Strict Mode
	repo, err := loam.Init(tmpDir, loam.WithStrict(true))
	require.NoError(t, err)

	ctx := context.Background()
	files := []string{"file_json", "file_yaml", "file_md"}

	for _, id := range files {
		t.Run(id, func(t *testing.T) {
			doc, err := repo.Get(ctx, id)
			require.NoError(t, err)
			require.NotNil(t, doc.Metadata)

			val, ok := doc.Metadata["val"]
			require.True(t, ok)

			// Expectation: Strict mode -> json.Number (or compatible string-based number)
			// assert.IsType(t, json.Number(""), val, "Should be json.Number in strict mode")
			// Depending on implementation, checking string rep of type is safer for generic assertions
			assert.Equal(t, "json.Number", fmt.Sprintf("%T", val), "Type mismatch for %s", id)

			// Verify value correctness
			// json.Number prints as the original string number
			s := fmt.Sprintf("%v", val)
			assert.Equal(t, "9007199254740991", s)
		})
	}
}

// TestNoNormalization verifies the REGRESSION: when WithStrict is NOT enabled (default),
// we get inconsistent types (Schema Drift risk).
// JSON -> float64 (Dangerous for large ints)
// YAML/Markdown -> int (Go default for YAML)
func TestNoNormalization(t *testing.T) {
	tmpDir := t.TempDir()
	setupUniqueTestFiles(t, tmpDir)

	// Init Loam WITHOUT strict mode (Default)
	repo, err := loam.Init(tmpDir) // No options
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Check JSON (Expected: float64)
	docJSON, err := repo.Get(ctx, "file_json")
	require.NoError(t, err)
	valJSON := docJSON.Metadata["val"]
	assert.Equal(t, "float64", fmt.Sprintf("%T", valJSON), "Default JSON should be float64")

	// 2. Check YAML (Expected: int or int64 depending on parser default)
	docYAML, err := repo.Get(ctx, "file_yaml")
	require.NoError(t, err)
	valYAML := docYAML.Metadata["val"]
	// Usually int in 64-bit systems
	isInt := fmt.Sprintf("%T", valYAML) == "int" || fmt.Sprintf("%T", valYAML) == "int64"
	assert.True(t, isInt, "Default YAML should be int/int64, got %T", valYAML)

	// 3. Check Markdown (Expected: int same as YAML)
	docMD, err := repo.Get(ctx, "file_md")
	require.NoError(t, err)
	valMD := docMD.Metadata["val"]
	isIntMD := fmt.Sprintf("%T", valMD) == "int" || fmt.Sprintf("%T", valMD) == "int64"
	assert.True(t, isIntMD, "Default Markdown should be int/int64, got %T", valMD)

	// 4. Assert Inconsistency
	assert.NotEqual(t, fmt.Sprintf("%T", valJSON), fmt.Sprintf("%T", valMD),
		"Without Strict mode, types should be INCONSISTENT (float64 vs int)")
}
