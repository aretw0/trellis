package graph

import (
	"fmt"
	"path"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// GenerateMermaid produces a Mermaid flowchart syntax string from a list of nodes.
// It applies semantic styling:
// - Start: ((Circle))
// - Tool: [[Subroutine]]
// - Input (Question/Prompt): [/Parallelogram/]
// - Default: [Rectangle]
func GenerateMermaid(nodes []domain.Node) string {
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

	return sb.String()
}

func sanitizeMermaidID(id string) string {
	s := strings.ReplaceAll(id, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}
