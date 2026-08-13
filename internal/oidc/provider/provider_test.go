package provider

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	idcrypto "github.com/medaminerjb/saas-kit/internal/identity/crypto"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func TestSetupProvider_DiscoveryEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	keyPair, err := idcrypto.LoadOrGenerateKeyPair(t.TempDir(), "RS256", true)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	storageCfg := StorageConfig{
		KeyPair: keyPair,
		Logger:  logger,
	}
	storage := NewStorage(storageCfg)

	issuer := "http://localhost:8080"
	handler, err := SetupProvider(issuer, storage, logger)
	if err != nil {
		t.Fatalf("SetupProvider failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, issuer+"/.well-known/openid-configuration", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var discovery map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &discovery); err != nil {
		t.Fatalf("failed to parse discovery JSON: %v", err)
	}

	if discovery["issuer"] != issuer {
		t.Errorf("expected issuer %s, got %v", issuer, discovery["issuer"])
	}

	requiredFields := []string{
		"authorization_endpoint",
		"token_endpoint",
		"jwks_uri",
		"response_types_supported",
		"subject_types_supported",
		"id_token_signing_alg_values_supported",
	}

	for _, field := range requiredFields {
		if _, ok := discovery[field]; !ok {
			t.Errorf("missing required discovery field: %s", field)
		}
	}
}

func TestStorage_AuthRequestLifecycle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	keyPair, err := idcrypto.LoadOrGenerateKeyPair(t.TempDir(), "RS256", true)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	storage := NewStorage(StorageConfig{
		KeyPair: keyPair,
		Logger:  logger,
	})

	ctx := context.Background()
	authReq := &oidc.AuthRequest{
		ClientID:    "client_123",
		RedirectURI: "http://localhost:3000/callback",
		State:       "xyz123",
		Scopes:      []string{"openid", "profile", "email"},
	}

	userID := uuid.NewString()
	req, err := storage.CreateAuthRequest(ctx, authReq, userID)
	if err != nil {
		t.Fatalf("CreateAuthRequest failed: %v", err)
	}

	if req.GetClientID() != "client_123" {
		t.Errorf("expected client_123, got %s", req.GetClientID())
	}

	// Fetch by ID
	fetched, err := storage.AuthRequestByID(ctx, req.GetID())
	if err != nil {
		t.Fatalf("AuthRequestByID failed: %v", err)
	}
	if fetched.GetID() != req.GetID() {
		t.Errorf("expected ID %s, got %s", req.GetID(), fetched.GetID())
	}

	// Save code
	code := "test_code_abc"
	if err := storage.SaveAuthCode(ctx, req.GetID(), code); err != nil {
		t.Fatalf("SaveAuthCode failed: %v", err)
	}

	// Fetch by Code
	byCode, err := storage.AuthRequestByCode(ctx, code)
	if err != nil {
		t.Fatalf("AuthRequestByCode failed: %v", err)
	}
	if byCode.GetID() != req.GetID() {
		t.Errorf("expected ID %s, got %s", req.GetID(), byCode.GetID())
	}

	// Delete
	if err := storage.DeleteAuthRequest(ctx, req.GetID()); err != nil {
		t.Fatalf("DeleteAuthRequest failed: %v", err)
	}

	if _, err := storage.AuthRequestByID(ctx, req.GetID()); err == nil {
		t.Error("expected error fetching deleted auth request, got nil")
	}
}

func TestStorage_TokenCreationAndIntrospection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	keyPair, err := idcrypto.LoadOrGenerateKeyPair(t.TempDir(), "RS256", true)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	storage := NewStorage(StorageConfig{
		KeyPair: keyPair,
		Logger:  logger,
	})

	userID := uuid.NewString()
	clientID := "app_test"
	scopes := []string{"openid", "profile"}

	token := storage.createOIDCToken(clientID, "refresh_1", userID, []string{clientID}, scopes)
	if token.ID == "" {
		t.Fatal("expected non-empty token ID")
	}

	if token.Expiration.Before(time.Now()) {
		t.Error("token expiration should be in the future")
	}
}

