package middleware_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/persistence/middleware"
)

func TestPIIMiddleware_Masking(t *testing.T) {
	// Setup
	underlyingStore := NewMockStore()
	// Mask keys containing "password" or "ssn"
	mw := middleware.NewPIIMiddleware([]string{"password", "ssn"})
	secureStore := mw(underlyingStore)

	ctx := context.Background()
	sessionID := "pii-session"
	state := domain.NewState(sessionID, "start")

	// Populate with mixed data
	state.Context["username"] = "jdoe"
	state.Context["user_password"] = "secret123"
	state.Context["details"] = map[string]any{
		"address":    "123 St",
		"ssn_number": "999-99-9999",
	}
	state.Context["safe_data"] = "public"

	// 1. Save
	if err := secureStore.Save(ctx, sessionID, state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify In-Memory State is NOT MODIFIED (Immutability check)
	if state.Context["user_password"] != "secret123" {
		t.Error("Middleware modified original state in memory!")
	}

	// 2. Load from Underlying Store (Should be masked)
	storedState, err := underlyingStore.Load(ctx, sessionID)
	if err != nil {
		t.Fatalf("Underlying load failed: %v", err)
	}

	// Check masking
	if storedState.Context["username"] != "jdoe" {
		t.Error("Username shouldn't be masked")
	}
	if storedState.Context["user_password"] != "***" {
		t.Errorf("Password should be masked, got: %v", storedState.Context["user_password"])
	}

	details := storedState.Context["details"].(map[string]any)
	if details["ssn_number"] != "***" {
		t.Errorf("Nested SSN should be masked, got: %v", details["ssn_number"])
	}
}
