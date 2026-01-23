package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestEngine_Signal(t *testing.T) {
	// Define a flow with an interrupt handler
	/*
		start:
		  on_signal:
			interrupt: shutdown
		  transitions:
			- to: next

		next:
		  type: text

		shutdown:
		  type: text
		  content: "Shutting down..."
	*/
	startNode := domain.Node{
		ID:   "start",
		Type: domain.NodeTypeText,
		OnSignal: map[string]string{
			domain.SignalInterrupt: "shutdown",
		},
		Transitions: []domain.Transition{
			{ToNodeID: "next"},
		},
	}
	nextNode := domain.Node{
		ID:      "next",
		Type:    domain.NodeTypeText,
		Content: []byte("Next Step"),
	}
	shutdownNode := domain.Node{
		ID:      "shutdown",
		Type:    domain.NodeTypeText,
		Content: []byte("Shutting down..."),
	}

	loader, _ := inmemory.NewFromNodes(startNode, nextNode, shutdownNode)
	engine := runtime.NewEngine(loader, nil, nil)

	t.Run("Successfully Handles Interrupt Signal", func(t *testing.T) {
		// Start
		state, _ := engine.Start(context.Background(), "signal-test", nil)
		assert.Equal(t, "start", state.CurrentNodeID)

		// Send Signal
		nextState, err := engine.Signal(context.Background(), state, domain.SignalInterrupt)
		assert.NoError(t, err)
		assert.Equal(t, "shutdown", nextState.CurrentNodeID)
	})

	t.Run("Returns Error For Unhandled Signal", func(t *testing.T) {
		// Start
		state, _ := engine.Start(context.Background(), "signal-test", nil)

		// Send Unknown Signal
		_, err := engine.Signal(context.Background(), state, "unknown_signal")
		assert.ErrorIs(t, err, domain.ErrUnhandledSignal)
	})

	t.Run("Ignores Signal If Not Configured", func(t *testing.T) {
		// Move to 'next' node which has no handlers
		state := domain.NewState("signal-test", "next")

		// Send Signal
		_, err := engine.Signal(context.Background(), state, domain.SignalInterrupt)
		assert.ErrorIs(t, err, domain.ErrUnhandledSignal)
	})
}
