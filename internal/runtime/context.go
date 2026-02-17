package runtime

import (
	"fmt"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/schema"
)

// ContextValidationError represents a failure to meet context requirements.
type ContextValidationError struct {
	NodeID      string
	MissingKeys []string
}

func (e *ContextValidationError) Error() string {
	return fmt.Sprintf("Node '%s' requires context keys that are missing: %v", e.NodeID, e.MissingKeys)
}

func (e *Engine) validateContext(node *domain.Node, state *domain.State) error {
	if len(node.RequiredContext) == 0 {
		return e.validateContextTypes(node, state)
	}

	var missing []string
	for _, key := range node.RequiredContext {
		found := false
		if _, ok := state.Context[key]; ok {
			found = true
		} else if _, ok := state.SystemContext[key]; ok {
			found = true
		}

		if !found {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return &ContextValidationError{
			NodeID:      node.ID,
			MissingKeys: missing,
		}
	}

	return e.validateContextTypes(node, state)
}

// ContextTypeValidationError represents a failure to meet context schema requirements.
type ContextTypeValidationError struct {
	NodeID string
	Err    error
}

func (e *ContextTypeValidationError) Error() string {
	return fmt.Sprintf("Node '%s' has invalid context types: %s", e.NodeID, e.Err.Error())
}

func (e *Engine) validateContextTypes(node *domain.Node, state *domain.State) error {
	if len(node.ContextSchema) == 0 {
		return nil
	}

	data := make(map[string]any)
	for key, value := range state.Context {
		data[key] = value
	}
	for key, value := range state.SystemContext {
		if _, exists := data[key]; !exists {
			data[key] = value
		}
	}

	if err := schema.Validate(node.ContextSchema, data); err != nil {
		return &ContextTypeValidationError{
			NodeID: node.ID,
			Err:    err,
		}
	}

	return nil
}

// applyInput handles the Update Phase: Creates new state and applies SaveTo logic.
// It also automatically populates the implicit 'sys.ans' variable for zero-friction data propagation.
func (e *Engine) applyInput(currentState *domain.State, node *domain.Node, input any) (*domain.State, error) {
	nextState := e.cloneState(currentState)

	// Implicit Propagation: Always store the latest result in 'sys.ans'
	if nextState.SystemContext == nil {
		nextState.SystemContext = make(map[string]any)
	}
	nextState.SystemContext["ans"] = input

	// Explicit Persistence: Save to a named key if configured
	if node.SaveTo != "" {
		if node.SaveTo == "sys" || strings.HasPrefix(node.SaveTo, "sys.") {
			return nil, fmt.Errorf("security violation: cannot save to reserved namespace 'sys' in node %s", node.ID)
		}
		nextState.Context[node.SaveTo] = input
	}
	return nextState, nil
}
