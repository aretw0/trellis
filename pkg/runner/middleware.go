package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// ToolInterceptor is a middleware that can intercept, modify, or block a tool call.
// It returns true if execution should proceed, or false to block it.
// If blocked, it should return a ToolResult describing the denial (typically an error).
type ToolInterceptor func(ctx context.Context, call domain.ToolCall) (bool, domain.ToolResult, error)

// MultiInterceptor chains multiple interceptors.
func MultiInterceptor(interceptors ...ToolInterceptor) ToolInterceptor {
	return func(ctx context.Context, call domain.ToolCall) (bool, domain.ToolResult, error) {
		for _, interceptor := range interceptors {
			allowed, result, err := interceptor(ctx, call)
			if err != nil {
				return false, domain.ToolResult{}, err // System Error
			}
			if !allowed {
				return false, result, nil // Blocked by policy
			}
		}
		return true, domain.ToolResult{}, nil // All allowed
	}
}

// ConfirmationMiddleware prompts the user via the provided Handler before allowing execution.
// It is "aware" of the IOHandler to use its Input/Output methods, but keeps the policy logic separate.
//
// Note: This leverages the IOHandler.SystemOutput capability to send meta-messages.
// This allows the prompt ("Allow execution?") to be distinct from the flow content.
func ConfirmationMiddleware(handler IOHandler) ToolInterceptor {
	return func(ctx context.Context, call domain.ToolCall) (bool, domain.ToolResult, error) {
		// 1. Construct Actions to show the user (System Message)
		if err := handler.SystemOutput(ctx, fmt.Sprintf("Tool Request: '%s' (ID: %s)\nArgs: %v\nAllow execution?", call.Name, call.ID, call.Args)); err != nil {
			return false, domain.ToolResult{}, err
		}

		// 2. Request Input (separately)
		actions := []domain.ActionRequest{
			{
				Type: domain.ActionRequestInput,
				Payload: domain.InputRequest{
					Type: domain.InputText,
				},
			},
		}

		// 3. Output Request Input
		if _, err := handler.Output(ctx, actions); err != nil {
			return false, domain.ToolResult{}, err
		}

		// 3. Read Response
		input, err := handler.Input(ctx)
		if err != nil {
			return false, domain.ToolResult{}, err
		}

		// 4. Validate
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			return true, domain.ToolResult{}, nil
		}

		return false, domain.ToolResult{
			ID:      call.ID,
			IsError: true,
			Error:   "User denied execution by policy",
		}, nil
	}
}

// AutoApproveMiddleware allows everything.
func AutoApproveMiddleware() ToolInterceptor {
	return func(ctx context.Context, call domain.ToolCall) (bool, domain.ToolResult, error) {
		return true, domain.ToolResult{}, nil
	}
}
