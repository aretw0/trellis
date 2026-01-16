package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aretw0/trellis/pkg/domain"
)

// FileStore implements ports.StateStore using the local filesystem.
// It stores sessions as JSON files in a configured directory.
type FileStore struct {
	BasePath string
}

// NewFileStore creates a new FileStore with the given base path.
// If basePath is empty, it defaults to ".trellis/sessions".
func NewFileStore(basePath string) *FileStore {
	if basePath == "" {
		basePath = filepath.Join(".trellis", "sessions")
	}
	return &FileStore{BasePath: basePath}
}

// Save persists the session state to a JSON file.
func (f *FileStore) Save(ctx context.Context, sessionID string, state *domain.State) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID cannot be empty")
	}

	// Ensure directory exists
	if err := os.MkdirAll(f.BasePath, 0755); err != nil {
		return fmt.Errorf("failed to ensure session directory: %w", err)
	}

	filePath := filepath.Join(f.BasePath, sessionID+".json")

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file (Atomic write could be an improvement, but standard write is fine for v0.6)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load retrieves the session state from a JSON file.
func (f *FileStore) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID cannot be empty")
	}

	filePath := filepath.Join(f.BasePath, sessionID+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var state domain.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session state: %w", err)
	}

	return &state, nil
}

// Delete removes the session file.
func (f *FileStore) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID cannot be empty")
	}

	filePath := filepath.Join(f.BasePath, sessionID+".json")

	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

// List returns all active session IDs.
func (f *FileStore) List(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(f.BasePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessions []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Remove .json extension
			name := entry.Name()
			id := name[:len(name)-len(".json")]
			sessions = append(sessions, id)
		}
	}

	return sessions, nil
}
