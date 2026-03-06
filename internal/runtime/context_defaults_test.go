package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_Start_WithDefaultContext(t *testing.T) {
	startNode := domain.Node{
		ID:             "start",
		Type:           domain.NodeTypeStart,
		Content:        []byte("Start Node"),
		DefaultContext: map[string]any{"env": "dev", "retries": 3},
	}
	loader, _ := memory.NewFromNodes(startNode)
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

func TestEngine_DefaultContext_E2E(t *testing.T) {
	node1 := domain.Node{
		ID:             "start",
		Type:           domain.NodeTypeQuestion,
		Content:        []byte("Hello {{ .username }}! Role: {{ default \"guest\" .role }}. Next?"),
		DefaultContext: map[string]any{"username": "Alice", "role": "admin"},
		Transitions: []domain.Transition{
			{ToNodeID: "step2"},
		},
	}

	node2 := domain.Node{
		ID:             "step2",
		Type:           domain.NodeTypeText,
		Content:        []byte("Step 2. Status: {{ .status }}. Username: {{ .username }}"),
		DefaultContext: map[string]any{"status": "active"},
	}

	loader, err := memory.NewFromNodes(node1, node2)
	require.NoError(t, err)

	engine := runtime.NewEngine(loader, nil, nil)
	ctx := context.Background()

	// 1. Start Engine
	state, err := engine.Start(ctx, "session-1", map[string]any{})
	require.NoError(t, err)

	assert.Equal(t, "Alice", state.Context["username"])
	assert.Equal(t, "admin", state.Context["role"])

	// 2. Render start node
	actions, _, err := engine.Render(ctx, state)
	require.NoError(t, err)

	var renderedText string
	for _, act := range actions {
		if act.Type == domain.ActionRenderContent {
			renderedText = act.Payload.(string)
		}
	}
	assert.Equal(t, "Hello Alice! Role: admin. Next?", renderedText)

	// 3. Navigate to step2 (Username Alice should persist, Status active should be applied)
	nextState, err := engine.Navigate(ctx, state, "go")
	require.NoError(t, err)

	// 4. Render step2
	actions2, _, err := engine.Render(ctx, nextState)
	require.NoError(t, err)

	var renderedText2 string
	for _, act := range actions2 {
		if act.Type == domain.ActionRenderContent {
			renderedText2 = act.Payload.(string)
		}
	}
	assert.Equal(t, "Step 2. Status: active. Username: Alice", renderedText2)
}
