package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/adapters"
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
	flowDir := filepath.Join(wd, "examples", "secure-storage")
	fileStore := adapters.NewFileStore(flowDir)

	// 3. Setup Encryption Middleware
	// In production, fetch this from ENV or KMS
	secretKey := []byte("01234567890123456789012345678901") // 32 bytes

	encryptionMW := middleware.NewEncryptionMiddleware(middleware.EncryptionConfig{
		ActiveKey: secretKey,
	})

	// 4. Setup PII Middleware
	// We want to hide "api_key" and "pass" in the storage, even if encryption wasn't there (defense in depth)
	// But note: Encryption hides EVERYTHING. PII Middleware masks keys inside the state BEFORE encryption.
	// So if you look at the trace file, it will be encrypted blob.
	// If you decrypt it, the PII fields will be "***".
	piiMW := middleware.NewPIIMiddleware([]string{"api_key", "password"})

	// Chain: Store <- Encryption <- PII <- SessionManager
	// The Manager calls Save() on PII, which calls Save() on Encryption, which calls Save() on FileStore.
	// Wait, PII masks the state. Encryption Encrypts the state.
	// We want PII Masking to happen FIRST?
	// If Encryption protects everything, PII masking is redundant for security of the file,
	// BUT useful if we want to ensure that even if the key leaks, the PII is gone?
	// Or maybe for logging?
	// Let's chain: Encryption(PII(Store))?
	// secureStore = Encryption(PII(FileStore))
	// Save(state) -> Encryption.Save(state) -> encrypts -> PII.Save(encryptedState)?
	// NO. Encrypted state has NO keys except "__encrypted__".
	//
	// Correct Chain:
	// We want to Mask PII -> Encrypt -> Store
	//
	// Middleware wrapping:
	// Store = Encryption(Store)
	// Store = PII(Store) // This wraps Encryption!
	//
	// Call Chain:
	// PII.Save(state) -> Masks State -> calls Next.Save(maskedState)
	// Next is Encryption.Save(maskedState) -> Encrypts Masked State -> calls Next.Save(envelope)
	// Next is FileStore.Save(envelope) -> Writes to disk.

	var secureStore ports.StateStore = fileStore
	secureStore = encryptionMW(secureStore)
	secureStore = piiMW(secureStore)

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

	// 7. Run
	r := runner.NewRunner(
		runner.WithStore(sessionMgr),
		runner.WithLogger(logger),
		runner.WithInputHandler(runner.NewTextHandler(os.Stdin, os.Stdout)),
	)
	r.SessionID = sessionID // Important: Set SessionID on runner for persistence

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

	// Execute
	if _, err := r.Run(ctx, eng, state); err != nil {
		panic(err)
	}

	fmt.Println("\n--- Verification ---")
	fmt.Printf("Check file: %s/%s.trace.json\n", flowDir, sessionID)
	fmt.Println("It should be an encrypted envelope.")
}
