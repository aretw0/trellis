package loam

// NodeMetadata represents the header/metadata of a Trellis Node.
// It uses "mapstructure" tags to match standard Frontmatter/YAML keys (to, from).
type NodeMetadata struct {
	ID          string             `json:"id" mapstructure:"id"`
	Type        string             `json:"type" mapstructure:"type"`
	Transitions []LoaderTransition `json:"transitions" mapstructure:"transitions"`
	Options     []LoaderTransition `json:"options" mapstructure:"options"`
	// To is a shorthand for a single unconditional transition
	To       string            `json:"to" mapstructure:"to"`
	OnError  string            `json:"on_error" mapstructure:"on_error"`
	OnDenied string            `json:"on_denied" mapstructure:"on_denied"`
	OnSignal map[string]string `json:"on_signal" mapstructure:"on_signal"`
	// OnTimeout is syntactic sugar for on_signal["timeout"]
	OnTimeout string `json:"on_timeout" mapstructure:"on_timeout"`
	// OnInterrupt is syntactic sugar for on_signal["interrupt"]
	OnInterrupt string `json:"on_interrupt" mapstructure:"on_interrupt"`
	Wait        bool   `json:"wait" mapstructure:"wait"`
	// SaveTo captures the input into a variable in the context
	SaveTo string `json:"save_to" mapstructure:"save_to"`

	// RequiredContext lists keys that MUST exist in the context
	RequiredContext []string `json:"required_context" mapstructure:"required_context"`

	// DefaultContext provides fallback values for context keys
	DefaultContext map[string]any `json:"default_context" mapstructure:"default_context"`

	// ContextSchema defines expected types for context values
	ContextSchema map[string]any `json:"context_schema" mapstructure:"context_schema"`

	// Timeout defines the maximum duration (e.g. "30s") to wait for input.
	Timeout string `json:"timeout,omitempty" mapstructure:"timeout"`

	// Interactive Input Config
	InputType    string   `json:"input_type" mapstructure:"input_type"`
	InputOptions []string `json:"input_options" mapstructure:"input_options"`
	InputDefault string   `json:"input_default" mapstructure:"input_default"`

	// Tool Config
	ToolCall *LoaderToolCall `json:"tool_call" mapstructure:"tool_call"`
	Do       *LoaderToolCall `json:"do" mapstructure:"do"`
	Tools    []any           `json:"tools" mapstructure:"tools"`
	Undo     *LoaderToolCall `json:"undo,omitempty" mapstructure:"undo"`

	// General Metadata
	Metadata map[string]any `json:"metadata" mapstructure:"metadata"`
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

// LoaderToolCall is a permissive version of domain.ToolCall for YAML decoding.
type LoaderToolCall struct {
	ID             string         `json:"id" mapstructure:"id"`
	Name           string         `json:"name" mapstructure:"name"`
	Args           map[string]any `json:"args" mapstructure:"args"`
	Metadata       map[string]any `json:"metadata" mapstructure:"metadata"`
	IdempotencyKey string         `json:"idempotency_key" mapstructure:"idempotency_key"`
}
