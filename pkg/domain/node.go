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

	// NodeTypeTool executes an external side-effect (tool).
	NodeTypeTool = "tool"
)

// Node represents a logical unit in the graph.
// It can contain text content (for Wiki-style) or logic instructions (for Logic-style).
type Node struct {
	ID   string `json:"id" yaml:"id"`
	Type string `json:"type" yaml:"type"` // e.g., "text", "question", "logic", "tool"

	// Content holds the raw data for this node.
	// For a text node, it might be the markdown content.
	// For a logic node, it might be the script or parameters.
	Content []byte `json:"content" yaml:"content"`

	// Metadata allows for extensible key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Transitions defines the possible paths from this node.
	Transitions []Transition `json:"transitions" yaml:"transitions"`

	// Input Configuration (Optional)
	InputType    string   `json:"input_type,omitempty" yaml:"input_type,omitempty"`
	InputOptions []string `json:"input_options,omitempty" yaml:"input_options,omitempty"`
	InputDefault string   `json:"input_default,omitempty" yaml:"input_default,omitempty"`

	// Tool Configuration (Optional, used if Type == "tool")
	ToolCall *ToolCall `json:"tool_call,omitempty" yaml:"tool_call,omitempty"`
}
