package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/aretw0/trellis/internal/compiler"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// ConditionEvaluator is a function that determines if a transition condition is met.
type ConditionEvaluator func(ctx context.Context, condition string, input string) (bool, error)

// Engine is the core state machine runner.
type Engine struct {
	loader    ports.GraphLoader
	parser    *compiler.Parser
	evaluator ConditionEvaluator
}

// DefaultEvaluator implements the basic "condition: input == 'value'" logic.
func DefaultEvaluator(ctx context.Context, condition string, input string) (bool, error) {
	if condition == "" {
		return true, nil
	}
	// Simple input matching "input == 'yes'"
	if strings.Contains(condition, "input ==") {
		parts := strings.Split(condition, "==")
		if len(parts) == 2 {
			expected := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
			// Case-insensitive matching
			if strings.EqualFold(strings.TrimSpace(input), expected) {
				return true, nil
			}
		}
	}
	return false, nil
}

// NewEngine creates a new engine with dependencies.
func NewEngine(loader ports.GraphLoader) *Engine {
	return &Engine{
		loader:    loader,
		parser:    compiler.NewParser(),
		evaluator: DefaultEvaluator, // Set default
	}
}

// SetEvaluator sets the condition evaluator for the engine.
func (e *Engine) SetEvaluator(eval ConditionEvaluator) {
	e.evaluator = eval
}

// Step executes a single step in the state machine.
// It loads the current node, evaluates transitions, and returns action requests.
// For MVP, it doesn't persist state, just inputs state and outputs next state + actions.
// But wait, the Step function usually takes the current State and Input?
// BOOTSTRAP: "- Input: Estado Atual + Grafo de Decisão + Input do Usuário."
func (e *Engine) Step(ctx context.Context, currentState *domain.State, input string) ([]domain.ActionRequest, *domain.State, error) {
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
	// Present content only if we are "visiting" the node (input is empty).
	// If input is provided, we are "submitting", so we skip re-rendering the content.
	if (node.Type == "text" || node.Type == "question") && input == "" {
		// Interpolation could happen here.
		text := string(node.Content)
		actions = append(actions, domain.ActionRequest{
			Type:    domain.ActionRenderContent,
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

		// Evaluate Condition
		if e.evaluator != nil {
			ok, err := e.evaluator(ctx, t.Condition, input)
			if err != nil {
				return nil, nil, fmt.Errorf("condition evaluation failed for '%s': %w", t.Condition, err)
			}
			if ok {
				nextNodeID = t.ToNodeID
				break
			}
		}
	}

	nextState := *currentState // Copy

	// Check for Sink State
	if len(node.Transitions) == 0 {
		nextState.Terminated = true
	}
	if nextNodeID != "" {
		nextState.CurrentNodeID = nextNodeID
		nextState.History = append(nextState.History, nextNodeID)
	}

	return actions, &nextState, nil
}
