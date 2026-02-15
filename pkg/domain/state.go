package domain

// ExecutionStatus defines the current mode of the engine mechanics.
type ExecutionStatus string

const (
	StatusActive         ExecutionStatus = "active"           // Normal operation
	StatusWaitingForTool ExecutionStatus = "waiting_for_tool" // Engine is paused, waiting for Host result
	StatusRollingBack    ExecutionStatus = "rolling_back"     // Engine is unwinding history (SAGA)
	StatusTerminated     ExecutionStatus = "terminated"       // Sink state reached
)

// State represents the current snapshot of the execution.
type State struct {
	// SessionID uniquely identifies the session this state belongs to.
	SessionID string `json:"session_id"`

	// CurrentNodeID is the identifier of the active node.
	CurrentNodeID string `json:"current_node_id"`

	// Status indicates if the engine is running, waiting, or done.
	Status ExecutionStatus `json:"status"`

	// PendingToolCall holds the ID of the tool call we are waiting for (if Status == WaitingForTool).
	PendingToolCall string `json:"pending_tool_call,omitempty"`

	// Context holds variable state for the session (User space).
	Context map[string]any `json:"context"`

	// SystemContext holds system-level state (Read-only for templates, Host-writable).
	// Reserved namespace: "sys".
	SystemContext map[string]any `json:"system_context"`

	// History could track the path taken (optional for now, but good for debugging)
	History []string `json:"history,omitempty"`

	// Terminated indicates if the execution has reached a sink state (no transitions).
	// Deprecated: Use Status == StatusTerminated instead. Kept for backward compat.
	Terminated bool `json:"terminated,omitempty"`
}

// Snapshot creates a deep copy of the state for observability purposes.
// This is thread-safe as long as the caller ensures exclusive access to the source state during copying.
func (s *State) Snapshot() *State {
	if s == nil {
		return nil
	}

	// Deep copy maps
	ctxCopy := make(map[string]any, len(s.Context))
	for k, v := range s.Context {
		ctxCopy[k] = v // Shallow copy of values (assuming primitives or treated as such)
	}

	sysCtxCopy := make(map[string]any, len(s.SystemContext))
	for k, v := range s.SystemContext {
		sysCtxCopy[k] = v
	}

	// Deep copy slices
	histCopy := make([]string, len(s.History))
	copy(histCopy, s.History)

	return &State{
		SessionID:       s.SessionID,
		CurrentNodeID:   s.CurrentNodeID,
		Status:          s.Status,
		PendingToolCall: s.PendingToolCall,
		Context:         ctxCopy,
		SystemContext:   sysCtxCopy,
		History:         histCopy,
		Terminated:      s.Terminated,
	}
}

// NewState creates a clean state starting at a specific node.
func NewState(sessionID, startNodeID string) *State {
	return &State{
		SessionID:     sessionID,
		CurrentNodeID: startNodeID,
		Status:        StatusActive,
		Context:       make(map[string]any),
		SystemContext: make(map[string]any),
		History:       []string{startNodeID},
	}
}
