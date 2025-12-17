package domain

// NodeType constants define the control flow behavior.
const (
	// NodeTypeText displays content and continues immediately (soft step).
	NodeTypeText = "text"
	// NodeTypeQuestion displays content and halts waiting for input (hard step).
	// NOTE: Future architecture may merge this with InputType logic.
	NodeTypeQuestion = "question"
	// NodeTypeLogic executes internal script/logic (silent step).
	NodeTypeLogic = "logic"
)

// Node represents a logical unit in the graph.
// It can contain text content (for Wiki-style) or logic instructions (for Logic-style).
type Node struct {
	ID   string `json:"id"`
	Type string `json:"type"` // e.g., "text", "question", "logic"

	// Content holds the raw data for this node.
	// For a text node, it might be the markdown content.
	// For a logic node, it might be the script or parameters.
	Content []byte `json:"content"`

	// Metadata allows for extensible key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Transitions defines the possible paths from this node.
	Transitions []Transition `json:"transitions"`

	// Input Configuration (Optional)
	InputType    string   `json:"input_type,omitempty"`
	InputOptions []string `json:"input_options,omitempty"`
	InputDefault string   `json:"input_default,omitempty"`
}
