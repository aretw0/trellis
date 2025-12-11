package domain

// Transition defines a rule to move from one node to another.
type Transition struct {
	FromNodeID string `json:"from_node_id,omitempty"`
	ToNodeID   string `json:"to_node_id"`

	// Condition is a simple expression string that must evaluate to true
	// for this transition to be valid. e.g., "user_age >= 18"
	// If empty, it's considered an "always" transition (default).
	Condition string `json:"condition,omitempty"`
}
