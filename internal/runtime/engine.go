package runtime

import (
	"fmt"
	"strings"

	"github.com/aretw0/trellis/internal/compiler"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// Engine is the core state machine runner.
type Engine struct {
	loader ports.GraphLoader
	parser *compiler.Parser
}

// NewEngine creates a new engine with dependencies.
func NewEngine(loader ports.GraphLoader) *Engine {
	return &Engine{
		loader: loader,
		parser: compiler.NewParser(),
	}
}

// Step executes a single step in the state machine.
// It loads the current node, evaluates transitions, and returns action requests.
// For MVP, it doesn't persist state, just inputs state and outputs next state + actions.
// But wait, the Step function usually takes the current State and Input?
// BOOTSTRAP: "- Input: Estado Atual + Grafo de Decisão + Input do Usuário."
func (e *Engine) Step(currentState *domain.State, input string) ([]domain.ActionRequest, *domain.State, error) {
	// 1. Load the definition of the current node
	raw, err := e.loader.GetNode(currentState.CurrentNodeID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load node %s: %w", currentState.CurrentNodeID, err)
	}

	node, err := e.parser.Parse(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse node %s: %w", currentState.CurrentNodeID, err)
	}

	// 2. Process the node's logic
	// For MVP, we'll just check if it's a text node and return a print action.
	actions := []domain.ActionRequest{}

	// If it's a text node, we probably want to display it.
	// FIX: Only display content if we are "visiting" (input is empty), not "submitting".
	if (node.Type == "text" || node.Type == "question") && input == "" {
		// Interpolation could happen here.
		text := string(node.Content)
		actions = append(actions, domain.ActionRequest{
			Type:    "CLI_PRINT",
			Payload: text,
		})
	}

	// 3. Determine next transition
	var nextNodeID string
	for _, t := range node.Transitions {
		if t.Condition == "" {
			// Default/Always transition
			nextNodeID = t.ToNodeID
			break
		}
		// Simple input matching "input == 'yes'"
		if input != "" && strings.Contains(t.Condition, "input ==") {
			// quick hacky parser for MVP
			parts := strings.Split(t.Condition, "==")
			if len(parts) == 2 {
				expected := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
				// Case-insensitive matching
				if strings.EqualFold(strings.TrimSpace(input), expected) {
					nextNodeID = t.ToNodeID
					break
				}
			}
		}
	}

	nextState := *currentState // Copy
	if nextNodeID != "" {
		nextState.CurrentNodeID = nextNodeID
		nextState.History = append(nextState.History, nextNodeID)
	}

	return actions, &nextState, nil
}
