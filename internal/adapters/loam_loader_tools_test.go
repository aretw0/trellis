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

func TestLoamLoader_Tools(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Create a file with tool definitions
	content := `---
id: with_tools
type: text
tools:
  - name: my_tool
    description: A test tool
    parameters:
      type: object
      properties:
        arg1:
          type: string
---
Node with tools`

	err := os.WriteFile(filepath.Join(tmpDir, "with_tools.md"), []byte(content), 0644)
	require.NoError(t, err)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[dto.NodeMetadata](repo)
	loader := NewLoamLoader(typedRepo)

	// Get Node
	data, err := loader.GetNode("with_tools")
	require.NoError(t, err)

	// Use json to unmarshal and check
	var nodeMap map[string]any
	err = json.Unmarshal(data, &nodeMap)
	require.NoError(t, err)

	tools, ok := nodeMap["tools"].([]any)
	require.True(t, ok, "tools field should be present")
	require.Len(t, tools, 1)

	tool := tools[0].(map[string]any)
	assert.Equal(t, "my_tool", tool["name"])
	assert.Equal(t, "A test tool", tool["description"])

	params, ok := tool["parameters"].(map[string]any)
	require.True(t, ok, "parameters should be a map")
	assert.Equal(t, "object", params["type"])
}
