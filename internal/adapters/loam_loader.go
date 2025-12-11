package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/pkg/domain"
)

// NodeMetadata represents the header/metadata of a Trellis Node.
// It matches the domain.Node fields minus Content.
type NodeMetadata struct {
	ID          string              `json:"id"`
	Type        string              `json:"type"`
	Transitions []domain.Transition `json:"transitions"`
}

// LoamLoader adapts the Loam library to the Trellis GraphLoader interface.
type LoamLoader struct {
	Repo *loam.TypedRepository[NodeMetadata]
}

// NewLoamLoader creates a new Loam adapter.
func NewLoamLoader(repo *loam.TypedRepository[NodeMetadata]) *LoamLoader {
	return &LoamLoader{
		Repo: repo,
	}
}

// GetNode retrieves a node from the Loam repository using the direct Service API.
// Note: Loam Service.GetDocument is a direct convenience lookup.
func (l *LoamLoader) GetNode(id string) ([]byte, error) {
	ctx := context.Background()

	// Loam Normalized Retrieval.
	// We trust Loam to find the file (e.g. start.md) even if we ask for "start",
	// or we assume the seeding created "start" (which maps to start.md).
	doc, err := l.Repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loam get failed for %s: %w", id, err)
	}

	// Synthesize valid JSON for the compiler.
	// Merge Metadata (doc.Data) + Content (doc.Content).

	// Create a map to avoid "double json" issues if we just marshaled struct.
	// Actually, we can just marshal a temporary struct that mimics domain.Node logic?
	// No, compiler expects JSON bytes.
	// We already have `NodeMetadata` struct populated.
	// We just need to add Content to it and Marshal.

	// We can't just set Content on NodeMetadata because it doesn't have it.
	// We can define a helper or just use map.

	data := make(map[string]any)
	data["id"] = doc.Data.ID // Or doc.ID (filename)? Stick to metadata ID.
	data["type"] = doc.Data.Type
	data["transitions"] = doc.Data.Transitions
	data["content"] = []byte(doc.Content) // As Base64

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node data: %w", err)
	}

	return bytes, nil
}
