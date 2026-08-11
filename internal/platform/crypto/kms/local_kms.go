package kms

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// LocalKMS implements KMS using a local 32-byte master key and AES-256-GCM.
type LocalKMS struct {
	masterKey []byte
}

// NewLocalKMS creates a LocalKMS instance from a 32-byte master key.
func NewLocalKMS(masterKey []byte) (*LocalKMS, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("%w, got %d", ErrInvalidKeySize, len(masterKey))
	}
	keyCopy := make([]byte, 32)
	copy(keyCopy, masterKey)
	return &LocalKMS{masterKey: keyCopy}, nil
}

// NewLocalKMSFromHex creates a LocalKMS instance from a 64-character hex string.
func NewLocalKMSFromHex(hexKey string) (*LocalKMS, error) {
	raw, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decoding hex master key: %w", err)
	}
	return NewLocalKMS(raw)
}

func (k *LocalKMS) Name() string {
	return "local"
}

func (k *LocalKMS) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.masterKey)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (k *LocalKMS) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.masterKey)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	payload := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, payload, nil)
}
