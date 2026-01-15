package runtime

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/aretw0/trellis/internal/compiler"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// ConditionEvaluator is a function that determines if a transition condition is met.
type ConditionEvaluator func(ctx context.Context, condition string, input any) (bool, error)

// Engine is the core state machine runner.
type Engine struct {
	loader       ports.GraphLoader
	parser       *compiler.Parser
	evaluator    ConditionEvaluator
	interpolator Interpolator
}

// DefaultEvaluator implements the basic "condition: input == 'value'" logic.
func DefaultEvaluator(ctx context.Context, condition string, input any) (bool, error) {
	// For backward compatibility and simplicity in string matching,
	// we coerce the input to string for the default evaluator.
	// Users needing complex object evaluation should provide a custom evaluator.
	strInput := fmt.Sprintf("%v", input)
	if condition == "" {
		return true, nil
	}
	// Simple input matching "input == 'yes'"
	if strings.Contains(condition, "input ==") {
		parts := strings.Split(condition, "==")
		if len(parts) == 2 {
			expected := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
			// Case-insensitive matching
			if strings.EqualFold(strings.TrimSpace(strInput), expected) {
				return true, nil
			}
		}
	}
	return false, nil
}

// Interpolator is a function that replaces variables in a string with values from data.
type Interpolator func(ctx context.Context, templateStr string, data any) (string, error)

