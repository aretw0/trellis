package domain

// Node represents a logical unit in the graph.
// It can contain text content (for Wiki-style) or logic instructions (for Logic-style).
type Node struct {
	ID   string
	Type string // e.g., "text", "question", "logic"

	// Content holds the raw data for this node.
	// For a text node, it might be the markdown content.
	// For a logic node, it might be the script or parameters.
	Content []byte

	// Metadata allows for extensible key-value pairs.
	Metadata map[string]string

	// Transitions defines the possible paths from this node.
	Transitions []Transition
}
