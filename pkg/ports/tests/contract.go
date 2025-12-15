package tests

import (
	"testing"

	"github.com/aretw0/trellis/pkg/ports"
)

// GraphLoaderContractTest is a reusable test suite that verifies if an adapter complies with ports.GraphLoader.
func GraphLoaderContractTest(t *testing.T, loader ports.GraphLoader, setupData map[string][]byte) {
	t.Helper()

	// 1. Test GetNode (Success)
	t.Run("GetNode_Success", func(t *testing.T) {
		for id, expectedContent := range setupData {
			content, err := loader.GetNode(id)
			if err != nil {
				t.Fatalf("unexpected error getting node %s: %v", id, err)
			}
			if string(content) != string(expectedContent) {
				t.Errorf("content mismatch for %s. got %q, want %q", id, content, expectedContent)
			}
		}
	})

	// 2. Test GetNode (NotFound)
	t.Run("GetNode_NotFound", func(t *testing.T) {
		_, err := loader.GetNode("non-existent-node")
		if err == nil {
			t.Error("expected error for non-existent node, got nil")
		}
	})

	// 3. Test ListNodes
	t.Run("ListNodes", func(t *testing.T) {
		nodes, err := loader.ListNodes()
		if err != nil {
			t.Fatalf("unexpected error listing nodes: %v", err)
		}

		if len(nodes) != len(setupData) {
			t.Errorf("expected %d nodes, got %d", len(setupData), len(nodes))
		}

		// Verify all expected IDs are present
		lookup := make(map[string]bool)
		for _, id := range nodes {
			lookup[id] = true
		}

		for id := range setupData {
			if !lookup[id] {
				t.Errorf("node %s missing from list", id)
			}
		}
	})
}
