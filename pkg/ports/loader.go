package ports

// GraphLoader defines how the engine retrieves node definitions.
// This allows the storage layer (Loam, FS, Memory) to be decoupled.
type GraphLoader interface {
	// GetNode retrieves the raw definition of a node by ID.
	// It returns the raw bytes (which the compiler will parse) or an error.
	GetNode(id string) ([]byte, error)
}
