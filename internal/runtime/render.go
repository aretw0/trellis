package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aretw0/trellis/pkg/domain"
)

// renderContent handles interpolation and formatting of node text.
func (e *Engine) renderContent(ctx context.Context, node *domain.Node, state *domain.State) (string, error) {
	text := string(node.Content)
	if e.interpolator == nil {
		return text, nil
	}

	// Prepare data for interpolation
	data := make(map[string]any)
	for k, v := range state.Context {
		data[k] = v
	}
	data["sys"] = state.SystemContext

	interpolated, err := e.interpolator(ctx, text, data)
	if err != nil {
		return "", fmt.Errorf("rendering failed during interpolation: %w", err)
	}
	return interpolated, nil
}

// renderInputRequest calculates the action for user input based on node config.
func (e *Engine) renderInputRequest(node *domain.Node) (*domain.ActionRequest, error) {
	needsInput := node.Wait || node.Type == domain.NodeTypeQuestion || node.InputType != ""
	if !needsInput {
		return nil, nil
	}

	inputType := domain.InputType(node.InputType)
	if inputType == "" {
		inputType = domain.InputText
	}

	var timeoutDuration time.Duration
	if node.Timeout != "" {
		if d, err := time.ParseDuration(node.Timeout); err == nil {
			timeoutDuration = d
		} else {
			e.logger.Warn("Failed to parse node timeout", "node_id", node.ID, "timeout", node.Timeout, "error", err)
		}
	}

	return &domain.ActionRequest{
		Type: domain.ActionRequestInput,
		Payload: domain.InputRequest{
			Type:    inputType,
			Options: node.InputOptions,
			Default: node.InputDefault,
			Timeout: timeoutDuration,
		},
	}, nil
}

// renderToolCall calculates the action for a side-effect (Do or Undo).
func (e *Engine) renderToolCall(ctx context.Context, node *domain.Node, state *domain.State) (*domain.ActionRequest, error) {
	var toolCallToRender *domain.ToolCall

	// SAGA logic for tool selection
	if state.Status == domain.StatusRollingBack {
		if node.Undo != nil {
			toolCallToRender = node.Undo
		}
	} else if node.Do != nil {
		toolCallToRender = node.Do
	}

	if toolCallToRender == nil {
		// Strict check for tool-type nodes
		if node.Type == domain.NodeTypeTool && state.Status != domain.StatusRollingBack {
			return nil, fmt.Errorf("node %s is type 'tool' but missing tool_call definition", node.ID)
		}
		return nil, nil
	}

	// Clone and enrich call metadata.
	// We must preserve existing metadata from the tool definition (e.g. x-exec).
	call := *toolCallToRender
	if call.Metadata == nil {
		call.Metadata = make(map[string]string)
	} else {
		// Create a copy to avoid mutating the source node's tool definition
		newMeta := make(map[string]string)
		for k, v := range call.Metadata {
			newMeta[k] = v
		}
		call.Metadata = newMeta
	}

	// Merge Node-level metadata overrides
	if node.Metadata != nil {
		for k, v := range node.Metadata {
			call.Metadata[k] = v
		}
	}

	// Idempotency
	key := e.generateIdempotencyKey(state, node.ID, call.Name)
	call.IdempotencyKey = key
	call.Metadata[domain.KeyIdempotency] = key

	// Arg Interpolation
	if e.interpolator != nil && len(call.Args) > 0 {
		data := make(map[string]any)
		for k, v := range state.Context {
			data[k] = v
		}
		data["sys"] = state.SystemContext

		interpolatedArgs := make(map[string]any)
		for k, v := range call.Args {
			if strVal, ok := v.(string); ok && strings.Contains(strVal, "{{") {
				val, err := e.interpolator(ctx, strVal, data)
				if err != nil {
					return nil, fmt.Errorf("failed to interpolate tool arg '%s': %w", k, err)
				}
				interpolatedArgs[k] = val
			} else {
				interpolatedArgs[k] = v
			}
		}
		call.Args = interpolatedArgs
	}

	return &domain.ActionRequest{
		Type:    domain.ActionCallTool,
		Payload: call,
	}, nil
}
