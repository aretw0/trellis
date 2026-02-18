package runtime

import (
	"fmt"

	"github.com/aretw0/trellis/pkg/domain"
)

// validateExecution checks if the node configuration is logically sound.
func (e *Engine) validateExecution(node *domain.Node) error {
	if node == nil {
		return fmt.Errorf("cannot execute nil node")
	}

	// Forbidden: Concurrent side-effect (Do) and UI pause (Wait/Input)
	hasTool := node.Do != nil
	hasInput := node.Wait || node.InputType != "" || node.Type == domain.NodeTypeQuestion

	if hasTool && hasInput {
		return fmt.Errorf("node %s violation: cannot have both 'do' (tool) and 'wait/input' in the same node", node.ID)
	}

	// Warning: on_signal_default is only valid on the root/entry node
	if node.ID != e.entryNodeID && len(node.OnSignalDefault) > 0 {
		e.logger.Warn("on_signal_default configuration ignored: this property is only supported on the entry node",
			"node_id", node.ID,
			"entry_node_id", e.entryNodeID)
	}

	return nil
}
