package kms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalKMS(t *testing.T) {
	key := []byte("01234567890123456789012345678901") // 32 bytes
	kms, err := NewLocalKMS(key)
	if err != nil {
		t.Fatalf("NewLocalKMS error: %v", err)
	}

	if kms.Name() != "local" {
		t.Errorf("expected name local, got %s", kms.Name())
	}

	ctx := context.Background()
	plaintext := []byte("secret payload 123")

	ciphertext, err := kms.Encrypt(ctx, plaintext)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	decrypted, err := kms.Decrypt(ctx, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("expected %s, got %s", string(plaintext), string(decrypted))
	}
}

func TestAWSKMS_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-Amz-Target")
		if target == "TrentService.Encrypt" {
			res := map[string]string{
				"CiphertextBlob": base64.StdEncoding.EncodeToString([]byte("mocked-aws-ciphertext")),
			}
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		if target == "TrentService.Decrypt" {
			res := map[string]string{
				"Plaintext": base64.StdEncoding.EncodeToString([]byte("secret payload 123")),
			}
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		http.Error(w, "unknown target", http.StatusBadRequest)
	}))
	defer server.Close()

	kms, err := NewAWSKMS(AWSKMSConfig{
		KeyARN:   "arn:aws:kms:us-east-1:123456789012:key/test-key",
		Endpoint: server.URL,
	})
	if err != nil {
		t.Fatalf("NewAWSKMS error: %v", err)
	}

	ctx := context.Background()
	cipher, err := kms.Encrypt(ctx, []byte("secret payload 123"))
	if err != nil {
		t.Fatalf("AWS Encrypt error: %v", err)
	}

	plain, err := kms.Decrypt(ctx, cipher)
	if err != nil {
		t.Fatalf("AWS Decrypt error: %v", err)
	}

	if string(plain) != "secret payload 123" {
		t.Errorf("expected secret payload 123, got %s", string(plain))
	}
}

func TestGCPKMS_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/test-key:encrypt" {
			res := map[string]string{
				"ciphertext": base64.StdEncoding.EncodeToString([]byte("mocked-gcp-ciphertext")),
			}
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		if r.URL.Path == "/v1/test-key:decrypt" {
			res := map[string]string{
				"plaintext": base64.StdEncoding.EncodeToString([]byte("secret payload 123")),
			}
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	kms, err := NewGCPKMS(GCPKMSConfig{
		KeyName:  "test-key",
		Endpoint: server.URL + "/v1/test-key:encrypt",
	})
	if err != nil {
		t.Fatalf("NewGCPKMS error: %v", err)
	}

	ctx := context.Background()
	cipher, err := kms.Encrypt(ctx, []byte("secret payload 123"))
	if err != nil {
		t.Fatalf("GCP Encrypt error: %v", err)
	}

	kmsDecrypt, _ := NewGCPKMS(GCPKMSConfig{
		KeyName:  "test-key",
		Endpoint: server.URL + "/v1/test-key:decrypt",
	})
	plain, err := kmsDecrypt.Decrypt(ctx, cipher)
	if err != nil {
		t.Fatalf("GCP Decrypt error: %v", err)
	}

	if string(plain) != "secret payload 123" {
		t.Errorf("expected secret payload 123, got %s", string(plain))
	}
}

func TestVaultKMS_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/transit/encrypt/saaskit-key" {
			res := map[string]interface{}{
				"data": map[string]string{
					"ciphertext": "vault:v1:mocked-ciphertext",
				},
			}
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		if r.URL.Path == "/v1/transit/decrypt/saaskit-key" {
			res := map[string]interface{}{
				"data": map[string]string{
					"plaintext": base64.StdEncoding.EncodeToString([]byte("secret payload 123")),
				},
			}
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	kms, err := NewVaultKMS(VaultKMSConfig{
		Address: server.URL,
		Token:   "s.mocktoken",
		KeyName: "saaskit-key",
	})
	if err != nil {
		t.Fatalf("NewVaultKMS error: %v", err)
	}

	ctx := context.Background()
	cipher, err := kms.Encrypt(ctx, []byte("secret payload 123"))
	if err != nil {
		t.Fatalf("Vault Encrypt error: %v", err)
	}

	plain, err := kms.Decrypt(ctx, cipher)
	if err != nil {
		t.Fatalf("Vault Decrypt error: %v", err)
	}

	if string(plain) != "secret payload 123" {
		t.Errorf("expected secret payload 123, got %s", string(plain))
	}
}
