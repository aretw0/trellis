package adapters

import (
	"fmt"
)

// InMemoryLoader is a simple implementation of ports.GraphLoader for testing.
type InMemoryLoader struct {
	nodes map[string][]byte
}

// NewInMemoryLoader creates a new empty loader.
func NewInMemoryLoader() *InMemoryLoader {
	return &InMemoryLoader{
		nodes: make(map[string][]byte),
	}
}

// AddNode allows pre-populating the loader for tests.
func (l *InMemoryLoader) AddNode(id string, data []byte) {
	l.nodes[id] = data
}

// GetNode retrieves a node from memory.
func (l *InMemoryLoader) GetNode(id string) ([]byte, error) {
	data, ok := l.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", id)
	}
	return data, nil
}
