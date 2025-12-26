package ports

import "context"

// GraphLoader defines how the engine retrieves node definitions.
// This allows the storage layer (Loam, FS, Memory) to be decoupled.
type GraphLoader interface {
	// GetNode retrieves the raw definition of a node by ID.
	// It returns the raw bytes (which the compiler will parse) or an error.
	GetNode(id string) ([]byte, error)

	// ListNodes returns a simplified list of all node IDs available in the graph.
	// This is used for introspection and visualization tools (e.g. 'trellis graph').
	ListNodes() ([]string, error)
}

// Watchable defines an interface for loaders that can notify about backend changes.
// This is typically used for hot-reload or dev-mode functionality.
type Watchable interface {
	// Watch returns a channel that is signaled when the underlying graph changes.
	// It abstracts away the specific event details, signaling only that a reload is required.
	Watch(ctx context.Context) (<-chan struct{}, error)
}
