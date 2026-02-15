package tests

import (
	"context"
	"testing"
	"time"

	"github.com/aretw0/introspection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

func TestIntrospection_RunnerState(t *testing.T) {
	// Define a graph with a wait state
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:          "start",
			Type:        domain.NodeTypeText,
			Transitions: []domain.Transition{{ToNodeID: "wait_node"}},
			Content:     []byte("Start"),
		},
		domain.Node{
			ID:          "wait_node",
			Type:        domain.NodeTypeText,
			Wait:        true, // Requires input, so it will pause here
			Transitions: []domain.Transition{{ToNodeID: "end"}},
			Content:     []byte("Waiting"),
		},
		domain.Node{
			ID:      "end",
			Type:    domain.NodeTypeText,
			Content: []byte("Done"),
		},
	)
	require.NoError(t, err)

	eng, err := trellis.New("", trellis.WithLoader(loader))
	require.NoError(t, err)

	// Use a channel handler to control flow
	inputCh := make(chan string)
	handler := &ChannelHandler{InputCh: inputCh}

	r := runner.NewRunner(
		runner.WithEngine(eng),
		runner.WithInputHandler(handler),
	)

	// 1. Verify TypedWatcher interface compliance
	var watcher introspection.TypedWatcher[*domain.State] = r
	_ = watcher

	// 2. Start Runner in background
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- r.Run(ctx)
	}()

	// 3. Poll for "wait_node" state
	// Loop until the runner reaches the wait node or timeout
	stateReached := false
	for range 10 {
		state := r.State()
		// State() might return nil or initial state at first
		if state != nil && state.CurrentNodeID == "wait_node" {
			stateReached = true
			// Status usually stays active while waiting for input in handler
			assert.Equal(t, domain.StatusActive, state.Status)
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.True(t, stateReached, "Runner should have reached 'wait_node'")

	// 4. Verify snapshot isolation
	snap1 := r.State()
	assert.NotNil(t, snap1)

	// Send input to proceed
	select {
	case inputCh <- "next":
	case <-ctx.Done():
		t.Fatal("Context timeout while sending input")
	}

	// 5. Poll for "end" state
	endReached := false
	for i := 0; i < 10; i++ {
		state := r.State()
		if state != nil && state.CurrentNodeID == "end" {
			endReached = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.True(t, endReached, "Runner should have reached 'end'")

	// 6. Verify snap1 didn't change (immutability check)
	assert.Equal(t, "wait_node", snap1.CurrentNodeID, "Snapshot should preserve old state")

	// Wait for runner to finish
	err = <-errCh
	assert.NoError(t, err)
}
