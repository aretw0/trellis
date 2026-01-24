package domain

// ToolCall represents a request from the Engine to the Host to perform a side-effect.
// Ideally compatible with OpenAI/MCP tool call schemas.
type ToolCall struct {
	ID             string            `json:"id" yaml:"id" mapstructure:"id"`                                       // Unique ID for this specific call (e.g. from LLM or generated)
	Name           string            `json:"name" yaml:"name" mapstructure:"name"`                                 // Function name to call
	Args           map[string]any    `json:"args,omitempty" yaml:"args,omitempty" mapstructure:"args"`             // Arguments for the function
	Metadata       map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" mapstructure:"metadata"` // Context/Safety metadata from the Node
	IdempotencyKey string            `json:"idempotency_key,omitempty" yaml:"idempotency_key,omitempty" mapstructure:"idempotency_key"`
}

// ToolResult represents the output of a side-effect returned by the Host.
type ToolResult struct {
	ID       string `json:"id"` // Must match the ToolCall.ID
	Result   any    `json:"result,omitempty"`
	IsError  bool   `json:"is_error,omitempty"`
	IsDenied bool   `json:"is_denied,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Tool defines metadata about a tool available to the engine.
// This is used for generating schemas/prompts.
type Tool struct {
	Name        string         `json:"name" yaml:"name" mapstructure:"name"`
	Description string         `json:"description" yaml:"description" mapstructure:"description"`
	Parameters  map[string]any `json:"parameters,omitempty" yaml:"parameters,omitempty" mapstructure:"parameters"`
}
