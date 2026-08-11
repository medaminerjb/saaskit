package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrGenerateKeyPair_RS256(t *testing.T) {
	tmpDir := t.TempDir()

	kp, err := LoadOrGenerateKeyPair(tmpDir, "RS256", true)
	if err != nil {
		t.Fatalf("LoadOrGenerateKeyPair() error: %v", err)
	}

	if kp.Algorithm != "RS256" {
		t.Errorf("algorithm = %q, want RS256", kp.Algorithm)
	}
	if kp.PrivateKey == nil {
		t.Error("private key should not be nil")
	}
	if kp.PublicKey == nil {
		t.Error("public key should not be nil")
	}

	// Files should exist
	if _, err := os.Stat(filepath.Join(tmpDir, "active.pem")); os.IsNotExist(err) {
		t.Error("private key file should exist")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "active.pub")); os.IsNotExist(err) {
		t.Error("public key file should exist")
	}

	// Loading again should reuse existing keys
	kp2, err := LoadOrGenerateKeyPair(tmpDir, "RS256", true)
	if err != nil {
		t.Fatalf("second LoadOrGenerateKeyPair() error: %v", err)
	}
	if kp2.PrivateKey == nil {
		t.Error("reloaded private key should not be nil")
	}
}

func TestLoadOrGenerateKeyPair_ES256(t *testing.T) {
	tmpDir := t.TempDir()

	kp, err := LoadOrGenerateKeyPair(tmpDir, "ES256", true)
	if err != nil {
		t.Fatalf("LoadOrGenerateKeyPair(ES256) error: %v", err)
	}
	if kp.Algorithm != "ES256" {
		t.Errorf("algorithm = %q, want ES256", kp.Algorithm)
	}
}

func TestLoadOrGenerateKeyPair_EdDSA(t *testing.T) {
	tmpDir := t.TempDir()

	kp, err := LoadOrGenerateKeyPair(tmpDir, "EdDSA", true)
	if err != nil {
		t.Fatalf("LoadOrGenerateKeyPair(EdDSA) error: %v", err)
	}
	if kp.Algorithm != "EdDSA" {
		t.Errorf("algorithm = %q, want EdDSA", kp.Algorithm)
	}
}

func TestLoadOrGenerateKeyPair_ProductionFailsIfMissing(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadOrGenerateKeyPair(tmpDir, "RS256", false)
	if err == nil {
		t.Error("expected error in production mode when keys are missing")
	}
}

func TestKeyRing_Rotate(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	kp1, err := LoadOrGenerateKeyPair(tmpDir1, "RS256", true)
	if err != nil {
		t.Fatalf("Generate kp1: %v", err)
	}
	kp1.KeyID = "key-v1"

	kp2, err := LoadOrGenerateKeyPair(tmpDir2, "RS256", true)
	if err != nil {
		t.Fatalf("Generate kp2: %v", err)
	}
	kp2.KeyID = "key-v2"

	kr := NewKeyRing(kp1)
	if kr.GetActive().KeyID != "key-v1" {
		t.Errorf("expected active key-v1, got %s", kr.GetActive().KeyID)
	}

	foundKey, ok := kr.GetKey("key-v1")
	if !ok || foundKey.KeyID != "key-v1" {
		t.Errorf("expected to find key-v1")
	}

	// Perform rotation
	kr.Rotate(kp2)

	if kr.GetActive().KeyID != "key-v2" {
		t.Errorf("expected active key-v2 after rotation, got %s", kr.GetActive().KeyID)
	}

	// Previous key must still be present for verification
	oldKey, ok := kr.GetKey("key-v1")
	if !ok || oldKey.KeyID != "key-v1" {
		t.Errorf("expected retired key-v1 to remain in KeyRing for token verification")
	}

	if len(kr.AllKeys()) != 2 {
		t.Errorf("expected 2 keys in KeyRing, got %d", len(kr.AllKeys()))
	}
}
