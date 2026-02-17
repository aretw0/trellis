package dsl

import (
	"fmt"

	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

// Builder manages the graph construction.
type Builder struct {
	nodes map[string]*NodeBuilder
}

// New creates a new graph builder.
func New() *Builder {
	return &Builder{
		nodes: make(map[string]*NodeBuilder),
	}
}

// Add creates a new node in the graph.
// If the node already exists, it returns the existing builder.
func (b *Builder) Add(id string) *NodeBuilder {
	if nb, ok := b.nodes[id]; ok {
		return nb
	}
	nb := &NodeBuilder{
		node: domain.Node{
			ID: id,
		},
		builder: b,
	}
	b.nodes[id] = nb
	return nb
}

// Build compiles the graph into a MemoryLoader.
func (b *Builder) Build() (*memory.Loader, error) {
	nodes := make([]domain.Node, 0, len(b.nodes))
	for _, nb := range b.nodes {
		nodes = append(nodes, nb.node)
	}

	loader, err := memory.NewFromNodes(nodes...)
	if err != nil {
		return nil, fmt.Errorf("failed to build memory loader: %w", err)
	}

	return loader, nil
}
