package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aretw0/trellis/pkg/domain"
)

// Store implements ports.StateStore using the local filesystem.
// It stores sessions as JSON files in a configured directory.
type Store struct {
	BasePath string
}

// New creates a new Store with the given base path.
// If basePath is empty, it defaults to ".trellis/sessions".
func New(basePath string) *Store {
	if basePath == "" {
		basePath = filepath.Join(".trellis", "sessions")
	}
	return &Store{BasePath: basePath}
}

// Save persists the session state to a JSON file atomically.
// It writes to a temporary file first, syncs via fsync, and then renames it to the destination.
func (s *Store) Save(ctx context.Context, sessionID string, state *domain.State) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID cannot be empty")
	}

	// Ensure directory exists
	if err := os.MkdirAll(s.BasePath, 0755); err != nil {
		return fmt.Errorf("failed to ensure session directory: %w", err)
	}

	destPath := filepath.Join(s.BasePath, sessionID+".json")

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// 1. Create Temp File
	// we use the same directory to ensure we are on the same filesystem (required for atomic rename)
	tmpFile, err := os.CreateTemp(s.BasePath, "tmp-"+sessionID+"-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Cleanup temp file in case of failure
	defer func() {
		// If we successfully renamed, tmpPath won't exist slightly, or we can check err.
		// But simpler: just try to remove. If it doesn't exist (because renamed), os.Remove returns logic that we can ignore or it handles it.
		// Actually best practice:
		_ = tmpFile.Close()    // Ensure closed
		_ = os.Remove(tmpPath) // Remove if still exists (not renamed)
	}()

	// 2. Write Data
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// 3. Fsync to ensure durability
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to fsync temp file: %w", err)
	}

	// 4. Close File (cannot rename open file on Windows)
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// 5. Atomic Rename
	// On Windows, os.Rename fails if dest exists. We must remove it first.
	// Check if dest exists
	if _, err := os.Stat(destPath); err == nil {
		// Dest exists, remove it.
		// There is a tiny window here where file is gone before replacement.
		// True atomicity on Windows requires MoveFileEx with MOVEFILE_REPLACE_EXISTING,
		// but that requires syscalls.
		// For v0.6 CLI usage, this "Delete+Rename" window is acceptable compared to "Write causing partial file".
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("failed to remove existing session file for overwrite: %w", err)
		}
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to rename temp file to valid session: %w", err)
	}

	// Success - defer will try to remove tmpPath but it's gone.
	return nil
}

// Load retrieves the session state from a JSON file.
func (s *Store) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID cannot be empty")
	}

	filePath := filepath.Join(s.BasePath, sessionID+".json")

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
func (s *Store) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID cannot be empty")
	}

	filePath := filepath.Join(s.BasePath, sessionID+".json")

	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

// List returns all active session IDs.
func (s *Store) List(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(s.BasePath)
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