// DefaultInterpolator uses Go's text/template logic.
func DefaultInterpolator(ctx context.Context, templateStr string, data any) (string, error) {
	// Fast path: no template tokens
	if !strings.Contains(templateStr, "{{") {
		return templateStr, nil
	}

	tmpl, err := template.New("node").Parse(templateStr)
	if err != nil {
		// Fallback: return raw string if parse fails, or error?
		// For robustness in text UIs, maybe returning error is better so dev sees mistake.
		return "", fmt.Errorf("invalid template '%s': %w", templateStr, err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return sb.String(), nil
}

// LegacyInterpolator implements the simple "strings.ReplaceAll" logic for backward compatibility.
func LegacyInterpolator(ctx context.Context, templateStr string, data any) (string, error) {
	ctxMap, ok := data.(map[string]any)
	if !ok || ctxMap == nil {
		return templateStr, nil
	}

	text := templateStr
	for key, val := range ctxMap {
		placeholder := fmt.Sprintf("{{ %s }}", key)
		// Basic string replacement compatible with previous version
		text = strings.ReplaceAll(text, placeholder, fmt.Sprint(val))
	}
	return text, nil
}

// NewEngine creates a new engine with dependencies.
// The engine is immutable after creation.
// interpolator is optional; if nil, DefaultInterpolator (Standard Go Templates) is used.
func NewEngine(loader ports.GraphLoader, evaluator ConditionEvaluator, interpolator Interpolator) *Engine {
	if evaluator == nil {
		evaluator = DefaultEvaluator
	}
	if interpolator == nil {
		interpolator = DefaultInterpolator
	}
	return &Engine{
		loader:       loader,
		parser:       compiler.NewParser(),
		evaluator:    evaluator,
		interpolator: interpolator,
	}
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
	if node.Type == domain.NodeTypeText || node.Type == domain.NodeTypeQuestion {
		text := string(node.Content)

		// Apply Interpolation
		if e.interpolator != nil {
			// Merge SystemContext for Interpolation
			data := make(map[string]any)
			for k, v := range currentState.Context {
				data[k] = v
			}
			data["sys"] = currentState.SystemContext

			interpolated, err := e.interpolator(ctx, text, data)
			if err != nil {
				// If interpolation fails, we currently error out.
				// Alternative: log error and return original text.
				return nil, false, fmt.Errorf("rendering failed during interpolation: %w", err)
			}
			text = interpolated
		}

		actions = append(actions, domain.ActionRequest{
			Type:    domain.ActionRenderContent,
			Payload: text,
		})
	}

	// Calculate Input Request
	// We only request input if explicitly configured (wait: true, type: question, or input_type set)
	needsInput := node.Wait || node.Type == domain.NodeTypeQuestion || node.InputType != ""

	if needsInput {
		inputType := domain.InputType(node.InputType)
		// Default to Text input if not specified but input is required
		if inputType == "" {
			inputType = domain.InputText
		}

		actions = append(actions, domain.ActionRequest{
			Type: domain.ActionRequestInput,
			Payload: domain.InputRequest{
				Type:    inputType,
				Options: node.InputOptions,
				Default: node.InputDefault,
			},
		})
	}

	// Calculate Tool Request
	if node.Type == domain.NodeTypeTool {
		if node.ToolCall == nil {
			// Fallback if ToolCall is missing in struct but Type is Tool
			return nil, false, fmt.Errorf("node %s is type 'tool' but missing tool_call definition", node.ID)
		}
		// Propagate Node Metadata to Tool Call
		// This enables Middleware to see "confirm_msg" etc.
		call := *node.ToolCall
		if node.Metadata != nil {
			call.Metadata = make(map[string]string)
			for k, v := range node.Metadata {
				call.Metadata[k] = v
			}
		}

		// Deep Interpolation for Tool Args
		if e.interpolator != nil && len(call.Args) > 0 {
			// Merge SystemContext for Interpolation (Reusing logic? We should extract this if it grows)
			data := make(map[string]any)
			for k, v := range currentState.Context {
				data[k] = v
			}
			data["sys"] = currentState.SystemContext

			interpolatedArgs := make(map[string]any)
			for k, v := range call.Args {
				// We only interpolate strings
				if strVal, ok := v.(string); ok && strings.Contains(strVal, "{{") {
					interpolated, err := e.interpolator(ctx, strVal, data)
					if err != nil {
						return nil, false, fmt.Errorf("failed to interpolate tool arg '%s': %w", k, err)
					}
					interpolatedArgs[k] = interpolated
				} else {
					interpolatedArgs[k] = v
				}
			}
			call.Args = interpolatedArgs
		}

		actions = append(actions, domain.ActionRequest{
			Type:    domain.ActionCallTool,
			Payload: call,
		})
	}

	isTerminal := len(node.Transitions) == 0
	return actions, isTerminal, nil
}

// Navigate determines the next state based on input.
// Input can be a string (user text) or a domain.ToolResult (side-effect result).
func (e *Engine) Navigate(ctx context.Context, currentState *domain.State, input any) (*domain.State, error) {
	// 1. Handle State: WaitingForTool
	if currentState.Status == domain.StatusWaitingForTool {
		result, ok := input.(domain.ToolResult)
		if !ok {
			return nil, fmt.Errorf("expected ToolResult input when in WaitingForTool status")
		}
		if result.ID != currentState.PendingToolCall {
			return nil, fmt.Errorf("tool result ID %s does not match pending call %s", result.ID, currentState.PendingToolCall)
		}

		// Handle Tool Error (Phase 0: Safety Check)
		if result.IsError {
			// Load current node to check for OnError handler
			raw, err := e.loader.GetNode(currentState.CurrentNodeID)
			if err != nil {
				return nil, fmt.Errorf("failed to load node %s during error handling: %w", currentState.CurrentNodeID, err)
			}
			node, err := e.parser.Parse(raw)
			if err != nil {
				return nil, fmt.Errorf("failed to parse node %s during error handling: %w", currentState.CurrentNodeID, err)
			}

			if node.OnError != "" {
				// Transition to Error Handler
				// We skip applyInput (Context Update) to avoid polluting state with error.

				// Initialize next state with clean context copy
				nextState := *currentState
				nextState.Context = make(map[string]any)
				for k, v := range currentState.Context {
					nextState.Context[k] = v
				}
				nextState.SystemContext = make(map[string]any)
				for k, v := range currentState.SystemContext {
					nextState.SystemContext[k] = v
				}

				// Reset status setup
				nextState.Status = domain.StatusActive
				nextState.PendingToolCall = ""

				return e.transitionTo(&nextState, node.OnError)
			}

			// Fail Fast: If no handler is defined, we stop execution to prevent context poisoning.
			// This enforces explicit error handling or tool reliability.
			return nil, &UnhandledToolError{
				NodeID:   node.ID,
				ToolName: result.ID,
				Cause:    result.Result,
			}
		}

		// Resume execution:
		// We pass the raw result to navigateInternal.
		// 1. applyInput: Captures the result (any) into Context if SaveTo is set.
		// 2. resolveTransition: Evaluates conditions against the result.
		//    Note: DefaultEvaluator coerces to string for checking, ensuring compatibility with existing flows.

		// Create a clean "Active" state to proceed with regular logic
		resumedState := *currentState
		resumedState.Status = domain.StatusActive
		resumedState.PendingToolCall = ""

		return e.navigateInternal(ctx, &resumedState, result.Result)
	}

	// 2. Handle State: Active (Standard Input)
	// Input is already any, so we just pass it through.
	return e.navigateInternal(ctx, currentState, input)
}

// navigateInternal contains the core transition logic (Node loading + Condition eval)
func (e *Engine) navigateInternal(ctx context.Context, currentState *domain.State, input any) (*domain.State, error) {
	raw, err := e.loader.GetNode(currentState.CurrentNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load node %s: %w", currentState.CurrentNodeID, err)
	}

	node, err := e.parser.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node %s: %w", currentState.CurrentNodeID, err)
	}

	// 1. Update Phase: Apply Input to Context
	nextState, err := e.applyInput(currentState, node, input)
	if err != nil {
		return nil, err
	}

	var nextNodeID string

	// 2. Resolve Phase: Evaluate Transitions
	nextNodeID, err = e.resolveTransition(ctx, node, input)
	if err != nil {
		return nil, err
	}

	// Default to whatever the current status was (usually Active if calling internal)
	// But if we fail to transition, we stay in the same state?
	// If no transition found, term?
	if len(node.Transitions) == 0 {
		nextState.Status = domain.StatusTerminated
		nextState.Terminated = true // Deprecated
	}

	if nextNodeID != "" {
		return e.transitionTo(nextState, nextNodeID)
	}

	return nextState, nil
}

// applyInput handles the Update Phase: Creates new state and applies SaveTo logic.
func (e *Engine) applyInput(currentState *domain.State, node *domain.Node, input any) (*domain.State, error) {
	// Initialize next state with a copy of context to support SaveTo
	nextState := *currentState
	nextState.Context = make(map[string]any)
	if currentState.Context != nil {
		for k, v := range currentState.Context {
			nextState.Context[k] = v
		}
	}
	// Copy SystemContext (Host-controlled, but safe to propagate)
	nextState.SystemContext = make(map[string]any)
	if currentState.SystemContext != nil {
		for k, v := range currentState.SystemContext {
			nextState.SystemContext[k] = v
		}
	}

	// Handle Data Binding (SaveTo)
	if node.SaveTo != "" {
		// Validating Namespace
		if node.SaveTo == "sys" || strings.HasPrefix(node.SaveTo, "sys.") {
			return nil, fmt.Errorf("security violation: cannot save to reserved namespace 'sys' in node %s", node.ID)
		}
		nextState.Context[node.SaveTo] = input
	}
	return &nextState, nil
}

// resolveTransition handles the Resolve Phase: Evaluates conditions to find the next node.
func (e *Engine) resolveTransition(ctx context.Context, node *domain.Node, input any) (string, error) {
	for _, t := range node.Transitions {
		if t.Condition == "" {
			return t.ToNodeID, nil
		}

		if e.evaluator != nil {
			ok, err := e.evaluator(ctx, t.Condition, input)
			if err != nil {
				return "", fmt.Errorf("condition evaluation failed for '%s': %w", t.Condition, err)
			}
			if ok {
				return t.ToNodeID, nil
			}
		}
	}
	return "", nil
}

// transitionTo handles the mechanics of moving the state to a new node ID.
func (e *Engine) transitionTo(nextState *domain.State, nextNodeID string) (*domain.State, error) {
	// Update State to new Node
	nextState.CurrentNodeID = nextNodeID
	nextState.History = append(nextState.History, nextNodeID)
	nextState.Status = domain.StatusActive // Default active

	// Check Next Node Type to set Status eagerly
	nextRaw, err := e.loader.GetNode(nextNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load next node %s: %w", nextNodeID, err)
	}
	nextNode, err := e.parser.Parse(nextRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse next node %s: %w", nextNodeID, err)
	}

	if nextNode.Type == domain.NodeTypeTool {
		nextState.Status = domain.StatusWaitingForTool
		if nextNode.ToolCall != nil {
			nextState.PendingToolCall = nextNode.ToolCall.ID
		}
	}

	return nextState, nil
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

// UnhandledToolError represents a failure to handle a tool error.
// It provides structured context for debugging.
type UnhandledToolError struct {
	NodeID   string
	ToolName string
	Cause    any
}

func (e *UnhandledToolError) Error() string {
	return fmt.Sprintf(
		"Tool '%s' (Node '%s') failed with: '%v'. Execution halted because no 'on_error' handler is defined. Fix: Add 'on_error: <node_id>' to node '%s'.",
		e.ToolName, e.NodeID, e.Cause, e.NodeID,
	)
}
