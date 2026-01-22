package middleware_test

import (
	"context"
	"crypto/rand"
	"io"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/persistence/middleware"
)

func generateKey(t *testing.T) []byte {
	k := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		t.Fatal(err)
	}
	return k
}

func TestEncryptionMiddleware_Roundtrip(t *testing.T) {
	// Setup
	underlyingStore := NewMockStore()
	key := generateKey(t)
	mw := middleware.NewEncryptionMiddleware(middleware.EncryptionConfig{ActiveKey: key})
	secureStore := mw(underlyingStore)

	ctx := context.Background()
	sessionID := "test-session"
	originalState := domain.NewState("start")
	originalState.Context["secret"] = "my-secret-sauce"

	// 1. Save
	if err := secureStore.Save(ctx, sessionID, originalState); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 2. Verify Underlying Store directly (Should be encrypted)
	storedState, err := underlyingStore.Load(ctx, sessionID)
	if err != nil {
		t.Fatalf("Underlying load failed: %v", err)
	}
	if val, ok := storedState.Context["secret"]; ok {
		t.Fatalf("Expected secret to be hidden, found: %v", val)
	}
	if _, ok := storedState.Context["__encrypted__"]; !ok {
		t.Fatal("Expected __encrypted__ field in context")
	}

	// 3. Load via Middleware (Should be decrypted)
	loadedState, err := secureStore.Load(ctx, sessionID)
	if err != nil {
		t.Fatalf("Load via middleware failed: %v", err)
	}
	if loadedState.Context["secret"] != "my-secret-sauce" {
		t.Errorf("Expected 'my-secret-sauce', got %v", loadedState.Context["secret"])
	}
}

func TestEncryptionMiddleware_KeyRotation(t *testing.T) {
	// Setup
	underlyingStore := NewMockStore()
	oldKey := generateKey(t)
	newKey := generateKey(t)

	// Create middleware with OLD key to save initial state
	mwOld := middleware.NewEncryptionMiddleware(middleware.EncryptionConfig{ActiveKey: oldKey})
	secureStoreOld := mwOld(underlyingStore)

	ctx := context.Background()
	sessionID := "rotation-session"
	originalState := domain.NewState("start")
	originalState.Context["data"] = "encrypted-with-old-key"

	// 1. Save with OLD key
	if err := secureStoreOld.Save(ctx, sessionID, originalState); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 2. Load with NEW key (Active) + OLD key (Fallback)
	mwNew := middleware.NewEncryptionMiddleware(middleware.EncryptionConfig{
		ActiveKey:    newKey,
		FallbackKeys: [][]byte{oldKey},
	})
	secureStoreNew := mwNew(underlyingStore)

	loadedState, err := secureStoreNew.Load(ctx, sessionID)
	if err != nil {
		t.Fatalf("Load with rotated key failed: %v", err)
	}

	if loadedState.Context["data"] != "encrypted-with-old-key" {
		t.Errorf("Decryption with fallback key key failed")
	}

	// 3. Save again (Should now define with NEW key)
	loadedState.Context["data"] = "encrypted-with-new-key"
	if err := secureStoreNew.Save(ctx, sessionID, loadedState); err != nil {
		t.Fatalf("Save with new key failed: %v", err)
	}

	// 4. Verify we CANNOT load with just OLD key anymore
	_, err = secureStoreOld.Load(ctx, sessionID)
	if err == nil {
		t.Error("Expected failure when loading new-key encryption with old-key middleware")
	}
}

func TestEncryptionMiddleware_InvalidKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid key size")
		}
	}()
	middleware.NewEncryptionMiddleware(middleware.EncryptionConfig{ActiveKey: []byte("short-key")})
}
