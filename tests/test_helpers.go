package tests

import (
	"context"

	"github.com/aretw0/trellis/pkg/domain"
)

// ChannelHandler is a mock handler for testing
type ChannelHandler struct {
	InputCh chan string
}

func (h *ChannelHandler) Output(ctx context.Context, requests []domain.ActionRequest) (bool, error) {
	for _, req := range requests {
		if req.Type == domain.ActionRequestInput {
			return true, nil
		}
	}
	return false, nil
}

func (h *ChannelHandler) Input(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case s := <-h.InputCh:
		return s, nil
	}
}

func (h *ChannelHandler) SystemOutput(ctx context.Context, msg string) error {
	return nil
}

func (h *ChannelHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	return domain.ToolResult{Result: "mock"}, nil
}

func (h *ChannelHandler) Signal(ctx context.Context, name string, args map[string]any) error {
	return nil
}
