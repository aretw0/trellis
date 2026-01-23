package domain

// Field constants for mapstructure and JSON standardization.
const (
	// KeyIdempotency is the metadata key used to store the deterministic idempotency key.
	// It is also the JSON field name in the ToolCall struct.
	KeyIdempotency = "idempotency_key"
)
