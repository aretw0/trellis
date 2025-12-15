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

// Step executes a single step (Render + Transition).
// Deprecated: Use Render and Navigate separately for better control.
func (e *Engine) Step(ctx context.Context, currentState *domain.State, input string) ([]domain.ActionRequest, *domain.State, error) {
	// 1. Render content (View)
	actions, _, err := e.Render(ctx, currentState)
	if err != nil {
		return nil, nil, err
	}

	// 2. Compute transition (Controller)
	nextState, err := e.Navigate(ctx, currentState, input)
	if err != nil {
		return nil, nil, err
	}

	return actions, nextState, nil
}

// Render calculates the presentation for the current state.
// It loads the node and generates actions (e.g. print text) but does NOT change state.
// It returns actions, isTerminal (true if no transitions), and error.
func (e *Engine) Render(ctx context.Context, currentState *domain.State) ([]domain.ActionRequest, bool, error) {
	raw, err := e.loader.GetNode(currentState.CurrentNodeID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load node %s: %w", currentState.CurrentNodeID, err)
	}

	node, err := e.parser.Parse(raw)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse node %s: %w", currentState.CurrentNodeID, err)
	}

	actions := []domain.ActionRequest{}

	// Only render content if we are NOT submitting data (which usually implies moving away)
	// But in the new architecture, Render is called explicitly before Navigate, so we always render.
	// It's up to the Runner to decide if it shows it or not based on previous history?
	// Actually, Render just returns what the node *says*.
	if node.Type == "text" || node.Type == "question" {
		text := string(node.Content)
		text = interpolate(text, currentState.Memory)

		actions = append(actions, domain.ActionRequest{
			Type:    domain.ActionRenderContent,
			Payload: text,
		})
	}

	isTerminal := len(node.Transitions) == 0
	return actions, isTerminal, nil
}

// Navigate determines the next state based on input.
func (e *Engine) Navigate(ctx context.Context, currentState *domain.State, input string) (*domain.State, error) {
	raw, err := e.loader.GetNode(currentState.CurrentNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load node %s: %w", currentState.CurrentNodeID, err)
	}

	node, err := e.parser.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node %s: %w", currentState.CurrentNodeID, err)
	}

	var nextNodeID string

	// Evaluate transitions
	for _, t := range node.Transitions {
		if t.Condition == "" {
			nextNodeID = t.ToNodeID
			break
		}

		if e.evaluator != nil {
			ok, err := e.evaluator(ctx, t.Condition, input)
			if err != nil {
				return nil, fmt.Errorf("condition evaluation failed for '%s': %w", t.Condition, err)
			}
			if ok {
				nextNodeID = t.ToNodeID
				break
			}
		}
	}

	nextState := *currentState
	if len(node.Transitions) == 0 {
		nextState.Terminated = true
	}
	if nextNodeID != "" {
		nextState.CurrentNodeID = nextNodeID
		nextState.History = append(nextState.History, nextNodeID)
	}

	return &nextState, nil
}

// interpolate replaces {{ key }} with values from memory
func interpolate(text string, memory map[string]any) string {
	if memory == nil {
		return text
	}
	for key, val := range memory {
		placeholder := fmt.Sprintf("{{ %s }}", key)
		// Basic string replacement for now.
		// Ideally we would use a regex to handle spacing variations like {{key}}.
		text = strings.ReplaceAll(text, placeholder, fmt.Sprint(val))
	}
	return text
}

// Inspect returns a structured view of the entire graph by walking all nodes.
func (e *Engine) Inspect() ([]domain.Node, error) {
	nodeIDs, err := e.loader.ListNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := make([]domain.Node, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		raw, err := e.loader.GetNode(id)
		if err != nil {
			// Warn but continue? Or fail? Fail is safer for now.
			return nil, fmt.Errorf("failed to load node %s: %w", id, err)
		}
		node, err := e.parser.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node %s: %w", id, err)
		}
		nodes = append(nodes, *node)
	}
	return nodes, nil
}
