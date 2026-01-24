package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// handleToolResult processes the outcome of a side-effect.
func (e *Engine) handleToolResult(ctx context.Context, currentState *domain.State, node *domain.Node, result domain.ToolResult) (*domain.State, error) {
	// 1. Policy Denial Handling
	if result.IsDenied {
		e.logger.Debug("tool execution denied", "tool", result.ID, "node", currentState.CurrentNodeID)

		target := node.OnDenied
		if target == "" {
			target = node.OnError
		}
		if target == "" {
			target = e.defaultErrorNodeID
		}

		if target != "" {
			nextState := e.cloneState(currentState)
			nextState.Status = domain.StatusActive
			nextState.PendingToolCall = ""
			return e.transitionTo(nextState, target)
		}

		// Unhandled Denial: Graceful termination
		nextState := e.cloneState(currentState)
		nextState.Status = domain.StatusTerminated
		return nextState, nil
	}

	// 2. Runtime Error Handling
	if result.IsError {
		e.emitToolReturn(ctx, currentState.CurrentNodeID, result.ID, result.Result, true)

		if node.OnError != "" {
			if node.OnError == "rollback" {
				e.emitNodeLeave(ctx, node)
				return e.startRollback(ctx, currentState)
			}

			nextState := e.cloneState(currentState)
			nextState.Status = domain.StatusActive
			nextState.PendingToolCall = ""
			return e.transitionTo(nextState, node.OnError)
		}

		// Global Fallback
		if e.defaultErrorNodeID != "" {
			e.emitNodeLeave(ctx, node)
			nextState := e.cloneState(currentState)
			nextState.Status = domain.StatusActive
			nextState.PendingToolCall = ""
			return e.transitionTo(nextState, e.defaultErrorNodeID)
		}

		// Prepare Error Cause for reporting
		cause := result.Error
		if cause == "" {
			cause = fmt.Sprintf("%v", result.Result)
		}

		return nil, &UnhandledToolError{
			NodeID:   node.ID,
			ToolName: result.ID,
			Cause:    cause,
		}
	}

	// 3. Success: Resume execution
	e.emitToolReturn(ctx, currentState.CurrentNodeID, result.ID, result.Result, false)

	resumedState := e.cloneState(currentState)
	resumedState.Status = domain.StatusActive
	resumedState.PendingToolCall = ""

	return e.navigateInternal(ctx, resumedState, result.Result)
}

// resolveEffectiveInput applies defaults and validations based on node configuration.
func (e *Engine) resolveEffectiveInput(node *domain.Node, input any) (any, error) {
	effectiveInput := input

	isEmpty := false
	switch v := input.(type) {
	case string:
		isEmpty = v == ""
	case nil:
		isEmpty = true
	}

	if node.InputType == "confirm" {
		if isEmpty {
			if node.InputDefault != "" {
				effectiveInput = node.InputDefault
			} else {
				effectiveInput = "yes"
			}
		}

		strVal := fmt.Sprintf("%v", effectiveInput)
		clean := strings.ToLower(strings.TrimSpace(strVal))
		isTruthy := clean == "y" || clean == "yes" || clean == "true" || clean == "1"
		isFalsy := clean == "n" || clean == "no" || clean == "false" || clean == "0"

		if !isTruthy && !isFalsy {
			return nil, fmt.Errorf("invalid confirmation input: '%s' (expected y/n/yes/no)", strVal)
		}

		if isTruthy {
			return "yes", nil
		}
		return "no", nil
	}

	if isEmpty && node.InputDefault != "" {
		return node.InputDefault, nil
	}

	return effectiveInput, nil
}

// resolveNextNodeID evaluates the priority-based transition rules.
func (e *Engine) resolveNextNodeID(ctx context.Context, node *domain.Node, input any) (string, error) {
	// Check for refusal (for on_denied handler synergy)
	isRefusal := false
	switch v := input.(type) {
	case bool:
		isRefusal = !v
	case string:
		clean := strings.ToLower(strings.TrimSpace(v))
		isRefusal = clean == "n" || clean == "no" || clean == "false" || clean == "deny"
	}

	// Priority 1: Conditional Transitions
	for _, t := range node.Transitions {
		if t.Condition != "" && e.evaluator != nil {
			ok, err := e.evaluator(ctx, t.Condition, input)
			if err == nil && ok {
				return t.ToNodeID, nil
			}
		}
	}

	// Priority 2: Policy Handler (on_denied)
	if isRefusal && node.OnDenied != "" {
		return node.OnDenied, nil
	}

	// Priority 3: Unconditional Transitions
	for _, t := range node.Transitions {
		if t.Condition == "" {
			return t.ToNodeID, nil
		}
	}

	return "", nil
}

// cloneState creates a shallow copy of the state with deep-copied contexts for safe mutation.
func (e *Engine) cloneState(src *domain.State) *domain.State {
	if src == nil {
		return nil
	}
	next := *src
	next.Context = make(map[string]any)
	for k, v := range src.Context {
		next.Context[k] = v
	}
	next.SystemContext = make(map[string]any)
	for k, v := range src.SystemContext {
		next.SystemContext[k] = v
	}
	return &next
}
