package loam

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/dto"
	"github.com/aretw0/trellis/internal/testutils"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoamLoader_Subgraph(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Create subdirectory structure
	// modules/checkout/start.md (Implicit ID)
	subDir := filepath.Join(tmpDir, "modules", "checkout")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	err := os.WriteFile(filepath.Join(subDir, "start.md"), []byte(`---
type: text
---
Checkout Start`), 0644)
	require.NoError(t, err)

	// Create root file using jump_to
	// intro.md
	err = os.WriteFile(filepath.Join(tmpDir, "intro.md"), []byte(`---
id: intro
type: text
transitions:
  - jump_to: modules/checkout/start
---
Intro`), 0644)
	require.NoError(t, err)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[dto.NodeMetadata](repo)
	loader := New(typedRepo)

	t.Run("Discovery of Namespaced IDs", func(t *testing.T) {
		ids, err := loader.ListNodes()
		require.NoError(t, err)

		// Expect normalized path separator (forward slash on all OS).
		// Trellis IDs are URI-like, so the adapter must ensure they are consistent.
		assert.Contains(t, ids, "intro")
		assert.Contains(t, ids, "modules/checkout/start")
	})

	t.Run("JumpTo is mapped to ToNodeID", func(t *testing.T) {
		// Get the intro node
		data, err := loader.GetNode("intro")
		require.NoError(t, err)

		var node domain.Node
		err = json.Unmarshal(data, &node)
		require.NoError(t, err)

		require.Len(t, node.Transitions, 1)
		// Verify that "jump_to" in YAML became "to_node_id" in Domain JSON
		assert.Equal(t, "modules/checkout/start", node.Transitions[0].ToNodeID)
	})
}
