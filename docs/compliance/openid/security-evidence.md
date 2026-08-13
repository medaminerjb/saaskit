# Security Evidence & Architecture

## Overview

This document describes the security controls, cryptographic mechanisms, and validation routines enforced within SaaSKit to satisfy OpenID Connect certification requirements.

> **Note**: No production secrets, actual private key bytes, client secrets, or user credentials are contained in this document.

---

## Security Implementation Mapping

| Security Control | Implementation Details | Source Code | Tests | Status |
| :--- | :--- | :--- | :--- | :--- |
| **PKCE Enforcement** | Requires `code_challenge` (S256) on authorization code requests; verifies `code_verifier` during token exchange. | [provider.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider.go#L14) | `TestStorage_AuthRequestLifecycle` | `PASS` |
| **Redirect URI Validation** | Strict whitelist matching against stored client redirect URIs. Rejects wildcards in production. | [models.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/models.go#L146) | `TestClientAdapter_RedirectURIs` | `PASS` |
| **Password Hashing** | Argon2id / bcrypt hashing engine with salt via `Hasher`. | [crypto/hasher.go](file:///home/user/dev/perso/oidc-project/internal/identity/crypto) | Unit tests in `internal/identity/crypto` | `PASS` |
| **Secret Storage** | Client secrets and sensitive parameters encrypted at rest using AES-256-GCM envelope encryption. | [platform/crypto](file:///home/user/dev/perso/oidc-project/internal/platform/crypto) | Unit tests in `internal/platform/crypto` | `PASS` |
| **JWT Key Management** | Asymmetric key pairs (RS256 / ES256 / EdDSA) managed by KMS KeyRing engine and published at `/keys`. | [crypto/kms](file:///home/user/dev/perso/oidc-project/internal/platform/crypto/kms) | Key rotation tests | `PASS` |
| **Token Expiration** | Access tokens expired after 5 minutes; authorization codes expired after 60 seconds. | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go#L521) | `TestStorage_TokenCreationAndIntrospection` | `PASS` |
| **Response Caching Controls** | Attachment of `Cache-Control: no-store` and `Pragma: no-cache` headers on token/userinfo HTTP responses. | [provider.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider.go#L34) | `TestProvider_CacheControlHeaders` | `PASS` |
| **Session Security** | In-memory thread-safe `userSessions` map with `AuthTime` tracking and `max_age` expiration checks. | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go#L35) | `TestStorage_PromptNoneAndSessionReuse` | `PASS` |

---

## Cryptographic Key Management

- **Signing Algorithms**: `RS256` (default for certification), `ES256`, `EdDSA`
- **JWKS Endpoint**: `GET /keys`
- **Rotation Strategy**: Keys auto-generated and signed dynamically or loaded from specified RSA/ECDSA key pairs on startup via `idcrypto.LoadOrGenerateKeyPair`.

---

## Authorization Code Exchange Security

1. **One-Time Code Usage**: Authorization codes are deleted immediately after exchange in `DeleteAuthRequest`.
2. **Client Binding**: Tokens are bound to `ClientID` and `Subject` (`UserID`).
3. **PKCE Verification**: Hash of `code_verifier` checked against stored `code_challenge`.
