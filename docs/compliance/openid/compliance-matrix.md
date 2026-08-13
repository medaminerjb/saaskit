# OpenID Connect Specification Compliance Matrix

## Overview

This matrix maps SaaSKit's OpenID Connect Provider implementation against the specifications required for the **Basic OpenID Provider Certification Profile**.

---

## Compliance Matrix

| Requirement | Specification Section | SaaSKit Implementation | Source File / Evidence | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Authorization Code Flow** | OIDCC 1.0 §3.1 | Full authorization code exchange implementation via `zitadel/oidc` | [provider.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider.go) | `PASS` |
| **Discovery Document** | OIDC Discovery 1.0 §4 | Discovery metadata endpoint at `/.well-known/openid-configuration` | [router.go](file:///home/user/dev/perso/oidc-project/internal/identity/handler/router.go#L125) | `PASS` |
| **JWKS Endpoint** | RFC 7517 / OIDCC §10.1 | Exposes active public key set at `/keys` | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go) | `PASS` |
| **ID Token Issuance** | OIDCC 1.0 §2 | Signed JWT ID tokens containing `iss`, `sub`, `aud`, `exp`, `iat`, `auth_time`, `nonce` | `zitadel/oidc` & `storage.go` | `PASS` |
| **UserInfo Endpoint** | OIDCC 1.0 §5.3 | `/userinfo` endpoint supporting Bearer token authentication | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go#L517) | `PASS` |
| **Standard Scope Mapping** | OIDCC 1.0 §5.4 | Standard claims for `openid`, `profile`, `email`, `address`, `phone` | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go#L480) | `PASS` |
| **PKCE Enforcement** | RFC 7636 | Enforces `S256` code challenge and verifier validation | [provider.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider.go#L14) | `PASS` |
| **Client Authentication** | RFC 6749 §2.3.1 | Supports `client_secret_basic` and `client_secret_post` | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go#L360) | `PASS` |
| **Redirect URI Validation** | OIDCC 1.0 §3.1.2.1 | Exact string matching against client's registered redirect URIs | [models.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/models.go#L146) | `PASS` |
| **Cache-Control Headers** | RFC 6749 §5.1 | Returns `Cache-Control: no-store` and `Pragma: no-cache` on token & userinfo | [provider.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider.go#L34) | `PASS` |
| **Prompt Parameter (`none`)** | OIDCC 1.0 §3.1.2.1 | Silent authentication when active session exists; `login_required` error when missing | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go#L103) | `PASS` |
| **Prompt Parameter (`login`)** | OIDCC 1.0 §3.1.2.1 | Forces re-authentication and updates `auth_time` claim | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go#L117) | `PASS` |
| **Max Age Parameter (`max_age`)** | OIDCC 1.0 §3.1.2.1 | Enforces session age check and prompts for re-login when session exceeds `max_age` | [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go#L104) | `PASS` |
| **Refresh Token Rotation** | RFC 6749 §6.0 | Issues refresh tokens with optional rotation support | `storage.go` & `provider.go` | `PASS` |
| **Signing Algorithms** | RFC 7518 | Supports `RS256`, `ES256`, and `EdDSA` via KMS KeyRing | [crypto/kms](file:///home/user/dev/perso/oidc-project/internal/platform/crypto/kms) | `PASS` |
