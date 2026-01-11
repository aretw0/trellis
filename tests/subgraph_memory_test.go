package tests

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/require"
)

// TestEngine_Subgraphs verifies that the Engine handles namespaced IDs (subgraphs) correctly
// without relying on Loam or file systems. This proves the core is agnostic to ID formatting.
func TestEngine_Subgraphs(t *testing.T) {
	// Define nodes with directory-style IDs
	rootID := "start"
	subID := "modules/checkout/start"
	endID := "modules/checkout/end"

	rootNode := domain.Node{
		ID:      rootID,
		Type:    domain.NodeTypeText,
		Content: []byte("Root"),
		Transitions: []domain.Transition{
			{ToNodeID: subID, Condition: ""},
		},
	}

	subNode := domain.Node{
		ID:      subID,
		Type:    domain.NodeTypeText,
		Content: []byte("Checkout Subgraph"),
		Transitions: []domain.Transition{
			{ToNodeID: endID, Condition: ""},
		},
	}

	endNode := domain.Node{
		ID:      endID,
		Type:    domain.NodeTypeText,
		Content: []byte("End of Checkout"),
		Transitions: []domain.Transition{
			{ToNodeID: rootID, Condition: ""}, // Loop back
		},
	}

	// Load into inmemoryLoader (Pre-compiled Domain Nodes)
	loader, err := inmemory.NewFromNodes(rootNode, subNode, endNode)
	require.NoError(t, err)

	engine := runtime.NewEngine(loader, nil)

	// 1. Start at Root
	state := domain.NewState(rootID)

	// 2. Navigate to Subgraph
	// Note: For Text nodes, Navigate("") with empty input triggers default transition
	nextState, err := engine.Navigate(context.Background(), state, "")
	require.NoError(t, err)
	require.Equal(t, subID, nextState.CurrentNodeID, "Should navigate to namespaced ID")

	// 3. Navigate deeper in Subgraph
	nextState, err = engine.Navigate(context.Background(), nextState, "")
	require.NoError(t, err)
	require.Equal(t, endID, nextState.CurrentNodeID)

	// 4. Return to Root
	nextState, err = engine.Navigate(context.Background(), nextState, "")
	require.NoError(t, err)
	require.Equal(t, rootID, nextState.CurrentNodeID)
}
