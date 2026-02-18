package runner

import (
	"context"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// RichResponse combines state and rendering actions for rich clients (Web, MCP, etc).
// This encapsulates the common pattern of: Navigate -> Render -> Return Actions.
type RichResponse struct {
	State    *domain.State          `json:"state"`
	Actions  []domain.ActionRequest `json:"actions,omitempty"`
	Terminal bool                   `json:"terminal"`
}

// NavigateAndRender performs a navigation step and immediately renders the resulting state.
// This ensures that rich clients always receive the content/instructions for the node they just entered.
func NavigateAndRender(ctx context.Context, engine ports.StatelessEngine, currentState *domain.State, input any) (*RichResponse, error) {
	newState, err := engine.Navigate(ctx, currentState, input)
	if err != nil {
		return nil, err
	}

	actions, terminal, err := engine.Render(ctx, newState)
	if err != nil {
		// Even if render fails, we return the new state to allow the client to recover.
		// However, we still return the error to let the adapter decide how to log/handle it.
		return &RichResponse{State: newState, Terminal: terminal}, err
	}

	return &RichResponse{
		State:    newState,
		Actions:  actions,
		Terminal: terminal,
	}, nil
}

// SignalAndRender performs a signal transition and immediately renders the resulting state.
func SignalAndRender(ctx context.Context, engine ports.StatelessEngine, currentState *domain.State, signalName string) (*RichResponse, error) {
	newState, err := engine.Signal(ctx, currentState, signalName)
	if err != nil {
		return nil, err
	}

	actions, terminal, err := engine.Render(ctx, newState)
	if err != nil {
		return &RichResponse{State: newState, Terminal: terminal}, err
	}

	return &RichResponse{
		State:    newState,
		Actions:  actions,
		Terminal: terminal,
	}, nil
}
