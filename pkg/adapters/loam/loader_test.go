package loam

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"

	"github.com/aretw0/trellis/internal/testutils"
	"github.com/aretw0/trellis/pkg/ports/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Contract(t *testing.T) {
	// 1. Setup Loam (via testutils which ensures Strict Mode)
	// dir, repo := testutils.SetupTestRepo(t)
	// Actually we need the directory path only if we were writing files manually.
	// But `repo.Save` handles saving documents.
	// We need `repo` for the adapter.

	_, repo := testutils.SetupTestRepo(t)

	// 2. Setup Data
	ctx := context.Background()

	// Node A: Regular
	// Node B: With transitions

	setupData := map[string][]byte{
		"a": []byte(`{"content":"Tm9kZSBB","id":"a","transitions":[],"type":"text"}`),                   // Base64 "Node A"
		"b": []byte(`{"content":"Tm9kZSBC","id":"b","transitions":[{"to_node_id":"a"}],"type":"text"}`), // Base64 "Node B"
	}

	// We need to save these as Loam documents first
	// Note: Loader.GetNode returns JSON compatible with domain.Node.
	// But to populate Loam, we write Markdown/YAML.

	docA := core.Document{
		ID: "a.md",
		Content: `---
id: a
type: text
---
Node A`,
	}

	docB := core.Document{
		ID: "b.md",
		Content: `---
id: b
type: text
transitions:
  - to: a
---
Node B`,
	}

	if err := repo.Save(ctx, docA); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, docB); err != nil {
		t.Fatal(err)
	}

	// 3. Create Adapter
	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

	// 4. Run Contract
	tests.GraphLoaderContractTest(t, loader, setupData)
}

func TestLoader_ListNodes_NormalizesIDs(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

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
	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

	// Execute ListNodes
	ids, err := loader.ListNodes()
	require.NoError(t, err)

	// Verify IDs are normalized (extensions stripped)
	assert.Contains(t, ids, "start", "start.md should become start")
	assert.Contains(t, ids, "choice", "choice.json should become choice")
	assert.Contains(t, ids, "implicit", "implicit.md should become implicit")
	assert.Len(t, ids, 3)
}

func TestLoader_ListNodes_DetectsCollisions(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Seed files that result in the same ID
	files := map[string]string{
		"foo.md": `---
id: foo
type: text
---
Explicit ID`,
		"foo.json": `{
  "id": "foo",
  "type": "text"
}`,
	}

	for filename, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

	// Execute ListNodes - Should Fail
	_, err := loader.ListNodes()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collision detected")
	assert.Contains(t, err.Error(), "foo")
}

func TestLoader_GetNode_NormalizesID(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Create a file with explicit ID having extension
	err := os.WriteFile(filepath.Join(tmpDir, "node.json"), []byte(`{ "id": "node.json", "type": "text" }`), 0644)
	require.NoError(t, err)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

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

func TestLoader_ToolCall_DefaultsID(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Create a node with tool_call missing an explicit ID
	content := `---
id: implicit_tool
type: tool
tool_call:
  name: my_awesome_tool
  args:
    foo: bar
---`
	err := os.WriteFile(filepath.Join(tmpDir, "implicit_tool.md"), []byte(content), 0644)
	require.NoError(t, err)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

	// Execute GetNode
	data, err := loader.GetNode("implicit_tool")
	require.NoError(t, err)

	// Verify JSON output
	// The loader should have copied "name" to "id"
	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"id":"my_awesome_tool"`)
}

func TestLoader_DefaultContext(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Create a node with default_context
	content := `---
id: start
type: start
default_context:
  env: dev
  retries: 5
---
# Start`
	err := os.WriteFile(filepath.Join(tmpDir, "start.md"), []byte(content), 0644)
	require.NoError(t, err)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

	// Execute GetNode
	data, err := loader.GetNode("start")
	require.NoError(t, err)

	// Verify JSON output
	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"default_context":{`)
	assert.Contains(t, jsonStr, `"env":"dev"`)
	// Note: We check specifically for the key-value pair in JSON
}

func TestLoader_ContextSchema(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Create a node with context_schema
	content := `---
id: start
type: start
context_schema:
  api_key: string
  retries: int
  tags: [string]
---
# Start`
	err := os.WriteFile(filepath.Join(tmpDir, "start.md"), []byte(content), 0644)
	require.NoError(t, err)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

	// Execute GetNode
	data, err := loader.GetNode("start")
	require.NoError(t, err)

	// Verify JSON output
	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"context_schema":{`)
	assert.Contains(t, jsonStr, `"api_key":"string"`)
	assert.Contains(t, jsonStr, `"retries":"int"`)
	assert.Contains(t, jsonStr, `"tags":"[string]"`)
}

func TestLoader_TransitionShorthand(t *testing.T) {
	// Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Create a node using the "to" shorthand
	content := `---
id: jump
type: text
to: destination
---
Jumping...`
	err := os.WriteFile(filepath.Join(tmpDir, "jump.md"), []byte(content), 0644)
	require.NoError(t, err)

	// Initialize Adapter
	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

	// Execute GetNode
	data, err := loader.GetNode("jump")
	require.NoError(t, err)

	// Verify JSON output
	// The loader should have converted "to" -> transitions: [{to_node_id: destination}]
	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"to_node_id":"destination"`)
}
