package kms

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GCPKMSConfig contains configuration for GCP Cloud KMS.
type GCPKMSConfig struct {
	KeyName    string // e.g. projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{cryptoKey}
	Endpoint   string // Optional custom endpoint for testing
	HTTPClient *http.Client
}

// GCPKMS implements KMS using GCP Cloud Key Management Service API.
type GCPKMS struct {
	cfg        GCPKMSConfig
	httpClient *http.Client
}

// NewGCPKMS creates a new GCPKMS instance.
func NewGCPKMS(cfg GCPKMSConfig) (*GCPKMS, error) {
	if cfg.KeyName == "" {
		return nil, fmt.Errorf("gcp kms key name is required")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &GCPKMS{
		cfg:        cfg,
		httpClient: cfg.HTTPClient,
	}, nil
}

func (k *GCPKMS) Name() string {
	return "gcp-kms"
}

func (k *GCPKMS) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	payload := map[string]interface{}{
		"plaintext": base64.StdEncoding.EncodeToString(plaintext),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	endpoint := k.cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://cloudkms.googleapis.com/v1/%s:encrypt", k.cfg.KeyName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling gcp kms: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gcp kms returned status %d: %s", resp.StatusCode, string(b))
	}

	var res struct {
		Ciphertext string `json:"ciphertext"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return base64.StdEncoding.DecodeString(res.Ciphertext)
}

func (k *GCPKMS) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	payload := map[string]interface{}{
		"ciphertext": base64.StdEncoding.EncodeToString(ciphertext),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	endpoint := k.cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://cloudkms.googleapis.com/v1/%s:decrypt", k.cfg.KeyName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling gcp kms: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gcp kms returned status %d: %s", resp.StatusCode, string(b))
	}

	var res struct {
		Plaintext string `json:"plaintext"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return base64.StdEncoding.DecodeString(res.Plaintext)
}
