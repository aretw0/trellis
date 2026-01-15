package domain

// NodeType constants define the control flow behavior.
const (
	// NodeTypeText displays content and continues immediately (soft step).
	NodeTypeText = "text"
	// NodeTypeQuestion displays content and halts waiting for input (hard step).
	// NOTE: Future architecture may merge this with InputType logic.
	NodeTypeQuestion = "question"

	// NodeTypeTool executes an external side-effect (tool).
	NodeTypeTool = "tool"

	// NodeTypeStart indicates the entry point (typically convention-based, but can be explicit).
	NodeTypeStart = "start"
	// NodeTypePrompt is an alias for "question", waiting for user input.
	NodeTypePrompt = "prompt"
)

// Node represents a logical unit in the graph.
// It can contain text content (for Wiki-style) or logic instructions (for Logic-style).
type Node struct {
	ID   string `json:"id" yaml:"id"`
	Type string `json:"type" yaml:"type"` // e.g., "text", "question", "logic", "tool"
	// Wait indicates if the engine should pause for input after rendering.
	Wait bool `json:"wait" yaml:"wait"`
	// SaveTo indicates the variable name in Context where input should be stored.
	SaveTo string `json:"save_to,omitempty" yaml:"save_to,omitempty"`

	// RequiredContext lists keys that MUST exist in the context for this node to execute.
	RequiredContext []string `json:"required_context,omitempty" yaml:"required_context,omitempty"`

	// Content holds the raw data for this node.
	// For a text node, it might be the markdown content.
	// For a logic node, it might be the script or parameters.
	Content []byte `json:"content" yaml:"content"`

	// Metadata allows for extensible key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Transitions defines the possible paths from this node.
	Transitions []Transition `json:"transitions" yaml:"transitions"`

	// OnError defines the node ID to transition to if a Tool returns an error.
	OnError string `json:"on_error,omitempty" yaml:"on_error,omitempty"`

	// Input Configuration (Optional)
	InputType    string   `json:"input_type,omitempty" yaml:"input_type,omitempty"`
	InputOptions []string `json:"input_options,omitempty" yaml:"input_options,omitempty"`
	InputDefault string   `json:"input_default,omitempty" yaml:"input_default,omitempty"`

	// Tool Configuration (Optional, used if Type == "tool")
	ToolCall *ToolCall `json:"tool_call,omitempty" yaml:"tool_call,omitempty"`

	// Tools defined within this node (e.g. for LLM context)
	Tools []Tool `json:"tools,omitempty" yaml:"tools,omitempty"`
}
