package dsl

import "github.com/aretw0/trellis/pkg/domain"

// NodeBuilder provides a fluent API for configuring a node.
type NodeBuilder struct {
	node    domain.Node
	builder *Builder
}

// Text sets the content of the node and marks it as a text node (soft step).
func (n *NodeBuilder) Text(content string) *NodeBuilder {
	n.node.Type = domain.NodeTypeText
	n.node.Content = []byte(content)
	n.node.Wait = false
	return n
}

// Do configures the primary action (tool) for the node.
// It automatically sets the node type to "tool".
func (n *NodeBuilder) Do(name string, args map[string]any) *NodeBuilder {
	n.node.Type = domain.NodeTypeTool
	n.node.Do = &domain.ToolCall{
		Name: name,
		Args: args,
	}
	return n
}

// Undo defines the compensating action (SAGA pattern) to revert this node's effect.
func (n *NodeBuilder) Undo(name string, args map[string]any) *NodeBuilder {
	n.node.Undo = &domain.ToolCall{
		Name: name,
		Args: args,
	}
	return n
}

// Tools defines tools available to the engine within this node's context.
func (n *NodeBuilder) Tools(tools ...domain.Tool) *NodeBuilder {
	n.node.Tools = append(n.node.Tools, tools...)
	return n
}

// Question sets the content of the node and marks it as a question node (hard step).
func (n *NodeBuilder) Question(content string) *NodeBuilder {
	n.node.Type = domain.NodeTypeQuestion
	n.node.Content = []byte(content)
	n.node.Wait = true
	return n
}

// Input configures the input type and options for a question node.
func (n *NodeBuilder) Input(inputType string, options ...string) *NodeBuilder {
	n.node.InputType = inputType
	n.node.InputOptions = options
	return n
}

// SaveTo specifies the context variable to save the input to.
func (n *NodeBuilder) SaveTo(variable string) *NodeBuilder {
	n.node.SaveTo = variable
	return n
}

// Context adds a default context value to the node.
func (n *NodeBuilder) Context(key string, value any) *NodeBuilder {
	if n.node.DefaultContext == nil {
		n.node.DefaultContext = make(map[string]any)
	}
	n.node.DefaultContext[key] = value
	return n
}

// Go adds an unconditional transition to the target node.
func (n *NodeBuilder) Go(target string) *NodeBuilder {
	n.node.Transitions = append(n.node.Transitions, domain.Transition{
		ToNodeID: target,
	})
	return n
}

// Branch adds a conditional transition to the target node.
func (n *NodeBuilder) Branch(condition string, target string) *NodeBuilder {
	n.node.Transitions = append(n.node.Transitions, domain.Transition{
		Condition: condition,
		ToNodeID:  target,
	})
	return n
}

// Error sets the target node for error handling.
func (n *NodeBuilder) Error(target string) *NodeBuilder {
	n.node.OnError = target
	return n
}

// On adds a signal handler to the node.
func (n *NodeBuilder) On(signal string, target string) *NodeBuilder {
	if n.node.OnSignal == nil {
		n.node.OnSignal = make(map[string]string)
	}
	n.node.OnSignal[signal] = target
	return n
}

// Terminal marks the node as a terminal node (end of the flow).
func (n *NodeBuilder) Terminal() *NodeBuilder {
	n.node.Transitions = nil
	return n
}

// Build returns the underlying domain.Node.
// This is primarily used by the Builder, but exposed for advanced usage.
func (n *NodeBuilder) Build() domain.Node {
	return n.node
}
