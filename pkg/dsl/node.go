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

// Build returns the underlying domain.Node.
// This is primarily used by the Builder, but exposed for advanced usage.
func (n *NodeBuilder) Build() domain.Node {
	return n.node
}
