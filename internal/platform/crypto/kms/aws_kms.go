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

// AWSKMSConfig contains configuration for AWS KMS integration.
type AWSKMSConfig struct {
	KeyARN     string
	Region     string
	Endpoint   string // Optional custom endpoint for testing / LocalStack
	HTTPClient *http.Client
}

// AWSKMS implements KMS using AWS Key Management Service API.
type AWSKMS struct {
	cfg        AWSKMSConfig
	httpClient *http.Client
}

// NewAWSKMS creates a new AWSKMS instance.
func NewAWSKMS(cfg AWSKMSConfig) (*AWSKMS, error) {
	if cfg.KeyARN == "" {
		return nil, fmt.Errorf("aws kms key ARN is required")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &AWSKMS{
		cfg:        cfg,
		httpClient: cfg.HTTPClient,
	}, nil
}

func (k *AWSKMS) Name() string {
	return "aws-kms"
}

func (k *AWSKMS) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	payload := map[string]interface{}{
		"KeyId":     k.cfg.KeyARN,
		"Plaintext": base64.StdEncoding.EncodeToString(plaintext),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	endpoint := k.cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://kms.%s.amazonaws.com", k.cfg.Region)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "TrentService.Encrypt")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling aws kms: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("aws kms returned status %d: %s", resp.StatusCode, string(b))
	}

	var res struct {
		CiphertextBlob string `json:"CiphertextBlob"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return base64.StdEncoding.DecodeString(res.CiphertextBlob)
}

func (k *AWSKMS) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	payload := map[string]interface{}{
		"CiphertextBlob": base64.StdEncoding.EncodeToString(ciphertext),
		"KeyId":          k.cfg.KeyARN,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	endpoint := k.cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://kms.%s.amazonaws.com", k.cfg.Region)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "TrentService.Decrypt")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling aws kms: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("aws kms returned status %d: %s", resp.StatusCode, string(b))
	}

	var res struct {
		Plaintext string `json:"Plaintext"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return base64.StdEncoding.DecodeString(res.Plaintext)
}
