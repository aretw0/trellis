package adapters_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// Ensure FileStore implements StateStore
var _ ports.StateStore = (*adapters.FileStore)(nil)

func TestFileStore_Contract(t *testing.T) {
	// Setup temporary directory for testing
	tempDir, err := os.MkdirTemp("", "trellis_store_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup

	store := adapters.NewFileStore(tempDir)
	ctx := context.Background()

	t.Run("LoadNonExistentSession", func(t *testing.T) {
		_, err := store.Load(ctx, "non-existent")
		if err != domain.ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound, got %v", err)
		}
	})

	t.Run("SaveAndLoadSession", func(t *testing.T) {
		sessionID := "session-1"
		state := domain.NewState("start")
		state.Context["foo"] = "bar"
		// Test serialization of complex types if supported
		state.Context["count"] = 42

		// Save
		if err := store.Save(ctx, sessionID, state); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Load
		loaded, err := store.Load(ctx, sessionID)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded.CurrentNodeID != state.CurrentNodeID {
			t.Errorf("expected NodeID %s, got %s", state.CurrentNodeID, loaded.CurrentNodeID)
		}
		if loaded.Context["foo"] != "bar" {
			t.Errorf("expected Context['foo'] = 'bar', got %v", loaded.Context["foo"])
		}
		// JSON unmarshal numbers as float64 by default unless handle strictly,
		// but loose check is fine here. Ideally we check loose equality.
		if val, ok := loaded.Context["count"].(float64); !ok || val != 42 {
			t.Errorf("expected Context['count'] = 42, got %v (%T)", loaded.Context["count"], loaded.Context["count"])
		}
	})

	t.Run("DeleteSession", func(t *testing.T) {
		sessionID := "session-to-delete"
		state := domain.NewState("start")

		if err := store.Save(ctx, sessionID, state); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify exists on disk
		path := filepath.Join(tempDir, sessionID+".json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("file should exist before delete")
		}

		// Delete
		if err := store.Delete(ctx, sessionID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify gone from disk
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file should not exist after delete")
		}

		// Load should fail
		_, err := store.Load(ctx, sessionID)
		if err != domain.ErrSessionNotFound {
			t.Errorf("expected ErrSessionNotFound after delete, got %v", err)
		}
	})

	t.Run("DeleteNonExistentSession", func(t *testing.T) {
		// Should not error (idempotent)
		err := store.Delete(ctx, "ghost-session")
		if err != nil {
			t.Errorf("Delete of non-existent session should not fail, got %v", err)
		}
	})

	t.Run("ListSessions", func(t *testing.T) {
		// Isolate this test
		listDir, err := os.MkdirTemp("", "trellis_store_list_test")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(listDir)
		listStore := adapters.NewFileStore(listDir)

		// Create a few sessions
		ids := []string{"s1", "s2", "s3"}
		for _, id := range ids {
			if err := listStore.Save(ctx, id, domain.NewState("start")); err != nil {
				t.Fatalf("Save failed: %v", err)
			}
		}

		// Create a garbage file (should be ignored)
		garbagePath := filepath.Join(listDir, "garbage.txt")
		if err := os.WriteFile(garbagePath, []byte("garbage"), 0644); err != nil {
			t.Fatalf("failed to create garbage file: %v", err)
		}

		// List
		list, err := listStore.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(list) != len(ids) {
			t.Errorf("expected %d sessions, got %d", len(ids), len(list))
		}

		// Verify IDs present
		mapped := make(map[string]bool)
		for _, id := range list {
			mapped[id] = true
		}
		for _, id := range ids {
			if !mapped[id] {
				t.Errorf("expected session %s in list", id)
			}
		}
	})
}
