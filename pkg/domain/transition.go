package domain

// Transition defines a rule to move from one node to another.
type Transition struct {
	FromNodeID string
	ToNodeID   string

	// Condition is a simple expression string that must evaluate to true
	// for this transition to be valid. e.g., "user_age >= 18"
	// If empty, it's considered an "always" transition (default).
	Condition string
}
