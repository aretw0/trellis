# Security & Privacy Guide

This guide explains how to configure Trellis to handle sensitive data securely using the built-in Middleware Hooks.

## 1. Protecting Sensitive Data (Encryption at Rest)

If your Trellis sessions store sensitive information (API Keys, User PII) and use persistent storage (File, Redis), you should enable **Encryption at Rest**.

Trellis uses an **Envelope Pattern** with **AES-GCM**. The entire session state is encrypted and wrapped in an opaque "envelope" state before being stored.

### Configuration

To enable encryption, wrap your `StateStore` with the `EncryptionMiddleware`. (See [`examples/manual-security`](../../examples/manual-security) for a full runnable example).

```go
import (
    "github.com/aretw0/trellis/pkg/persistence/middleware"
    "github.com/aretw0/trellis/pkg/adapters/file"
)

func main() {
    // 1. Initialize your base store (e.g., file.Store or Redis)
    baseStore := file.NewStore("./sessions")

    // 2. Define your Encryption Config (Keys should come from KMS/Env)
    config := middleware.EncryptionConfig{
        ActiveKey: []byte("your-32-byte-secret-key-12345678"), // AES-256
    }
    
    // 3. Wrap the store
    secureStore := middleware.NewEncryptionMiddleware(config)(baseStore)

    // 4. Use secureStore in your Runner/Manager
    manager := session.NewManager(secureStore)
}
```

### Key Rotation

Trellis supports **Zero-Downtime Key Rotation**. When you rotate encryption keys, you can provide the old keys as fallbacks.

The middleware will:

1. Try to decrypt with `ActiveKey`.
2. If that fails (e.g. state was encrypted with old key), try `FallbackKeys` in order.
3. When saving, it always re-encrypts using the new `ActiveKey`.

```go
config := middleware.EncryptionConfig{
    ActiveKey:    []byte("new-32-byte-secret-key-87654321"),
    FallbackKeys: [][]byte{
        []byte("old-32-byte-secret-key-12345678"),
    },
}
```

This effectively migrates your data lazily as sessions are accessed.

## 2. PII Sanitization (Masking)

If you need to ensure that sensitive fields (like `password`, `ssn`) are **never** written to disk in plaintext—even if not using encryption, or as a secondary defense layer—use the `PIIMiddleware`.

### Configuration

The PII middleware accepts a list of regular expressions. Any key in the Context matching these patterns will have its value replaced by `***`.

```go
piiMiddleware := middleware.NewPIIMiddleware([]string{
    "password", 
    "api_key", 
    "ssn_*",
})

// Chain it: Encryption(PII(Store))
// This means: Mask PII first -> Then Encrypt -> Then Save to Store
cautiousStore := middleware.NewEncryptionMiddleware(encConfig)(
    piiMiddleware(baseStore),
)
```

### Important Caveat

The PII Middleware masks data **before persistence**.

- **In-Memory**: The running Engine retains the original data, so the current execution is unaffected.
- **On Disk**: The data is destroyed (`***`).

**Effect on Resume**: If your application crashes and restarts, it will load the session from disk. Since the sensitive data was masked, the resumed state will contain `***` instead of the real values.

> **Use Case**: Use PII Masking primarily for **Logging compliance** or **Ephemeral sessions** where you prefer to crash rather than leak data. If you need Durable Execution with sensitive data, rely on **Encryption** instead of Masking.
