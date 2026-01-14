package domain

// ExecutionStatus defines the current mode of the engine mechanics.
type ExecutionStatus string

const (
	StatusActive         ExecutionStatus = "active"           // Normal operation
	StatusWaitingForTool ExecutionStatus = "waiting_for_tool" // Engine is paused, waiting for Host result
	StatusTerminated     ExecutionStatus = "terminated"       // Sink state reached
)

// State represents the current snapshot of the execution.
type State struct {
	// CurrentNodeID is the identifier of the active node.
	CurrentNodeID string

	// Status indicates if the engine is running, waiting, or done.
	Status ExecutionStatus

	// PendingToolCall holds the ID of the tool call we are waiting for (if Status == WaitingForTool).
	PendingToolCall string

	// Context holds variable state for the session.
	// Context holds variable state for the session (User space).
	Context map[string]any

	// SystemContext holds system-level state (Read-only for templates, Host-writable).
	// Reserved namespace: "sys".
	SystemContext map[string]any

	// History could track the path taken (optional for now, but good for debugging)
	History []string

	// Terminated indicates if the execution has reached a sink state (no transitions).
	// Deprecated: Use Status == StatusTerminated instead. Kept for backward compat.
	Terminated bool
}

// NewState creates a clean state starting at a specific node.
func NewState(startNodeID string) *State {
	return &State{
		CurrentNodeID: startNodeID,
		Status:        StatusActive,
		Context:       make(map[string]any),
		SystemContext: make(map[string]any),
		History:       []string{startNodeID},
	}
}
