package tests

import (
	"context"
	"testing"
	"time"

	"github.com/aretw0/trellis"
)

// MockWatchableLoader implements GraphLoader and Watchable
type MockWatchableLoader struct {
	watchCh chan string
}

func (m *MockWatchableLoader) GetNode(id string) ([]byte, error) {
	return nil, nil // Not needed for this test
}

func (m *MockWatchableLoader) ListNodes() ([]string, error) {
	return nil, nil
}

func (m *MockWatchableLoader) Watch(ctx context.Context) (<-chan string, error) {
	return m.watchCh, nil
}

// MockLoader implements GraphLoader but NOT Watchable
type MockLoader struct{}

func (m *MockLoader) GetNode(id string) ([]byte, error) { return nil, nil }
func (m *MockLoader) ListNodes() ([]string, error)      { return nil, nil }

func TestEngine_Watch_Success(t *testing.T) {
	mockLoader := &MockWatchableLoader{
		watchCh: make(chan string),
	}
	// Pre-fill channel to verify we receive it
	go func() {
		mockLoader.watchCh <- "reload"
	}()

	engine, err := trellis.New("", trellis.WithLoader(mockLoader))
	if err != nil {
		t.Fatalf("Failed to init engine: %v", err)
	}

	ctx := context.Background()
	ch, err := engine.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	select {
	case <-ch:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for watch event")
	}
}

func TestEngine_Watch_NotSupported(t *testing.T) {
	mockLoader := &MockLoader{}
	engine, err := trellis.New("", trellis.WithLoader(mockLoader))
	if err != nil {
		t.Fatalf("Failed to init engine: %v", err)
	}

	_, err = engine.Watch(context.Background())
	if err == nil {
		t.Fatal("Expected error when loader is not watchable, got nil")
	}
}
