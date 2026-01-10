package domain

// ActionRequest represents a side-effect that the engine requests the host to perform.
type ActionRequest struct {
	Type    string // e.g., "CLI_PRINT", "HTTP_GET"
	Payload any    // The data needed to perform the action
}

// Standard Action Types
const (
	// ActionRenderContent requests the host to display content to the user.
	// Payload: string (the content)
	ActionRenderContent = "RENDER_CONTENT"

	// ActionRequestInput requests the host to collect input from the user.
	// Payload: InputRequest
	ActionRequestInput = "REQUEST_INPUT"

	// ActionCallTool requests the host to execute a side-effect (tool).
	// Payload: ToolCall
	ActionCallTool = "CALL_TOOL"
)

// InputType defines the kind of input requested.
type InputType string

const (
	InputText    InputType = "text"
	InputConfirm InputType = "confirm"
	InputChoice  InputType = "choice"
)

// InputRequest describes the constraints and type of input needed.
type InputRequest struct {
	Type    InputType `json:"type"`
	Options []string  `json:"options,omitempty"`
	Default string    `json:"default,omitempty"`
}

// ActionResponse represents the result of an ActionRequest.
type ActionResponse struct {
	Success bool
	Data    any
	Error   error
}
