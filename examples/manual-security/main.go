package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/file"
	"github.com/aretw0/trellis/pkg/persistence/middleware"
	"github.com/aretw0/trellis/pkg/ports"
	"github.com/aretw0/trellis/pkg/runner"
	"github.com/aretw0/trellis/pkg/session"
)

func main() {
	// 1. Setup Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// 2. Setup File Store (The "Insecure" Backend)
	wd, _ := os.Getwd()
	flowDir := filepath.Join(wd, "examples", "manual-security")
	fileStore := file.New(flowDir)

	// 3. Setup Encryption Middleware
	// In production, fetch this from ENV or KMS
	secretKey := []byte("01234567890123456789012345678901") // 32 bytes

	encryptionMW := middleware.NewEncryptionMiddleware(middleware.EncryptionConfig{
		ActiveKey: secretKey,
	})

	// 4. Setup PII Middleware
	piiMW := middleware.NewPIIMiddleware([]string{"api_key", "password"})

	// Chain: Encryption(PII(Store))
	// 1. PII Middleware: Masks sensitive data (deep copy).
	// 2. Encryption Middleware: Encrypts the masked state + Envelope.
	// 3. FileStore: Writes to disk.
	var secureStore ports.StateStore = fileStore
	secureStore = encryptionMW(secureStore)
	secureStore = piiMW(secureStore) // PII wraps Encryption (Logic: PII runs first on Save)

	// 5. Setup Manager
	sessionMgr := session.NewManager(secureStore)
	sessionID := fmt.Sprintf("secure-%d", time.Now().Unix())

	// 6. Setup Engine (Let Trellis manage Loam)
	eng, err := trellis.New(
		flowDir,
		trellis.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	// 7. Prepare State
	// Initialize context with sensitive data
	initialState := map[string]any{
		"api_key": "sk-12345-very-secret",
		"user":    "james_bond",
	}

	fmt.Printf("Starting Secure Session: %s\n", sessionID)

	// Load or Create Session (Atomic)
	ctx := context.Background()
	state, err := sessionMgr.LoadOrStart(ctx, sessionID, "start")
	if err != nil {
		panic(err)
	}

	// Update context (simulating input injection)
	state.Context["api_key"] = initialState["api_key"]
	state.Context["password"] = "007"

	// 8. Create and run
	r := runner.NewRunner(
		runner.WithStore(sessionMgr),
		runner.WithSessionID(sessionID),
		runner.WithLogger(logger),
		runner.WithInputHandler(runner.NewTextHandler(os.Stdout)),
		runner.WithEngine(eng),
		runner.WithInitialState(state),
	)

	// Execute
	if err := r.Run(ctx); err != nil {
		panic(err)
	}

	fmt.Println("\n--- Verification ---")
	fmt.Printf("Check file: %s/%s.trace.json\n", flowDir, sessionID)
	fmt.Println("It should be an encrypted envelope.")
}
