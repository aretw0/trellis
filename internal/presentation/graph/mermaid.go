package graph

import (
	"fmt"
	"path"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// GraphOverlay contains dynamic state data to visualize on the graph.
type GraphOverlay struct {
	VisitedNodes []string
	CurrentNode  string
}

// GenerateMermaid produces a Mermaid flowchart syntax string from a list of nodes.
// It applies semantic styling:
// - Start: ((Circle))
// - Tool: [[Subroutine]]
// - Input (Question/Prompt): [/Parallelogram/]
// - Default: [Rectangle]
// It also applies overlay styles (Visited/Current) if provided.
func GenerateMermaid(nodes []domain.Node, overlay *GraphOverlay) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	for _, node := range nodes {
		// Sanitize ID for Mermaid
		safeID := sanitizeMermaidID(node.ID)

		// Node Shape based on Type
		opener, closer := "[", "]"

		switch {
		case node.ID == "start" || node.Type == domain.NodeTypeStart: // "start" (legacy or explicit)
			opener, closer = "((", "))" // Circle
		case node.Type == domain.NodeTypeTool:
			opener, closer = "[[", "]]" // Subroutine
		case node.Type == domain.NodeTypeQuestion:
			opener, closer = "[/", "/]" // Parallelogram (Input)
		}

		label := fmt.Sprintf("    %s%s\"%s\"%s\n", safeID, opener, node.ID, closer)
		if node.Timeout != "" {
			// Annotate node with Timeout clock icon or text
			label = fmt.Sprintf("    %s%s\"%s <br/> ⏱️ %s\"%s\n", safeID, opener, node.ID, node.Timeout, closer)
		}
		sb.WriteString(label)

		// Transitions
		for _, t := range node.Transitions {
			safeTo := sanitizeMermaidID(t.ToNodeID)

			// Determine if it's a cross-module transition (Jump)
			fromDir := path.Dir(node.ID)
			toDir := path.Dir(t.ToNodeID)
			isJump := fromDir != toDir

			arrow := "-->"
			if isJump {
				arrow = "-.->"
			}
			if t.Condition != "" {
				// Escape double quotes in condition for Mermaid label
				safeCondition := strings.ReplaceAll(t.Condition, "\"", "'")
				arrow = fmt.Sprintf("-- \"%s\" -->", safeCondition)
				if isJump {
					arrow = fmt.Sprintf("-. \"%s\" .->", safeCondition)
				}
			}
			sb.WriteString(fmt.Sprintf("    %s %s %s\n", safeID, arrow, safeTo))
		}

		// Signal Transitions (Intervention)
		for signalName, targetID := range node.OnSignal {
			safeTo := sanitizeMermaidID(targetID)
			// Use dotted line with lightning bolt/signal icon
			arrow := fmt.Sprintf("-. ⚡ %s .->", signalName)
			sb.WriteString(fmt.Sprintf("    %s %s %s\n", safeID, arrow, safeTo))
		}
	}

	// Apply Overlay Styles
	if overlay != nil {
		sb.WriteString("\n    %% Overlay Styles\n")
		// Force black text (color:#000) for high-contrast on light backgrounds, regardless of theme (Light/Dark)
		sb.WriteString("    classDef visited fill:#e1f5fe,stroke:#01579b,stroke-width:2px,color:#000;\n")
		sb.WriteString("    classDef current fill:#ffeb3b,stroke:#fbc02d,stroke-width:4px,color:#000;\n")

		// Deduplicate visited nodes (using safeIDs)
		visitedSet := make(map[string]bool)
		for _, id := range overlay.VisitedNodes {
			// Only style valid nodes (some history might point to deleted/dynamic nodes?)
			safeID := sanitizeMermaidID(id)
			if !visitedSet[safeID] && safeID != "" {
				visitedSet[safeID] = true
				sb.WriteString(fmt.Sprintf("    class %s visited;\n", safeID))
			}
		}

		if overlay.CurrentNode != "" {
			safeCurrent := sanitizeMermaidID(overlay.CurrentNode)
			sb.WriteString(fmt.Sprintf("    class %s current;\n", safeCurrent))
		}
	}

	return sb.String()
}

func sanitizeMermaidID(id string) string {
	s := strings.ReplaceAll(id, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}
