package domain

import (
	"reflect"
)

// StateDiff represents the changes between two states.
// It is designed to be serialized to JSON for partial updates on the client.
type StateDiff struct {
	// SessionID is always present to identify the target.
	SessionID string `json:"session_id"`

	// OldNodeID needed? Maybe not. NewNodeID is enough.
	CurrentNodeID *string `json:"current_node_id,omitempty"`

	// Status changed?
	Status *ExecutionStatus `json:"status,omitempty"`

	// ContextDelta contains only changed, added or deleted keys.
	// For deletions, the key is present with a nil value.
	// Clients should merge these updates into their local state.
	Context map[string]any `json:"context,omitempty"`

	// HistoryDelta contains *new* items appended to history.
	// If history was rewritten (rare), we might send the whole list?
	// Let's optimize for append-only.
	HistoryParams *HistoryDelta `json:"history,omitempty"`

	// Terminated changed?
	Terminated *bool `json:"terminated,omitempty"`
}

// HistoryDelta represents changes to the history stack.
type HistoryDelta struct {
	Appended []string `json:"appended"`
}

// Diff calculates the difference between oldState and newState.
// If oldState is nil, it returns a diff representing the entire newState (initial load).
func Diff(oldState, newState *State) *StateDiff {
	if newState == nil {
		return nil
	}

	diff := &StateDiff{
		SessionID: newState.SessionID,
	}

	// 1. Check ID/Graph/Status changes
	if oldState == nil || oldState.CurrentNodeID != newState.CurrentNodeID {
		diff.CurrentNodeID = &newState.CurrentNodeID
	}
	if oldState == nil || oldState.Status != newState.Status {
		diff.Status = &newState.Status
	}
	if oldState == nil {
		if newState.Terminated {
			diff.Terminated = &newState.Terminated
		}
	} else if oldState.Terminated != newState.Terminated {
		diff.Terminated = &newState.Terminated
	}

	// 2. Context Diff
	diff.Context = diffContext(oldState, newState)

	// 3. History Diff
	diff.HistoryParams = diffHistory(oldState, newState)

	// Optimization: If nothing changed (besides ID which is always present?), return nil?
	// But we initialized SessionID. Let's check if there are actual payloads.
	if diff.CurrentNodeID == nil &&
		diff.Status == nil &&
		diff.Terminated == nil &&
		len(diff.Context) == 0 &&
		diff.HistoryParams == nil {
		return nil
	}

	return diff
}

func diffContext(old *State, new *State) map[string]any {
	delta := make(map[string]any)

	// If old is nil, everything in new is a delta
	if old == nil {
		for k, v := range new.Context {
			delta[k] = v
		}
		return delta
	}

	// Check for Added or Modified
	for k, newVal := range new.Context {
		oldVal, exists := old.Context[k]
		if !exists {
			delta[k] = newVal
		} else {
			if !reflect.DeepEqual(oldVal, newVal) {
				delta[k] = newVal
			}
		}
	}

	// Check for Deletions
	for k := range old.Context {
		if _, exists := new.Context[k]; !exists {
			delta[k] = nil
		}
	}

	// Optimization: Return nil if delta is empty so omitempty can remove the key
	if len(delta) == 0 {
		return nil
	}
	return delta
}

// diffHistory assumes standard append-only behavior for History.
func diffHistory(old *State, new *State) *HistoryDelta {
	if new == nil || len(new.History) == 0 {
		return nil
	}

	if old == nil {
		return &HistoryDelta{Appended: new.History}
	}

	// Detect Append
	oldLen := len(old.History)
	newLen := len(new.History)

	if newLen > oldLen {
		// Verify prefix matches matches (sanity check)
		// If prefix mismatch, it's a rewrite, so send everything?
		// For now assume append-only.
		return &HistoryDelta{
			Appended: new.History[oldLen:],
		}
	}

	return nil
}

// IsEmpty checks if the diff contains any actionable changes.
func (d *StateDiff) IsEmpty() bool {
	return d.CurrentNodeID == nil &&
		d.Status == nil &&
		d.Terminated == nil &&
		len(d.Context) == 0 &&
		d.HistoryParams == nil
}
