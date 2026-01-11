package registry

import (
	"context"
	"fmt"
	"sync"
)

// ToolFunction defines the signature for a tool implementation.
// It receives a context and a map of arguments, and returns a result or error.
type ToolFunction func(ctx context.Context, args map[string]any) (any, error)

// Registry manages the available tools.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]ToolFunction
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]ToolFunction),
	}
}

// Register adds a tool to the registry.
// If a tool with the same name exists, it is overwritten.
func (r *Registry) Register(name string, fn ToolFunction) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[name] = fn
}

// Execute looks up a tool by name and executes it.
// Returns an error if the tool is not found.
func (r *Registry) Execute(ctx context.Context, name string, args map[string]any) (any, error) {
	r.mu.RLock()
	fn, ok := r.tools[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return fn(ctx, args)
}
