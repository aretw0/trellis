package dto

import (
	"github.com/aretw0/trellis/pkg/domain"
)

// NodeMetadata represents the header/metadata of a Trellis Node.
// It uses "mapstructure" tags to match standard Frontmatter/YAML keys (to, from).
type NodeMetadata struct {
	ID          string             `json:"id" mapstructure:"id"`
	Type        string             `json:"type" mapstructure:"type"`
	Transitions []LoaderTransition `json:"transitions" mapstructure:"transitions"`

	// Interactive Input Config
	InputType    string   `json:"input_type" mapstructure:"input_type"`
	InputOptions []string `json:"input_options" mapstructure:"input_options"`
	InputDefault string   `json:"input_default" mapstructure:"input_default"`

	// Tool Config
	ToolCall *domain.ToolCall `json:"tool_call" mapstructure:"tool_call"`

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
}
