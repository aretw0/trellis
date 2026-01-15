package domain

import (
	"context"
	"time"
)

// EventType defines the category of the event.
type EventType string

const (
	EventNodeEnter  EventType = "node_enter"
	EventNodeLeave  EventType = "node_leave"
	EventToolCall   EventType = "tool_call"
	EventToolReturn EventType = "tool_return"
)

// EventBase contains common fields for all events.
type EventBase struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
	StateID   string    `json:"state_id"` // Optional execution ID/Correlation ID? For now just keep it simple.
}

// NodeEvent represents entry or exit from a node.
type NodeEvent struct {
	EventBase
	NodeID   string `json:"node_id"`
	NodeType string `json:"node_type"`
}

// ToolEvent represents a tool execution.
type ToolEvent struct {
	EventBase
	NodeID   string `json:"node_id"`
	ToolName string `json:"tool_name"`
	Input    any    `json:"input,omitempty"`
	Output   any    `json:"output,omitempty"`
	IsError  bool   `json:"is_error,omitempty"`
}

// LifecycleHooks defines callbacks for engine observability.
type LifecycleHooks struct {
	OnNodeEnter  func(context.Context, *NodeEvent)
	OnNodeLeave  func(context.Context, *NodeEvent)
	OnToolCall   func(context.Context, *ToolEvent)
	OnToolReturn func(context.Context, *ToolEvent)
}
