package kms

import (
	"context"
	"errors"
)

var (
	ErrInvalidKeySize = errors.New("invalid key size: expected 32 bytes")
	ErrKMSUnavailable = errors.New("kms provider unavailable")
)

// KMS defines the interface for Key Management Service providers.
type KMS interface {
	// Encrypt encrypts plaintext bytes using the KMS provider key.
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	// Decrypt decrypts ciphertext bytes using the KMS provider key.
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
	// Name returns the identifier of the KMS provider (e.g. "local", "aws-kms", "gcp-kms", "vault").
	Name() string
}
