package validator

import (
	"fmt"
	"strings"

	"github.com/aretw0/trellis/internal/compiler"
	"github.com/aretw0/trellis/pkg/ports"
)

// ValidateGraph checks for broken links and unreachable nodes starting from startNodeID.
func ValidateGraph(loader ports.GraphLoader, parser *compiler.Parser, startNodeID string) error {

	// 0. Structural Validation (Detect Collisions)
	if _, err := loader.ListNodes(); err != nil {
		return fmt.Errorf("structural validation failed: %w", err)
	}

	// 1. Get raw start node to verify existence and ID
	startNodeRaw, err := loader.GetNode(startNodeID)
	if err != nil {
		return fmt.Errorf("start node '%s' not found: %w", startNodeID, err)
	}

	startNode, err := parser.Parse(startNodeRaw)
	if err != nil {
		return fmt.Errorf("failed to parse start node '%s': %w", startNodeID, err)
	}

	// Resolve the canonical ID from the parsed node (e.g. if startNodeID alias was used)
	actualStartID := startNode.ID
	if actualStartID == "" {
		actualStartID = startNodeID
	}

	// 2. Crawler
	visited := make(map[string]bool)
	queue := []string{actualStartID}

	visited[actualStartID] = true

	var errors []string

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		// Load Node
		raw, err := loader.GetNode(currentID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Missing node or load error: '%s'", currentID))
			continue
		}

		node, err := parser.Parse(raw)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Invalid node content '%s': %v", currentID, err))
			continue
		}

		// Inspect Transitions
		for _, t := range node.Transitions {
			target := t.ToNodeID

			if target == "" {
				continue // Sink state or conditional without explicit target (shouldn't happen in valid graph?)
			}

			if !visited[target] {
				visited[target] = true
				queue = append(queue, target)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("found %d errors:\n- %s", len(errors), strings.Join(errors, "\n- "))
	}

	return nil
}
