package schema

import "fmt"

// ValidationError represents a single field validation failure.
type ValidationError struct {
	Key    string // Field name
	Reason string // Human-readable reason for failure
	Value  any    // The value that failed validation
}

func (e *ValidationError) Error() string {
	if e.Value == nil {
		return fmt.Sprintf("field %q: %s", e.Key, e.Reason)
	}
	return fmt.Sprintf("field %q: %s (got %T)", e.Key, e.Reason, e.Value)
}

// AggregateError represents multiple validation failures.
type AggregateError struct {
	Errors []error
}

func (e *AggregateError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	msg := fmt.Sprintf("%d validation errors:\n", len(e.Errors))
	for i, err := range e.Errors {
		msg += fmt.Sprintf("  %d. %s\n", i+1, err.Error())
	}
	return msg
}

// ValidationErrors returns all validation errors if err is an AggregateError.
// Otherwise returns nil.
func ValidationErrors(err error) []error {
	if aggr, ok := err.(*AggregateError); ok {
		return aggr.Errors
	}
	return nil
}
