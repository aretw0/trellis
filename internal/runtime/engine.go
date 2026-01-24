package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/aretw0/trellis/internal/compiler"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// ConditionEvaluator is a function that determines if a transition condition is met.
type ConditionEvaluator func(ctx context.Context, condition string, input any) (bool, error)

// Engine is the core state machine runner.
type Engine struct {
	loader             ports.GraphLoader
	parser             *compiler.Parser
	evaluator          ConditionEvaluator
	interpolator       Interpolator
	hooks              domain.LifecycleHooks
	entryNodeID        string
	defaultErrorNodeID string
	logger             *slog.Logger
}

// EngineOption allows configuring the engine via functional options.
type EngineOption func(*Engine)

// WithLifecycleHooks registers observability hooks.
func WithLifecycleHooks(hooks domain.LifecycleHooks) EngineOption {
	return func(e *Engine) {
		e.hooks = hooks
	}
}

// WithEntryNode configures the initial node ID (default: "start").
func WithEntryNode(nodeID string) EngineOption {
	return func(e *Engine) {
		e.entryNodeID = nodeID
	}
}

// WithLogger configures the structured logger for the engine.
func WithLogger(logger *slog.Logger) EngineOption {
	return func(e *Engine) {
		e.logger = logger
	}
}

// WithDefaultErrorNode sets a global fallback node for tool errors.
func WithDefaultErrorNode(nodeID string) EngineOption {
	return func(e *Engine) {
		e.defaultErrorNodeID = nodeID
	}
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
func NewEngine(loader ports.GraphLoader, evaluator ConditionEvaluator, interpolator Interpolator, opts ...EngineOption) *Engine {
	if evaluator == nil {
		evaluator = DefaultEvaluator
	}
	if interpolator == nil {
		interpolator = DefaultInterpolator
	}
	e := &Engine{
		loader:       loader,
		parser:       compiler.NewParser(),
		evaluator:    evaluator,
		interpolator: interpolator,
		entryNodeID:  "start",                                        // Default convention
		logger:       slog.New(slog.NewJSONHandler(io.Discard, nil)), // Default No-Op
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) generateIdempotencyKey(state *domain.State, nodeID string, toolName string) string {
	// Key = SessionID + NodeID + HistoryLength (Step Index) + ToolName
	stepIndex := len(state.History)
	raw := fmt.Sprintf("%s:%s:%d:%s", state.SessionID, nodeID, stepIndex, toolName)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// Start creates the initial state and triggers the OnNodeEnter hook.
func (e *Engine) Start(ctx context.Context, sessionID string, initialContext map[string]any) (*domain.State, error) {
	state := domain.NewState(sessionID, e.entryNodeID)
	// Load start node to get defaults and metadata
	var startNode *domain.Node
	raw, err := e.loader.GetNode(e.entryNodeID)
	if err == nil {
		startNode, _ = e.parser.Parse(raw)
	}

	// Apply Defaults if available
	if startNode != nil && startNode.DefaultContext != nil {
		for k, v := range startNode.DefaultContext {
			state.Context[k] = v
		}
	}

	for k, v := range initialContext {
		state.Context[k] = v
	}

	// Determine initial status based on Entry Node
	if startNode != nil && startNode.Do != nil {
		state.Status = domain.StatusWaitingForTool
		state.PendingToolCall = startNode.Do.ID
	}

	// Trigger OnNodeEnter for the start node
	if startNode != nil {
		e.emitNodeEnter(ctx, startNode, e.entryNodeID)
	}

	return state, nil
}

// Render calculates the presentation for the current state.
// It loads the node and generates actions (e.g. print text) but does NOT change state.
// It returns actions, isTerminal (true if no transitions), and error.
func (e *Engine) Render(ctx context.Context, currentState *domain.State) ([]domain.ActionRequest, bool, error) {
	if currentState == nil {
		return nil, false, fmt.Errorf("cannot render nil state")
	}

	raw, err := e.loader.GetNode(currentState.CurrentNodeID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load node %s: %w", currentState.CurrentNodeID, err)
	}

	node, err := e.parser.Parse(raw)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse node %s: %w", currentState.CurrentNodeID, err)
	}

	// Safety: Check for logical violations (e.g. Do + Wait)
	if err := e.validateExecution(node); err != nil {
		return nil, false, err
	}

	// Validate Context Requirements
	if err := e.validateContext(node, currentState); err != nil {
		return nil, false, err
	}

	actions := []domain.ActionRequest{}

	// 1. Render Content (Text/Markdown)
	if node.Type == domain.NodeTypeText || node.Type == domain.NodeTypeQuestion || len(node.Content) > 0 {
		text, err := e.renderContent(ctx, node, currentState)
		if err != nil {
			return nil, false, err
		}
		actions = append(actions, domain.ActionRequest{
			Type:    domain.ActionRenderContent,
			Payload: text,
		})
	}

	// 2. Render Input Request
	inputReq, err := e.renderInputRequest(node)
	if err != nil {
		return nil, false, err
	}
	if inputReq != nil {
		actions = append(actions, *inputReq)
	}

	// 3. Render Tool Call (Side-effect)
	toolCall, err := e.renderToolCall(ctx, node, currentState)
	if err != nil {
		return nil, false, err
	}
	if toolCall != nil {
		actions = append(actions, *toolCall)
		e.emitToolCall(ctx, currentState.CurrentNodeID, toolCall.Payload.(domain.ToolCall))
	}

	// 4. Terminal Logic
	hasStandardTransitions := len(node.Transitions) > 0
	hasSignalTransitions := len(node.OnSignal) > 0
	hasTimeout := node.Timeout != ""
	isTerminal := !hasStandardTransitions && !hasSignalTransitions && !hasTimeout

	return actions, isTerminal, nil
}

func (e *Engine) Navigate(ctx context.Context, currentState *domain.State, input any) (*domain.State, error) {
	if currentState == nil {
		return nil, fmt.Errorf("cannot navigate nil state")
	}

	// 1. Handle State: WaitingForTool (or RollingBack which works similarly for Undo)
	if currentState.Status == domain.StatusWaitingForTool || currentState.Status == domain.StatusRollingBack {
		result, ok := input.(domain.ToolResult)
		if !ok {
			return nil, fmt.Errorf("expected ToolResult input when in WaitingForTool/RollingBack status")
		}
		if result.ID != currentState.PendingToolCall {
			return nil, fmt.Errorf("tool result ID %s does not match pending call %s", result.ID, currentState.PendingToolCall)
		}

		// Handle State: RollingBack (Undo Tool Completed)
		if currentState.Status == domain.StatusRollingBack {
			return e.continueRollback(ctx, currentState, true)
		}

		// Handle Tool Result (Success/Error/Denied)
		raw, err := e.loader.GetNode(currentState.CurrentNodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to load node %s: %w", currentState.CurrentNodeID, err)
		}
		node, err := e.parser.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node %s: %w", currentState.CurrentNodeID, err)
		}

		return e.handleToolResult(ctx, currentState, node, result)
	}

	// 2. Handle State: Active (Standard Input)
	return e.navigateInternal(ctx, currentState, input)
}

// Signal triggers a transition based on a global signal (e.g., "interrupt").
func (e *Engine) Signal(ctx context.Context, currentState *domain.State, signalName string) (*domain.State, error) {
	if currentState == nil {
		return nil, fmt.Errorf("cannot signal nil state")
	}

	raw, err := e.loader.GetNode(currentState.CurrentNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load node %s during signal handling: %w", currentState.CurrentNodeID, err)
	}
	node, err := e.parser.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node %s: %w", currentState.CurrentNodeID, err)
	}

	e.emitNodeLeave(ctx, node)

	targetNodeID, ok := node.OnSignal[signalName]
	if !ok {
		return nil, domain.ErrUnhandledSignal
	}

	// Initialize next state with clean context copy
	nextState := e.cloneState(currentState)
	nextState.Status = domain.StatusActive
	nextState.PendingToolCall = ""

	return e.transitionTo(nextState, targetNodeID)
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

	// Safety: Check for logical violations
	if err := e.validateExecution(node); err != nil {
		return nil, err
	}

	if err := e.validateContext(node, currentState); err != nil {
		return nil, err
	}

	// 0. Build Effective Input (Defaults/Validation)
	effectiveInput, err := e.resolveEffectiveInput(node, input)
	if err != nil {
		return nil, err
	}

	// 1. Update Context (SaveTo)
	nextState, err := e.applyInput(currentState, node, effectiveInput)
	if err != nil {
		return nil, err
	}

	// 2. Resolve Next Node (Priority Logic: Conditional > Denial > Fallback)
	nextNodeID, err := e.resolveNextNodeID(ctx, node, effectiveInput)
	if err != nil {
		return nil, err
	}

	// 3. Process Resulting State
	if strings.EqualFold(nextNodeID, "rollback") {
		e.emitNodeLeave(ctx, node)
		return e.startRollback(ctx, nextState)
	}

	if nextNodeID == "" && len(node.Transitions) == 0 {
		nextState.Status = domain.StatusTerminated
		nextState.Terminated = true
		e.emitNodeLeave(ctx, node)
	}

	if nextNodeID != "" {
		e.emitNodeLeave(ctx, node)
		return e.transitionTo(nextState, nextNodeID)
	}

	return nextState, nil
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

	if nextNode.Do != nil {
		nextState.Status = domain.StatusWaitingForTool
		nextState.PendingToolCall = nextNode.Do.ID
	}

	// Emit Enter Event
	e.emitNodeEnter(context.Background(), nextNode, nextNodeID)

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

// emitNodeLeave emits the OnNodeLeave event if hooks are configured.
func (e *Engine) emitNodeLeave(ctx context.Context, node *domain.Node) {
	if e.hooks.OnNodeLeave != nil {
		e.hooks.OnNodeLeave(ctx, &domain.NodeEvent{
			EventBase: domain.EventBase{
				Timestamp: time.Now(),
				Type:      domain.EventNodeLeave,
			},
			NodeID:   node.ID,
			NodeType: node.Type,
		})
	}
}

func (e *Engine) emitNodeEnter(ctx context.Context, node *domain.Node, nodeID string) {
	if e.hooks.OnNodeEnter != nil {
		e.hooks.OnNodeEnter(ctx, &domain.NodeEvent{
			EventBase: domain.EventBase{
				Timestamp: time.Now(),
				Type:      domain.EventNodeEnter,
			},
			NodeID:   nodeID,
			NodeType: node.Type,
		})
	}
}

func (e *Engine) emitToolCall(ctx context.Context, nodeID string, call domain.ToolCall) {
	if e.hooks.OnToolCall != nil {
		e.hooks.OnToolCall(ctx, &domain.ToolEvent{
			EventBase: domain.EventBase{
				Timestamp: time.Now(),
				Type:      domain.EventToolCall,
			},
			NodeID:   nodeID,
			ToolName: call.Name, // Using Name as generic identifier, or ID? Event struct says ToolName.
			Input:    call.Args,
		})
	}
}

func (e *Engine) emitToolReturn(ctx context.Context, nodeID string, toolName string, output any, isError bool) {
	if e.hooks.OnToolReturn != nil {
		e.hooks.OnToolReturn(ctx, &domain.ToolEvent{
			EventBase: domain.EventBase{
				Timestamp: time.Now(),
				Type:      domain.EventToolReturn,
			},
			NodeID:   nodeID,
			ToolName: toolName,
			Output:   output,
			IsError:  isError,
		})
	}
}
