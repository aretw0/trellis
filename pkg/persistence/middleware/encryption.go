package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// EncryptionConfig holds the keys for encryption and decryption.
type EncryptionConfig struct {
	// ActiveKey is the key used for encrypting new data.
	// Must be 32 bytes for AES-256.
	ActiveKey []byte

	// FallbackKeys is a list of old keys to try when decryption fails.
	// This enables zero-downtime key rotation.
	FallbackKeys [][]byte
}

type encryptionMiddleware struct {
	next   ports.StateStore
	config EncryptionConfig
}

// NewEncryptionMiddleware creates a middleware that encrypts state using AES-GCM (Envelope Encryption)
func NewEncryptionMiddleware(config EncryptionConfig) Middleware {
	if len(config.ActiveKey) != 32 {
		panic("active key must be 32 bytes (AES-256)")
	}
	return func(next ports.StateStore) ports.StateStore {
		return &encryptionMiddleware{
			next:   next,
			config: config,
		}
	}
}

func (m *encryptionMiddleware) Save(ctx context.Context, sessionID string, state *domain.State) error {
	// 1. Serialize real state
	plainText, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// 2. Encrypt
	ciphertext, err := encrypt(plainText, m.config.ActiveKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt state: %w", err)
	}

	// 3. Create envelope
	// We create an opaque envelope state that hides all execution details.
	envelope := domain.NewState("encrypted")
	envelope.Status = state.Status // We might want to expose status for monitoring, but content is hidden.
	envelope.Context = map[string]any{
		"__encrypted__": base64.StdEncoding.EncodeToString(ciphertext),
	}
	// Explicitly clear sensitive fields that NewState might have populated
	envelope.History = nil
	envelope.SystemContext = nil

	return m.next.Save(ctx, sessionID, envelope)
}

func (m *encryptionMiddleware) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	// 1. Load envelope
	envelope, err := m.next.Load(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 2. Extract ciphertext
	encryptedStr, ok := envelope.Context["__encrypted__"].(string)
	if !ok {
		// If the state is not an envelope (e.g. migration from non-encrypted),
		// we could verify if it looks like a normal state.
		// For strict security, we should fail or return the state as is only if intended.
		// For now, let's assume if it has no encrypted blob, it's either an error or a plain state.
		// To be safe/strict, if we configured encryption, we expect encryption.
		// BUT to support enabling encryption on existing sessions, we might fail here.
		// Let's fail secure.
		return nil, errors.New("state is missing encrypted data envelope")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext base64: %w", err)
	}

	// 3. Decrypt (Try Active, then Fallback)
	plainText, err := decryptWithRotation(ciphertext, m.config.ActiveKey, m.config.FallbackKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt state: %w", err)
	}

	// 4. Deserialize
	var realState domain.State
	if err := json.Unmarshal(plainText, &realState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted state: %w", err)
	}

	return &realState, nil
}

func (m *encryptionMiddleware) Delete(ctx context.Context, sessionID string) error {
	return m.next.Delete(ctx, sessionID)
}

func (m *encryptionMiddleware) List(ctx context.Context) ([]string, error) {
	return m.next.List(ctx)
}

// Helpers

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptWithRotation(ciphertext []byte, activeKey []byte, fallbackKeys [][]byte) ([]byte, error) {
	// Try active key first
	if plain, err := decrypt(ciphertext, activeKey); err == nil {
		return plain, nil
	}

	// Try fallbacks in order
	for _, key := range fallbackKeys {
		if plain, err := decrypt(ciphertext, key); err == nil {
			return plain, nil
		}
	}

	return nil, errors.New("decryption failed with all available keys")
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertextBytes := ciphertext[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertextBytes, nil)
}
