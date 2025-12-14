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
)

// ActionResponse represents the result of an ActionRequest.
type ActionResponse struct {
	Success bool
	Data    any
	Error   error
}
