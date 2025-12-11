package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aretw0/loam/pkg/core"
)

// LoamLoader adapts the Loam library to the Trellis GraphLoader interface.
type LoamLoader struct {
	Service *core.Service
}

// NewLoamLoader creates a new Loam adapter.
func NewLoamLoader(service *core.Service) *LoamLoader {
	return &LoamLoader{
		Service: service,
	}
}

// GetNode retrieves a node from the Loam repository using the direct Service API.
// Note: Loam Service.GetDocument is a direct convenience lookup.
func (l *LoamLoader) GetNode(id string) ([]byte, error) {
	ctx := context.Background()

	docID := id + ".json"

	// Use GetDocument directly implies a read-only unit of work
	doc, err := l.Service.GetDocument(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("loam get failed for %s: %w", docID, err)
	}

	// Synthesize a complete JSON object from Metadata and Content
	// This allows the Trellis Compiler to parse it as a single unit,
	// while keeping storage separated in Loam.
	data := make(map[string]any)
	for k, v := range doc.Metadata {
		data[k] = v
	}

	// Treat content as bytes for JSON symmetry
	data["Content"] = []byte(doc.Content)

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node data: %w", err)
	}

	return bytes, nil
}