func TestProvider_CacheControlHeaders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	keyPair, err := idcrypto.LoadOrGenerateKeyPair(t.TempDir(), "RS256", true)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	storage := NewStorage(StorageConfig{
		KeyPair: keyPair,
		Logger:  logger,
	})

	issuer := "http://localhost:8080"
	handler, err := SetupProvider(issuer, storage, logger)
	if err != nil {
		t.Fatalf("SetupProvider failed: %v", err)
	}

	endpoints := []string{"/oauth/token", "/userinfo"}
	for _, endpoint := range endpoints {
		req := httptest.NewRequest(http.MethodGet, issuer+endpoint, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if cacheControl := rec.Header().Get("Cache-Control"); cacheControl != "no-store" {
			t.Errorf("endpoint %s expected Cache-Control: no-store, got %q", endpoint, cacheControl)
		}
		if pragma := rec.Header().Get("Pragma"); pragma != "no-cache" {
			t.Errorf("endpoint %s expected Pragma: no-cache, got %q", endpoint, pragma)
		}
	}
}

func TestStorage_PromptNoneAndSessionReuse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	keyPair, err := idcrypto.LoadOrGenerateKeyPair(t.TempDir(), "RS256", true)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	storage := NewStorage(StorageConfig{
		KeyPair: keyPair,
		Logger:  logger,
	})

	ctx := context.Background()

	// 1. prompt=none without active session should fail with login_required
	authReqNoneNoSession := &oidc.AuthRequest{
		ClientID:    "test_client_id",
		RedirectURI: "http://localhost:8080/callback",
		Prompt:      []string{"none"},
	}
	_, err = storage.CreateAuthRequest(ctx, authReqNoneNoSession, "user_123")
	if err == nil {
		t.Error("expected error for prompt=none without session, got nil")
	}

	// 2. Establish an active session
	initialTime := time.Now().Add(-10 * time.Minute)
	storage.userSessions["test_client_id"] = &userSession{
		UserID:   "user_123",
		AuthTime: initialTime,
	}

	// 3. prompt=none with active session should succeed and preserve AuthTime
	authReqNoneWithSession := &oidc.AuthRequest{
		ClientID:    "test_client_id",
		RedirectURI: "http://localhost:8080/callback",
		Prompt:      []string{"none"},
	}
	reqNone, err := storage.CreateAuthRequest(ctx, authReqNoneWithSession, "user_123")
	if err != nil {
		t.Fatalf("CreateAuthRequest prompt=none with session failed: %v", err)
	}
	if !reqNone.Done() {
		t.Error("expected prompt=none request to be completed (Done() == true)")
	}
	if !reqNone.GetAuthTime().Equal(initialTime) {
		t.Errorf("expected AuthTime %v, got %v", initialTime, reqNone.GetAuthTime())
	}
}

func TestClientAdapter_RedirectURIs(t *testing.T) {
	adapter := &ClientAdapter{
		ID:       "test_client_id",
		DevMode_: true,
	}

	uris := adapter.RedirectURIs()
	if len(uris) == 0 {
		t.Fatal("expected non-empty allowed redirect URIs")
	}

	expectedURIs := []string{
		"https://localhost.emobix.co.uk:8443/test/a/saaskit-basic/callback",
		"https://localhost.emobix.co.uk:8443/test/a/saaskit-config/callback",
		"https://localhost.emobix.co.uk:8443/test/a/saaskit-test/callback",
		"https://localhost.emobix.co.uk:8443/test/a/saaskit-refresh/callback",
	}

	for _, expected := range expectedURIs {
		found := false
		for _, u := range uris {
			if u == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected redirect URI %s to be allowed in DevMode", expected)
		}
	}
}
