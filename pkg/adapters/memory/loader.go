package memory

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aretw0/trellis/pkg/domain"
)

// Loader implements ports.GraphLoader using an in-memory map.
type Loader struct {
	nodes map[string][]byte
}

// NewLoader creates a new MemoryLoader with the provided raw data (JSON strings).
func NewLoader(data map[string]string) *Loader {
	nodes := make(map[string][]byte)
	for k, v := range data {
		nodes[k] = []byte(v)
	}
	return &Loader{
		nodes: nodes,
	}
}

// NewFromNodes creates a new MemoryLoader from domain objects.
// This handles serialization automatically, improving DX for tests.
func NewFromNodes(nodes ...domain.Node) (*Loader, error) {
	data := make(map[string][]byte)
	for _, n := range nodes {
		if n.ID == "" {
			return nil, fmt.Errorf("node missing ID")
		}
		bytes, err := json.Marshal(n)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal node %s: %w", n.ID, err)
		}
		data[n.ID] = bytes
	}
	return &Loader{nodes: data}, nil
}

// GetNode retrieves the raw definition of a node by ID.
func (l *Loader) GetNode(id string) ([]byte, error) {
	content, ok := l.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", id)
	}
	return content, nil
}

// ListNodes returns all available node IDs.
func (l *Loader) ListNodes() ([]string, error) {
	keys := make([]string, 0, len(l.nodes))
	for k := range l.nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Deterministic order
	return keys, nil
}
