package validator

import (
	"context"
	"fmt"
	"strings"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/trellis/internal/adapters"
)

// ValidateGraph checks for broken links and unreachable nodes starting from startNodeID.
func ValidateGraph(repo core.Repository, startNodeID string) error {
	// We use the TypedRepository to easily parse metadata
	typedRepo := loam.NewTypedRepository[adapters.NodeMetadata](repo)
	ctx := context.Background()

	// 1. Load Start Node
	startDoc, err := typedRepo.Get(ctx, startNodeID)
	if err != nil {
		return fmt.Errorf("start node '%s' not found: %w", startNodeID, err)
	}

	// 2. Crawler
	visited := make(map[string]bool)
	// We need to resolve the actual ID if "start" maps to "start.md"
	actualStartID := startDoc.Data.ID
	if actualStartID == "" {
		actualStartID = startNodeID
	}

	queue := []string{actualStartID}

	// fmt.Printf("Starting validation from '%s'...\n", actualStartID)

	var errors []string

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		// Load Node
		doc, err := typedRepo.Get(ctx, currentID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Missing node or load error: '%s'", currentID))
			continue
		}

		// Inspect Transitions
		for _, t := range doc.Data.Transitions {
			target := t.To
			if target == "" {
				target = t.ToFull
			}

			if target == "" {
				continue // Sink state
			}

			// Check if target causes error ?
			// We optimize by just adding to queue.
			// If target doesn't exist, the next loop iteration's Get() will fail and report it.

			if !visited[target] {
				queue = append(queue, target)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("found %d errors:\n- %s", len(errors), strings.Join(errors, "\n- "))
	}

	return nil
}
