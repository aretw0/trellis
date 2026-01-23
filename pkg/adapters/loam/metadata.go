package loam

import (
	"github.com/aretw0/trellis/pkg/domain"
)

// NodeMetadata represents the header/metadata of a Trellis Node.
// It uses "mapstructure" tags to match standard Frontmatter/YAML keys (to, from).
type NodeMetadata struct {
	ID          string             `json:"id" mapstructure:"id"`
	Type        string             `json:"type" mapstructure:"type"`
	Transitions []LoaderTransition `json:"transitions" mapstructure:"transitions"`
	Options     []LoaderTransition `json:"options" mapstructure:"options"`
	OnError     string             `json:"on_error" mapstructure:"on_error"`
	OnSignal    map[string]string  `json:"on_signal" mapstructure:"on_signal"`
	Wait        bool               `json:"wait" mapstructure:"wait"`
	// SaveTo captures the input into a variable in the context
	SaveTo string `json:"save_to" mapstructure:"save_to"`

	// RequiredContext lists keys that MUST exist in the context
	RequiredContext []string `json:"required_context" mapstructure:"required_context"`

	// DefaultContext provides fallback values for context keys
	DefaultContext map[string]any `json:"default_context" mapstructure:"default_context"`

	// Timeout defines the maximum duration (e.g. "30s") to wait for input.
	Timeout string `json:"timeout,omitempty" mapstructure:"timeout"`

	// Interactive Input Config
	InputType    string   `json:"input_type" mapstructure:"input_type"`
	InputOptions []string `json:"input_options" mapstructure:"input_options"`
	InputDefault string   `json:"input_default" mapstructure:"input_default"`

	// Tool Config
	ToolCall *domain.ToolCall `json:"tool_call" mapstructure:"tool_call"`
	Do       *domain.ToolCall `json:"do" mapstructure:"do"`
	Tools    []any            `json:"tools" mapstructure:"tools"`
	Undo     *domain.ToolCall `json:"undo,omitempty" mapstructure:"undo"`

	// General Metadata
	Metadata map[string]string `json:"metadata" mapstructure:"metadata"`
}

type LoaderTransition struct {
	From      string `json:"from" mapstructure:"from"`
	FromFull  string `json:"from_node_id" mapstructure:"from_node_id"`
	To        string `json:"to" mapstructure:"to"`
	ToFull    string `json:"to_node_id" mapstructure:"to_node_id"`
	JumpTo    string `json:"jump_to" mapstructure:"jump_to"`
	Condition string `json:"condition" mapstructure:"condition"`
	// Text is the display label for options/buttons.
	// It is also used as the implicit match condition (Condition="input == Text") if Condition is empty.
	Text string `json:"text" mapstructure:"text"`
}
