package runtime

import (
	"context"
	"fmt"

	"github.com/aretw0/trellis/pkg/domain"
)

// startRollback initiates the SAGA rollback process.
// It is called when a tool fails and on_error is set to "rollback" or an unconditional rollback is triggered.
func (e *Engine) startRollback(ctx context.Context, failedState *domain.State) (*domain.State, error) {
	// We assume the *current* node failed or was aborted, so it had no side effect.
	// We pop it immediately to begin unwinding from the last successful node.
	return e.continueRollback(ctx, failedState, true)
}

// continueRollback unwinds the history stack until it finds a node with an Undo action.
// popCurrent: If true, removes the current head of history before searching.
func (e *Engine) continueRollback(ctx context.Context, state *domain.State, popCurrent bool) (*domain.State, error) {
	e.logger.InfoContext(ctx, "continuing rollback", "history_len", len(state.History), "pop_current", popCurrent)

	nextState := e.cloneState(state)
	nextState.Status = domain.StatusRollingBack
	nextState.PendingToolCall = ""

	if popCurrent && len(nextState.History) > 0 {
		nextState.History = nextState.History[:len(nextState.History)-1]
	}

	// Unwind Loop: Search backwards through history for compensatable actions.
	for len(nextState.History) > 0 {
		currentNodeID := nextState.History[len(nextState.History)-1]
		nextState.CurrentNodeID = currentNodeID

		raw, err := e.loader.GetNode(currentNodeID)
		if err != nil {
			return nil, fmt.Errorf("rollback failed: could not load node %s: %w", currentNodeID, err)
		}
		node, err := e.parser.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("rollback failed: could not parse node %s: %w", currentNodeID, err)
		}

		if node.Undo != nil {
			// Compensation Found: Stay on this node and prepare the Undo action.
			nextState.PendingToolCall = node.Undo.ID
			if nextState.PendingToolCall == "" {
				nextState.PendingToolCall = node.Undo.Name
			}
			return nextState, nil
		}

		// Read-only step or no compensation defined: pop and continue unwinding.
		nextState.History = nextState.History[:len(nextState.History)-1]
	}

	// Termination Protocol:
	// If history is fully unwound, the rollback is complete.
	// The state is marked as Terminated to halt the runner loop gracefully.
	nextState.Status = domain.StatusTerminated
	nextState.CurrentNodeID = ""
	return nextState, nil
}
