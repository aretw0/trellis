package adapters_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoamLoader_ListNodes_NormalizesIDs(t *testing.T) {
	// Setup Temp Repository
	tmpDir := t.TempDir()
	repo, err := loam.Init(tmpDir, loam.WithVersioning(false))
	require.NoError(t, err)

	// Seed files with various extensions
	files := map[string]string{
		"start.md": `---
id: start.md
type: text
---
Hello`,
		"choice.json": `{
  "id": "choice.json",
  "type": "question"
}`,
		"implicit.md": `---
type: text
---
ID is implied from filename`,
	}

	for filename, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[adapters.NodeMetadata](repo)
	loader := adapters.NewLoamLoader(typedRepo)

	// Execute ListNodes
	ids, err := loader.ListNodes()
	require.NoError(t, err)

	// Verify IDs are normalized (extensions stripped)
	assert.Contains(t, ids, "start", "start.md should become start")
	assert.Contains(t, ids, "choice", "choice.json should become choice")
	assert.Contains(t, ids, "implicit", "implicit.md should become implicit")
	assert.Len(t, ids, 3)
}

func TestLoamLoader_GetNode_NormalizesID(t *testing.T) {
	// Setup Temp Repository
	tmpDir := t.TempDir()
	repo, err := loam.Init(tmpDir, loam.WithVersioning(false))
	require.NoError(t, err)

	// Create a file with explicit ID having extension
	err = os.WriteFile(filepath.Join(tmpDir, "node.json"), []byte(`{ "id": "node.json", "type": "text" }`), 0644)
	require.NoError(t, err)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[adapters.NodeMetadata](repo)
	loader := adapters.NewLoamLoader(typedRepo)

	// Execute GetNode using the normalized name "node"
	data, err := loader.GetNode("node")
	require.NoError(t, err)
	assert.NotNil(t, data)

	// Verify the JSON content has normalized ID
	// We need to unmarshal to check, as GetNode returns bytes
	// But simply checking string containment is a quick proxy
	assert.Contains(t, string(data), `"id":"node"`, "JSON output should have normalized ID")
	assert.NotContains(t, string(data), `"id":"node.json"`)
}
