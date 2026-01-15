package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSignalManager_Lifecycle(t *testing.T) {
	sm := NewSignalManager()
	defer sm.Stop()

	// 1. Initial State
	ctx1 := sm.Context()
	assert.NotNil(t, ctx1)
	assert.NoError(t, ctx1.Err())

	// 2. Reset (should create new context)
	sm.Reset()
	ctx2 := sm.Context()
	assert.NotNil(t, ctx2)
	assert.NotEqual(t, ctx1, ctx2, "Reset should generate a new context")
	assert.NoError(t, ctx2.Err())

	// 3. Stop (should cancel context)
	sm.Stop()
	assert.ErrorIs(t, ctx2.Err(), context.Canceled)
}

func TestSignalManager_CheckRace(t *testing.T) {
	// Logic check: CheckRace should not block indefinitely if context is not cancelled.
	// It waits 100ms. We verify it returns within reasonable time.
	sm := NewSignalManager()
	defer sm.Stop()

	start := time.Now()
	sm.CheckRace() // Should timeout after 100ms
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
	assert.Less(t, elapsed, 200*time.Millisecond, "CheckRace took too long")
}
