package domain

// ToolCall represents a request from the Engine to the Host to perform a side-effect.
// Ideally compatible with OpenAI/MCP tool call schemas.
type ToolCall struct {
	ID   string         `json:"id"`             // Unique ID for this specific call (e.g. from LLM or generated)
	Name string         `json:"name"`           // Function name to call
	Args map[string]any `json:"args,omitempty"` // Arguments for the function
}

// ToolResult represents the output of a side-effect returned by the Host.
type ToolResult struct {
	ID      string `json:"id"` // Must match the ToolCall.ID
	Result  any    `json:"result,omitempty"`
	IsError bool   `json:"is_error,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Tool defines metadata about a tool available to the engine.
// This is used for generating schemas/prompts.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Additional schema definition can be added here (JSON Schema)
}
