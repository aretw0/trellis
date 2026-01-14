package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/dto"
	"github.com/aretw0/trellis/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, dir, name, content string) {
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	require.NoError(t, err)
}

func TestLoamLoader_Tools_Polymorphic(t *testing.T) {
	tmpDir, repo := testutils.SetupTestRepo(t)

	// 1. Library File (to be imported)
	libContent := `---
id: lib
type: text
tools:
  - name: lib_tool_1
    description: Tool from library
    parameters: {}
  - name: shared_tool
    description: Shared tool (lib version)
    parameters: {}
---
Library Node`
	writeFile(t, tmpDir, "lib.md", libContent)

	// 2. Intermediate Library (for deep import)
	intermediateContent := `---
id: intermediate
type: text
tools:
  - lib
  - name: intermediate_tool
    description: Tool from intermediate
---
Intermediate Node`
	writeFile(t, tmpDir, "intermediate.md", intermediateContent)

	// 3. Consumer Node (uses tools)
	consumerContent := `---
id: consumer
type: text
tools:
  - intermediate
  - name: local_tool
    description: Tool from local
  - name: shared_tool
    description: Shared tool (local version) -- SHADOWS IMPORT
---
Consumer Node`
	writeFile(t, tmpDir, "consumer.md", consumerContent)

	// 4. Cycle A
	cycleAContent := `---
id: cycle_a
type: text
tools:
  - cycle_b
---
Cycle A`
	writeFile(t, tmpDir, "cycle_a.md", cycleAContent)

	// 5. Cycle B
	cycleBContent := `---
id: cycle_b
type: text
tools:
  - cycle_a
---
Cycle B`
	writeFile(t, tmpDir, "cycle_b.md", cycleBContent)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[dto.NodeMetadata](repo)
	loader := NewLoamLoader(typedRepo)

	t.Run("Resolves Deep Imports and Shadowing", func(t *testing.T) {
		data, err := loader.GetNode("consumer")
		require.NoError(t, err)

		var nodeMap map[string]any
		err = json.Unmarshal(data, &nodeMap)
		require.NoError(t, err)

		toolsRaw, ok := nodeMap["tools"].([]any)
		require.True(t, ok)

		// Convert to map for easy assertion
		toolMap := make(map[string]map[string]any)
		for _, tRaw := range toolsRaw {
			tData := tRaw.(map[string]any)
			toolMap[tData["name"].(string)] = tData
		}

		// Check existence
		assert.Contains(t, toolMap, "lib_tool_1", "Should have deep imported tool")
		assert.Contains(t, toolMap, "intermediate_tool", "Should have directly imported tool")
		assert.Contains(t, toolMap, "local_tool", "Should have local tool")
		assert.Contains(t, toolMap, "shared_tool", "Should have shared tool")

		// Check Shadowing
		assert.Equal(t, "Shared tool (local version) -- SHADOWS IMPORT", toolMap["shared_tool"]["description"], "Local definition should shadow imported one")
	})

	t.Run("Detects Cycles", func(t *testing.T) {
		_, err := loader.GetNode("cycle_a")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cycle detected")
	})

	t.Run("Legacy Inline Only", func(t *testing.T) {
		legacyContent := `---
id: legacy
type: text
tools:
  - name: just_local
    description: desc
---
Legacy`
		writeFile(t, tmpDir, "legacy.md", legacyContent)

		data, err := loader.GetNode("legacy")
		require.NoError(t, err)

		var nodeMap map[string]any
		err = json.Unmarshal(data, &nodeMap)
		require.NoError(t, err)
		tools := nodeMap["tools"].([]any)
		assert.Len(t, tools, 1)
		assert.Equal(t, "just_local", tools[0].(map[string]any)["name"])
	})
}
