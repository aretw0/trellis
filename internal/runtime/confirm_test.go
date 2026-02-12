package runtime

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfirmInputTypeWithOnDenied verifies that confirm type correctly handles on_denied
func TestConfirmInputTypeWithOnDenied(t *testing.T) {
	loader, _ := memory.NewFromNodes(
		domain.Node{
			ID:        "confirm_exit",
			Type:      domain.NodeTypeQuestion,
			InputType: "confirm",
			OnDenied:  "start",
			Transitions: []domain.Transition{
				{ToNodeID: "exit_now"},
			},
			Content: []byte("Are you sure?"),
		},
		domain.Node{
			ID:      "exit_now",
			Type:    domain.NodeTypeText,
			Content: []byte("Goodbye!"),
		},
		domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeText,
			Content: []byte("Welcome!"),
			Transitions: []domain.Transition{
				{ToNodeID: "confirm_exit"},
			},
		},
	)

	engine := NewEngine(loader, nil, nil)

	t.Run("yes input should go to exit_now", func(t *testing.T) {
		state := domain.NewState("test", "confirm_exit")

		nextState, err := engine.Navigate(context.Background(), state, "yes")
		require.NoError(t, err)
		assert.Equal(t, "exit_now", nextState.CurrentNodeID)
	})

	t.Run("no input should go to start via on_denied", func(t *testing.T) {
		state := domain.NewState("test", "confirm_exit")

		nextState, err := engine.Navigate(context.Background(), state, "no")
		require.NoError(t, err)
		assert.Equal(t, "start", nextState.CurrentNodeID,
			"Expected on_denied to route 'no' to start, but got %s", nextState.CurrentNodeID)
	})

	t.Run("n input should also trigger on_denied", func(t *testing.T) {
		state := domain.NewState("test", "confirm_exit")

		nextState, err := engine.Navigate(context.Background(), state, "n")
		require.NoError(t, err)
		assert.Equal(t, "start", nextState.CurrentNodeID)
	})

	t.Run("empty input should default to yes and go to exit_now", func(t *testing.T) {
		state := domain.NewState("test", "confirm_exit")

		nextState, err := engine.Navigate(context.Background(), state, "")
		require.NoError(t, err)
		assert.Equal(t, "exit_now", nextState.CurrentNodeID)
	})
}
