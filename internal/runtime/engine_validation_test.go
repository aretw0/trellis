package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestEngine_RequiredContext(t *testing.T) {
	t.Run("Start Node Missing Context", func(t *testing.T) {
		startNode := domain.Node{
			ID:              "start",
			Type:            domain.NodeTypeStart,
			Content:         []byte("Start"),
			RequiredContext: []string{"api_key"},
		}
		loader, _ := inmemory.NewFromNodes(startNode)
		engine := runtime.NewEngine(loader, nil, nil)

		// Start() gets invalid state (no context)
		state := domain.NewState("start")

		// Render should fail
		_, _, err := engine.Render(context.Background(), state)
		assert.Error(t, err)
		var validationErr *runtime.ContextValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "start", validationErr.NodeID)
		assert.Contains(t, validationErr.MissingKeys, "api_key")
	})

	t.Run("Start Node With Context", func(t *testing.T) {
		startNode := domain.Node{
			ID:              "start",
			Type:            domain.NodeTypeStart,
			Content:         []byte("Start"),
			RequiredContext: []string{"api_key"},
		}
		loader, _ := inmemory.NewFromNodes(startNode)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("start")
		// Manually hydrate context to simulate pre-seeding (or CLI injection)
		state.Context["api_key"] = "secret"

		_, _, err := engine.Render(context.Background(), state)
		assert.NoError(t, err)
	})

	t.Run("Transition To Node Missing Context", func(t *testing.T) {
		startNode := domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeText,
			Content: []byte("Start"),
			Transitions: []domain.Transition{
				{ToNodeID: "secure"},
			},
		}
		secureNode := domain.Node{
			ID:              "secure",
			Type:            domain.NodeTypeText,
			Content:         []byte("Secure"),
			RequiredContext: []string{"token"},
		}

		loader, _ := inmemory.NewFromNodes(startNode, secureNode)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("start")

		// Transition to Secure
		nextState, err := engine.Navigate(context.Background(), state, nil)
		assert.NoError(t, err)
		assert.Equal(t, "secure", nextState.CurrentNodeID)

		// Render Secure -> Should Fail
		_, _, err = engine.Render(context.Background(), nextState)
		assert.Error(t, err)
		var validationErr *runtime.ContextValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "secure", validationErr.NodeID)
		assert.Contains(t, validationErr.MissingKeys, "token")
	})
}
