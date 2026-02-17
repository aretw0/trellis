package runtime_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/schema"
)

func TestEngine_RequiredContext(t *testing.T) {
	t.Run("Start Node Missing Context", func(t *testing.T) {
		startNode := domain.Node{
			ID:              "start",
			Type:            domain.NodeTypeStart,
			Content:         []byte("Start"),
			RequiredContext: []string{"api_key"},
		}
		loader, _ := memory.NewFromNodes(startNode)
		engine := runtime.NewEngine(loader, nil, nil)

		// Start() gets invalid state (no context)
		state := domain.NewState("test-session", "start")

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
		loader, _ := memory.NewFromNodes(startNode)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("test-session", "start")
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

		loader, _ := memory.NewFromNodes(startNode, secureNode)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("test-session", "start")

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

func TestEngine_ContextSchema(t *testing.T) {
	t.Run("Start Node ContextSchema Valid", func(t *testing.T) {
		startNode := domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeStart,
			Content: []byte("Start"),
			ContextSchema: schema.Schema{
				"api_key": schema.String(),
				"retries": schema.Int(),
			},
		}
		loader, _ := memory.NewFromNodes(startNode)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("test-session", "start")
		state.Context["api_key"] = "secret"
		state.Context["retries"] = 3

		_, _, err := engine.Render(context.Background(), state)
		assert.NoError(t, err)
	})

	t.Run("Start Node ContextSchema Missing Field", func(t *testing.T) {
		startNode := domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeStart,
			Content: []byte("Start"),
			ContextSchema: schema.Schema{
				"api_key": schema.String(),
				"retries": schema.Int(),
			},
		}
		loader, _ := memory.NewFromNodes(startNode)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("test-session", "start")
		state.Context["api_key"] = "secret"

		_, _, err := engine.Render(context.Background(), state)
		assert.Error(t, err)
		var validationErr *runtime.ContextTypeValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "start", validationErr.NodeID)
	})

	t.Run("Start Node ContextSchema Invalid Type", func(t *testing.T) {
		startNode := domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeStart,
			Content: []byte("Start"),
			ContextSchema: schema.Schema{
				"api_key": schema.String(),
				"retries": schema.Int(),
			},
		}
		loader, _ := memory.NewFromNodes(startNode)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("test-session", "start")
		state.Context["api_key"] = "secret"
		state.Context["retries"] = "two"

		_, _, err := engine.Render(context.Background(), state)
		assert.Error(t, err)
		var validationErr *runtime.ContextTypeValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "start", validationErr.NodeID)
	})
}
