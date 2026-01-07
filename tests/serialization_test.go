package tests

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/adapters/fs"
	"github.com/stretchr/testify/require"
)

// TestStrictJSONSerialization verifies that the Loam configuration used by Trellis
// (specifically fs.NewJSONSerializer(true)) correctly preserves integer types
// during JSON unmarshaling, preventing the "float64 invasion" in map[string]any.
//
// This test mimics the internal initialization done in trellis.New.
func TestStrictJSONSerialization(t *testing.T) {
	// 1. Setup
	tmpDir := t.TempDir()

	// Create a JSON file with specific numeric types
	// "large_int": a number that definitely stays int (e.g. timestamp or ID)
	// "classic_int": standard 123
	// "real_float": 1.23
	content := `{
		"large_int": 9007199254740991, 
		"classic_int": 123,
		"real_float": 1.23
	}`
	// Note: 9007199254740991 is MAX_SAFE_INTEGER

	err := os.WriteFile(filepath.Join(tmpDir, "numbers.json"), []byte(content), 0644)
	require.NoError(t, err)

	// 2. Initialize Loam with the Strict Serializer (SAME CONFIG AS TRELLIS)
	repo, err := loam.Init(tmpDir,
		loam.WithSerializer(".json", fs.NewJSONSerializer(true)),
	)
	require.NoError(t, err)

	// 3. Get document from UNTYPED repo to inspect raw types.
	// This ensures we are testing the Serializer's behavior (Strict vs Default) configuration.
	ctx := context.Background()
	doc, err := repo.Get(ctx, "numbers")
	require.NoError(t, err)
	require.NotNil(t, doc)

	// 4. Inspect Types in doc.Metadata (map[string]any)
	data := doc.Metadata
	require.NotNil(t, data, "Metadata should be populated")

	// Verify large_int
	valLarge, ok := data["large_int"]
	require.True(t, ok, "large_int missing from Metadata")
	t.Logf("large_int type: %T, value: %v", valLarge, valLarge)

	// With strict=true, we expect int64 or json.Number.
	// Default (strict=false) would be float64.
	switch v := valLarge.(type) {
	case float64:
		require.Fail(t, "Strict mode failed: large_int is float64")
	case int64, int:
		// Success
	case json.Number:
		// Success (if configured to use Number, checking it converts to int64)
		_, err := v.Int64()
		require.NoError(t, err, "json.Number should compile to Int64")
	default:
		t.Logf("Got type %T", v)
		// If it's something else, we might need to adjust expectation, but definitely NOT float64.
	}

	// Verify classic_int
	valClassic, ok := data["classic_int"]
	require.True(t, ok)
	switch v := valClassic.(type) {
	case float64:
		require.Fail(t, "Strict mode failed: classic_int is float64")
	case int64, int, json.Number:
		// Success
		_ = v
	}

	// Verify real_float
	valFloat, ok := data["real_float"]
	require.True(t, ok)
	switch v := valFloat.(type) {
	case float64:
		// Expected for actual floats
	default:
		// json.Number is also valid
		if _, ok := v.(json.Number); ok {
			// valid
		} else {
			t.Logf("Unexpected real_float type: %T", v)
		}
	}
}
