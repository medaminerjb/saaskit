package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// KeyPair holds a private/public key pair for JWT signing.
type KeyPair struct {
	PrivateKey crypto.Signer
	PublicKey  crypto.PublicKey
	Algorithm  string // RS256, ES256, EdDSA
	KeyID      string // kid for JWKS
}

// KeyRing manages active signing keys and passive verification keys for JWKS key rotation.
type KeyRing struct {
	mu              sync.RWMutex
	ActiveKey       *KeyPair
	VerificationMap map[string]*KeyPair
}

// NewKeyRing initializes a KeyRing with an active key and optional retired keys.
func NewKeyRing(active *KeyPair, retired ...*KeyPair) *KeyRing {
	kr := &KeyRing{
		ActiveKey:       active,
		VerificationMap: make(map[string]*KeyPair),
	}
	if active != nil {
		kr.VerificationMap[active.KeyID] = active
	}
	for _, k := range retired {
		if k != nil {
			kr.VerificationMap[k.KeyID] = k
		}
	}
	return kr
}

// GetActive returns the current active signing key.
func (kr *KeyRing) GetActive() *KeyPair {
	kr.mu.RLock()
	defer kr.mu.RUnlock()
	return kr.ActiveKey
}

// GetKey returns a key by KeyID (kid) for token verification.
func (kr *KeyRing) GetKey(kid string) (*KeyPair, bool) {
	kr.mu.RLock()
	defer kr.mu.RUnlock()
	kp, ok := kr.VerificationMap[kid]
	return kp, ok
}

// AllKeys returns all keypairs (active and retired) currently in the key ring.
func (kr *KeyRing) AllKeys() []*KeyPair {
	kr.mu.RLock()
	defer kr.mu.RUnlock()
	keys := make([]*KeyPair, 0, len(kr.VerificationMap))
	for _, k := range kr.VerificationMap {
		keys = append(keys, k)
	}
	return keys
}

// Rotate transitions a new KeyPair to become the active key, retiring the previous active key.
func (kr *KeyRing) Rotate(newKey *KeyPair) {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	if kr.ActiveKey != nil {
		kr.VerificationMap[kr.ActiveKey.KeyID] = kr.ActiveKey
	}
	kr.ActiveKey = newKey
	if newKey != nil {
		kr.VerificationMap[newKey.KeyID] = newKey
	}
}

// LoadOrGenerateKeyRing loads active and retired keys from disk into a KeyRing.
func LoadOrGenerateKeyRing(keyPath, algorithm string, isDev bool) (*KeyRing, error) {
	active, err := LoadOrGenerateKeyPair(keyPath, algorithm, isDev)
	if err != nil {
		return nil, err
	}

	kr := NewKeyRing(active)

	// Load retired keys from keyPath/retired if present
	retiredDir := filepath.Join(keyPath, "retired")
	entries, err := os.ReadDir(retiredDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pem") {
				continue
			}
			kid := strings.TrimSuffix(entry.Name(), ".pem")
			privPath := filepath.Join(retiredDir, entry.Name())
			pubPath := filepath.Join(retiredDir, kid+".pub")
			kp, err := loadKeyPairWithKID(privPath, pubPath, algorithm, kid)
			if err == nil {
				kr.VerificationMap[kid] = kp
			}
		}
	}

	return kr, nil
}

// LoadOrGenerateKeyPair loads signing keys from disk.
// In development mode, generates keys if they don't exist.
// In production mode, returns an error if keys are missing.
func LoadOrGenerateKeyPair(keyPath, algorithm string, isDev bool) (*KeyPair, error) {
	privPath := filepath.Join(keyPath, "active.pem")
	pubPath := filepath.Join(keyPath, "active.pub")

	// Try to load existing keys
	kp, err := loadKeyPair(privPath, pubPath, algorithm)
	if err == nil {
		return kp, nil
	}

	if !isDev {
		return nil, fmt.Errorf("signing keys not found at %s — in production, keys must be provisioned manually (see keys/README.md): %w", keyPath, err)
	}

	// Development: auto-generate keys
	fmt.Printf("⚠️  Generating %s signing keys for development (path: %s)\n", algorithm, keyPath)
	if err := os.MkdirAll(keyPath, 0700); err != nil {
		return nil, fmt.Errorf("creating key directory: %w", err)
	}

	return generateKeyPair(privPath, pubPath, algorithm)
}

func loadKeyPair(privPath, pubPath, algorithm string) (*KeyPair, error) {
	return loadKeyPairWithKID(privPath, pubPath, algorithm, "active")
}

func loadKeyPairWithKID(privPath, pubPath, algorithm, kid string) (*KeyPair, error) {
	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", privPath)
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Fallback: try legacy PKCS1 for RSA
		rsaKey, rsaErr := x509.ParsePKCS1PrivateKey(block.Bytes)
		if rsaErr != nil {
			// Fallback: try EC
			ecKey, ecErr := x509.ParseECPrivateKey(block.Bytes)
			if ecErr != nil {
				return nil, fmt.Errorf("parsing private key (tried PKCS8, PKCS1, EC): %w", err)
			}
			privKey = ecKey
		} else {
			privKey = rsaKey
		}
	}

	signer, ok := privKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	return &KeyPair{
		PrivateKey: signer,
		PublicKey:  signer.Public(),
		Algorithm:  algorithm,
		KeyID:      kid,
	}, nil
}

func generateKeyPair(privPath, pubPath, algorithm string) (*KeyPair, error) {
	var privKey crypto.Signer
	var err error

	switch algorithm {
	case "RS256":
		privKey, err = rsa.GenerateKey(rand.Reader, 4096)
	case "ES256":
		privKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "EdDSA":
		_, privKey, err = ed25519.GenerateKey(rand.Reader)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	if err != nil {
		return nil, fmt.Errorf("generating %s key: %w", algorithm, err)
	}

	// Marshal private key to PKCS8
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("marshalling private key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		return nil, fmt.Errorf("writing private key: %w", err)
	}

	// Marshal public key
	pubBytes, err := x509.MarshalPKIXPublicKey(privKey.Public())
	if err != nil {
		return nil, fmt.Errorf("marshalling public key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		return nil, fmt.Errorf("writing public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: privKey,
		PublicKey:  privKey.Public(),
		Algorithm:  algorithm,
		KeyID:      "active",
	}, nil
}
