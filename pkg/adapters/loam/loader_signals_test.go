package loam

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/testutils"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_SyntacticSugar_Signals(t *testing.T) {
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Create nodes with syntax sugar
	files := map[string]string{
		"timeout.md": `---
type: text
on_timeout: my_timeout_handler
---
Timeout Test`,
		"interrupt.md": `---
type: text
on_interrupt: my_interrupt_handler
---
Interrupt Test`,
		"conflict.md": `---
type: text
on_timeout: sugar_handler
on_signal:
  timeout: explicit_handler
---
Conflict Test (Sugar should typically win or merge)`,
	}

	for filename, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	typedRepo := loam.NewTypedRepository[NodeMetadata](repo)
	loader := New(typedRepo)

	t.Run("on_timeout mapping", func(t *testing.T) {
		data, err := loader.GetNode("timeout")
		require.NoError(t, err)
		jsonStr := string(data)
		assert.Contains(t, jsonStr, `"`+domain.SignalTimeout+`":"my_timeout_handler"`)
	})

	t.Run("on_interrupt mapping", func(t *testing.T) {
		data, err := loader.GetNode("interrupt")
		require.NoError(t, err)
		jsonStr := string(data)
		assert.Contains(t, jsonStr, `"`+domain.SignalInterrupt+`":"my_interrupt_handler"`)
	})

	t.Run("sugar precedence", func(t *testing.T) {
		// Our implementation applies sugar AFTER reading on_signal,
		// so sugar overwrites explicit key in the map.
		data, err := loader.GetNode("conflict")
		require.NoError(t, err)
		jsonStr := string(data)
		assert.Contains(t, jsonStr, `"`+domain.SignalTimeout+`":"sugar_handler"`)
	})
}
