package domain

// Field constants for mapstructure and JSON standardization.
const (
	// KeyIdempotency is the metadata key used to store the deterministic idempotency key.
	// It is also the JSON field name in the ToolCall struct.
	KeyIdempotency = "idempotency_key"

	// Signal constants representing global events.
	SignalInterrupt = "interrupt" // CTRL+C or explicit cancellation
	SignalShutdown  = "shutdown"  // System termination request (SIGTERM)
	SignalTimeout   = "timeout"   // Node execution deadline exceeded
)
