package domain

// State represents the current snapshot of the execution.
type State struct {
	// CurrentNodeID is the identifier of the active node.
	CurrentNodeID string

	// Memory holds variable state for the session.
	Memory map[string]any

	// History could track the path taken (optional for now, but good for debugging)
	History []string

	// Terminated indicates if the execution has reached a sink state (no transitions).
	Terminated bool
}

// NewState creates a clean state starting at a specific node.
func NewState(startNodeID string) *State {
	return &State{
		CurrentNodeID: startNodeID,
		Memory:        make(map[string]any),
		History:       []string{startNodeID},
	}
}
