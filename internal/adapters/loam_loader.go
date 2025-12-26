package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/dto"
	"github.com/aretw0/trellis/pkg/domain"
)

// LoamLoader adapts the Loam library to the Trellis GraphLoader interface.
type LoamLoader struct {
	Repo *loam.TypedRepository[dto.NodeMetadata]
}

// NewLoamLoader creates a new Loam adapter.
func NewLoamLoader(repo *loam.TypedRepository[dto.NodeMetadata]) *LoamLoader {
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
	// We must map our loader struct (which matches YAML "to") to domain JSON ("to_node_id").

	domainTransitions := make([]domain.Transition, len(doc.Data.Transitions))
	for i, lt := range doc.Data.Transitions {
		// Support both "to" and "to_node_id"
		to := lt.To
		if to == "" {
			to = lt.ToFull
		}
		from := lt.From
		if from == "" {
			from = lt.FromFull
		}

		domainTransitions[i] = domain.Transition{
			FromNodeID: from,
			ToNodeID:   to,
			Condition:  lt.Condition,
		}
	}

	data := make(map[string]any)
	// Normalize ID: prefer metadata ID, fallback to filename ID, but always strip extension
	rawID := doc.Data.ID
	if rawID == "" {
		rawID = doc.ID
	}
	data["id"] = trimExtension(rawID)

	data["type"] = doc.Data.Type
	data["transitions"] = domainTransitions
	data["content"] = []byte(doc.Content) // As Base64

	// Map Input Configuration
	if doc.Data.InputType != "" {
		data["input_type"] = doc.Data.InputType
		data["input_options"] = doc.Data.InputOptions
		data["input_default"] = doc.Data.InputDefault
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node data: %w", err)
	}

	return bytes, nil
}

// ListNodes lists all nodes in the repository.
func (l *LoamLoader) ListNodes() ([]string, error) {
	ctx := context.Background()
	docs, err := l.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("loam list failed: %w", err)
	}

	ids := make([]string, len(docs))
	for i, doc := range docs {
		// Use the ID from metadata if available, otherwise filename ID
		rawID := doc.Data.ID
		if rawID == "" {
			rawID = doc.ID
		}
		ids[i] = trimExtension(rawID)
	}
	return ids, nil
}

func trimExtension(id string) string {
	ext := filepath.Ext(id)
	if ext != "" {
		return strings.TrimSuffix(id, ext)
	}
	return id
}

// Watch returns a channel that is signaled when key files change in the Loam repository.
func (l *LoamLoader) Watch(ctx context.Context) (<-chan struct{}, error) {
	// Watch for all relevant files (recursive) using doublestar pattern supported by Loam/Doublestar
	// This avoids manual filtering loop.
	events, err := l.Repo.Watch(ctx, "**/*.{md,json,yaml,yml}")
	if err != nil {
		return nil, fmt.Errorf("failed to start loam watcher: %w", err)
	}

	reloadCh := make(chan struct{}, 1)

	go func() {
		defer close(reloadCh)
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-events:
				if !ok {
					return
				}
				// Debounce logic could be here, but for now we just notify.
				// Non-blocking send to avoid stalling if receiver isn't ready
				select {
				case reloadCh <- struct{}{}:
				default:
				}
			}
		}
	}()

	return reloadCh, nil
}
