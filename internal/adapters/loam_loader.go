package adapters

import (
	"fmt"
)

// LoamLoader adapts the Loam library to the Trellis GraphLoader interface.
// Ideally, this would use the real Loam service, but for now we define the structure.
type LoamLoader struct {
	// loamService *loam.Service // To be injected
	RepoPath string
}

// NewLoamLoader creates a new Loam adapter.
func NewLoamLoader(repoPath string) *LoamLoader {
	return &LoamLoader{
		RepoPath: repoPath,
	}
}

// GetNode retrieves a node from the Loam repository.
// Note: In a real implementation, this would look up the file in the git repo,
// parse the content (or pass bytes to a parser), and return the Node.
func (l *LoamLoader) GetNode(id string) ([]byte, error) {
	// TODO: Connect to actual Loam implementation.
	// This is currently a stub to satisfy the interface check.
	return nil, fmt.Errorf("LoamLoader not yet implemented for node: %s", id)
}
