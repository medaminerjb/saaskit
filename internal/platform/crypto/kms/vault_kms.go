package kms

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VaultKMSConfig contains configuration for HashiCorp Vault Transit Engine.
type VaultKMSConfig struct {
	Address    string // e.g. "http://127.0.0.1:8200"
	Token      string
	KeyName    string // e.g. "saaskit-master-key"
	HTTPClient *http.Client
}

// VaultKMS implements KMS using HashiCorp Vault Transit Secrets Engine.
type VaultKMS struct {
	cfg        VaultKMSConfig
	httpClient *http.Client
}

// NewVaultKMS creates a new VaultKMS instance.
func NewVaultKMS(cfg VaultKMSConfig) (*VaultKMS, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("vault address is required")
	}
	if cfg.KeyName == "" {
		return nil, fmt.Errorf("vault key name is required")
	}
	cfg.Address = strings.TrimSuffix(cfg.Address, "/")
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &VaultKMS{
		cfg:        cfg,
		httpClient: cfg.HTTPClient,
	}, nil
}

func (k *VaultKMS) Name() string {
	return "vault"
}

func (k *VaultKMS) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	payload := map[string]interface{}{
		"plaintext": base64.StdEncoding.EncodeToString(plaintext),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/transit/encrypt/%s", k.cfg.Address, k.cfg.KeyName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if k.cfg.Token != "" {
		req.Header.Set("X-Vault-Token", k.cfg.Token)
	}

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling vault transit encrypt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(b))
	}

	var res struct {
		Data struct {
			Ciphertext string `json:"ciphertext"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return []byte(res.Data.Ciphertext), nil
}

func (k *VaultKMS) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	payload := map[string]interface{}{
		"ciphertext": string(ciphertext),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/transit/decrypt/%s", k.cfg.Address, k.cfg.KeyName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if k.cfg.Token != "" {
		req.Header.Set("X-Vault-Token", k.cfg.Token)
	}

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling vault transit decrypt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(b))
	}

	var res struct {
		Data struct {
			Plaintext string `json:"plaintext"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return base64.StdEncoding.DecodeString(res.Data.Plaintext)
}
