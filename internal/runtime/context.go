package runtime

import (
	"fmt"

	"github.com/aretw0/trellis/pkg/domain"
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
		return nil
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
	return nil
}
