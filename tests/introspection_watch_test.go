package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

// TestIntrospection_Watch validates the Watch() streaming functionality
func TestIntrospection_Watch(t *testing.T) {
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:          "start",
			Type:        domain.NodeTypeText,
			Transitions: []domain.Transition{{ToNodeID: "step2"}},
			Content:     []byte("Start"),
		},
		domain.Node{
			ID:          "step2",
			Type:        domain.NodeTypeText,
			Wait:        true,
			Transitions: []domain.Transition{{ToNodeID: "end"}},
			Content:     []byte("Step 2"),
		},
		domain.Node{
			ID:      "end",
			Type:    domain.NodeTypeText,
			Content: []byte("End"),
		},
	)
	require.NoError(t, err)

	eng, err := trellis.New("", trellis.WithLoader(loader))
	require.NoError(t, err)

	inputCh := make(chan string)
	handler := &ChannelHandler{InputCh: inputCh}

	r := runner.NewRunner(
		runner.WithEngine(eng),
		runner.WithInputHandler(handler),
	)

	// Setup watcher BEFORE starting execution
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	changes := r.Watch(ctx)

	// Track received changes
	var receivedChanges []string
	var mu sync.Mutex

	go func() {
		for change := range changes {
			mu.Lock()
			receivedChanges = append(receivedChanges, change.NewState.CurrentNodeID)
			mu.Unlock()
		}
	}()

	// Start runner
	errCh := make(chan error)
	go func() {
		errCh <- r.Run(ctx)
	}()

	// Wait for step2
	time.Sleep(100 * time.Millisecond)

	// Send input to proceed
	select {
	case inputCh <- "next":
	case <-ctx.Done():
		t.Fatal("Context timeout while sending input")
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	changeCount := len(receivedChanges)
	mu.Unlock()

	// Should have received at least 2 state changes (start->step2, step2->end)
	assert.GreaterOrEqual(t, changeCount, 2, "Should receive multiple state changes")

	// Verify watcher channel closes on context cancel
	cancel()
	time.Sleep(50 * time.Millisecond)

	// Channel should be closed
	_, ok := <-changes
	assert.False(t, ok, "Watcher channel should be closed after context cancel")
}

// TestIntrospection_MultipleWatchers validates concurrent watchers
func TestIntrospection_MultipleWatchers(t *testing.T) {
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:          "start",
			Type:        domain.NodeTypeText,
			Transitions: []domain.Transition{{ToNodeID: "end"}},
			Content:     []byte("Start"),
		},
		domain.Node{
			ID:      "end",
			Type:    domain.NodeTypeText,
			Content: []byte("End"),
		},
	)
	require.NoError(t, err)

	eng, err := trellis.New("", trellis.WithLoader(loader))
	require.NoError(t, err)

	inputCh := make(chan string, 10)
	handler := &ChannelHandler{InputCh: inputCh}

	r := runner.NewRunner(
		runner.WithEngine(eng),
		runner.WithInputHandler(handler),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create 3 concurrent watchers
	const watcherCount = 3
	var wg sync.WaitGroup
	wg.Add(watcherCount)

	receivedCounts := make([]int, watcherCount)
	var mu sync.Mutex

	for i := 0; i < watcherCount; i++ {
		idx := i
		changes := r.Watch(ctx)

		go func() {
			defer wg.Done()
			count := 0
			for range changes {
				count++
			}
			mu.Lock()
			receivedCounts[idx] = count
			mu.Unlock()
		}()
	}

	// Start runner
	errCh := make(chan error)
	go func() {
		errCh <- r.Run(ctx)
	}()

	// Wait for execution
	select {
	case <-errCh:
	case <-ctx.Done():
	}

	cancel()
	wg.Wait()

	// All watchers should have received events
	for i, count := range receivedCounts {
		assert.Greater(t, count, 0, "Watcher %d should receive events", i)
	}
}

// TestIntrospection_WatcherCancellation validates graceful cleanup
func TestIntrospection_WatcherCancellation(t *testing.T) {
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:          "start",
			Type:        domain.NodeTypeText,
			Wait:        true,
			Transitions: []domain.Transition{{ToNodeID: "end"}},
			Content:     []byte("Waiting"),
		},
		domain.Node{
			ID:      "end",
			Type:    domain.NodeTypeText,
			Content: []byte("End"),
		},
	)
	require.NoError(t, err)

	eng, err := trellis.New("", trellis.WithLoader(loader))
	require.NoError(t, err)

	inputCh := make(chan string, 10)
	handler := &ChannelHandler{InputCh: inputCh}

	r := runner.NewRunner(
		runner.WithEngine(eng),
		runner.WithInputHandler(handler),
	)

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	// Start runner
	go func() {
		_ = r.Run(runCtx)
	}()

	// Create watcher with independent context
	watchCtx, watchCancel := context.WithCancel(context.Background())
	changes := r.Watch(watchCtx)

	// Wait for initial state
	time.Sleep(100 * time.Millisecond)

	// Cancel watcher context (runner still running)
	watchCancel()

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Drain any remaining messages and verify channel eventually closes
	closed := false
	timeout := time.After(500 * time.Millisecond)
	for !closed {
		select {
		case _, ok := <-changes:
			if !ok {
				closed = true
			}
		case <-timeout:
			t.Fatal("Watcher channel did not close within timeout")
		}
	}
	assert.True(t, closed, "Watcher should close on context cancel")

	// Runner should still be operational
	state := r.State()
	assert.NotNil(t, state)
	assert.Equal(t, "start", state.CurrentNodeID)

	// Cleanup
	runCancel()
}

// TestIntrospection_SlowConsumer validates backpressure handling
func TestIntrospection_SlowConsumer(t *testing.T) {
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:          "start",
			Type:        domain.NodeTypeText,
			Transitions: []domain.Transition{{ToNodeID: "step2"}},
			Content:     []byte("Start"),
		},
		domain.Node{
			ID:          "step2",
			Type:        domain.NodeTypeText,
			Transitions: []domain.Transition{{ToNodeID: "step3"}},
			Content:     []byte("Step 2"),
		},
		domain.Node{
			ID:          "step3",
			Type:        domain.NodeTypeText,
			Transitions: []domain.Transition{{ToNodeID: "end"}},
			Content:     []byte("Step 3"),
		},
		domain.Node{
			ID:      "end",
			Type:    domain.NodeTypeText,
			Content: []byte("End"),
		},
	)
	require.NoError(t, err)

	eng, err := trellis.New("", trellis.WithLoader(loader))
	require.NoError(t, err)

	inputCh := make(chan string, 10)
	handler := &ChannelHandler{InputCh: inputCh}

	r := runner.NewRunner(
		runner.WithEngine(eng),
		runner.WithInputHandler(handler),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	changes := r.Watch(ctx)

	// Slow consumer - intentionally delay reading
	receivedCount := 0
	go func() {
		for range changes {
			time.Sleep(200 * time.Millisecond) // Slow consumer
			receivedCount++
		}
	}()

	// Start runner (should not block even with slow consumer)
	errCh := make(chan error)
	go func() {
		errCh <- r.Run(ctx)
	}()

	// Wait for completion
	select {
	case err := <-errCh:
		// Runner should complete successfully despite slow consumer
		assert.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("Runner should complete before timeout")
	}

	// Note: receivedCount may be less than total transitions
	// due to non-blocking send (drops are acceptable)
}
