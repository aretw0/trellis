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

func TestLoamLoader_TimeoutMapping(t *testing.T) {
	// 1. Setup Temp Repository
	tmpDir, repo := testutils.SetupTestRepo(t)

	// 2. Create a node with timeout
	content := `---
id: step_timeout
type: text
timeout: 5s
---
Waiting...`
	err := os.WriteFile(filepath.Join(tmpDir, "step_timeout.md"), []byte(content), 0644)
	require.NoError(t, err)

	// 3. Initialize Adapter
	typedRepo := loam.NewTypedRepository[dto.NodeMetadata](repo)
	loader := New(typedRepo)

	// 4. Load the node
	data, err := loader.GetNode("step_timeout")
	assert.NoError(t, err)

	// 5. Unmarshal to check domain object
	var node domain.Node
	err = json.Unmarshal(data, &node)
	assert.NoError(t, err)

	// 6. Verify Timeout mapping
	assert.Equal(t, "5s", node.Timeout)
}
