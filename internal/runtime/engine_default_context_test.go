package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestEngine_Start_WithDefaultContext(t *testing.T) {
	startNode := domain.Node{
		ID:             "start",
		Type:           domain.NodeTypeStart,
		Content:        []byte("Start Node"),
		DefaultContext: map[string]any{"env": "dev", "retries": 3},
	}
	loader, _ := inmemory.NewFromNodes(startNode)
	engine := runtime.NewEngine(loader, nil, nil)

	t.Run("Applies Defaults When No Context Provided", func(t *testing.T) {
		state, err := engine.Start(context.Background(), "test-session", nil)
		assert.NoError(t, err)
		assert.Equal(t, "dev", state.Context["env"])
		assert.EqualValues(t, 3, state.Context["retries"])
	})

	t.Run("Overrides Defaults With Initial Context", func(t *testing.T) {
		initial := map[string]any{
			"env": "prod",
		}
		state, err := engine.Start(context.Background(), "test-session", initial)
		assert.NoError(t, err)
		assert.Equal(t, "prod", state.Context["env"])
		assert.EqualValues(t, 3, state.Context["retries"]) // Preserves non-overridden default
	})
}
