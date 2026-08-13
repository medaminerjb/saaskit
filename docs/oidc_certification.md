# OpenID Connect Conformance & Certification Guide

This document details the OpenID Connect (OIDC) specification compliance, supported profiles, claim mappings, and instructions for executing the **OpenID Foundation Conformance Test Suite** against **SaaSKit**.

---

## 📋 Conformance Overview

SaaSKit implements a compliant OpenID Connect Provider (OP) built on top of `zitadel/oidc` and SaaSKit's PostgreSQL identity engine.

### Supported Profiles

| Profile Name | Status | Specification |
| --- | --- | --- |
| **Basic OP** | ✅ Supported | OpenID Connect Core 1.0 (Authorization Code Flow) |
| **Config OP** | ✅ Supported | OpenID Connect Discovery 1.0 |
| **JWK OP** | ✅ Supported | JSON Web Key (JWK) publishing |
| **Refresh OP** | ✅ Supported | Token Refresh Rotation |

---

## 🛠️ Discovery Metadata

The OpenID Connect Discovery document is served at `/.well-known/openid-configuration`.

```json
{
  "issuer": "http://localhost:8080",
  "authorization_endpoint": "http://localhost:8080/authorize",
  "token_endpoint": "http://localhost:8080/oauth/token",
  "userinfo_endpoint": "http://localhost:8080/userinfo",
  "jwks_uri": "http://localhost:8080/keys",
  "revocation_endpoint": "http://localhost:8080/revoke",
  "introspection_endpoint": "http://localhost:8080/oauth/introspect",
  "response_types_supported": ["code"],
  "response_modes_supported": ["query", "fragment"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256", "ES256", "EdDSA"],
  "scopes_supported": ["openid", "profile", "email", "offline_access"],
  "token_endpoint_auth_methods_supported": ["client_secret_basic", "client_secret_post", "none"],
  "grant_types_supported": ["authorization_code", "refresh_token"],
  "code_challenge_methods_supported": ["S256"],
  "claims_supported": [
    "sub", "iss", "aud", "exp", "iat", "auth_time", "nonce",
    "email", "email_verified", "name", "preferred_username", "picture", "updated_at"
  ]
}
```

---

## 🏷️ Standard Claim Mapping Matrix

The UserInfo endpoint (`/userinfo`) maps user attributes based on requested OIDC scopes:

| Scope | Claim | Type | Description |
| --- | --- | --- | --- |
| `openid` | `sub` | `string` | Unique User UUID |
| `email` | `email` | `string` | Primary user email address |
| `email` | `email_verified` | `boolean` | Verification status of the user email |
| `profile` | `name` | `string` | User display name |
| `profile` | `preferred_username` | `string` | User handle / email alias |
| `profile` | `picture` | `string` | Avatar image URL |
| `profile` | `updated_at` | `number` | Unix timestamp of last user profile modification |

---

## 🧪 Running the OpenID Foundation Conformance Suite

### Prerequisites

1. Run SaaSKit server locally or in Docker:
   ```bash
   make run
   ```
2. Clone and build the official OpenID Foundation Conformance Suite locally via Docker:
   ```bash
   git clone https://gitlab.com/openid/conformance-suite.git
   cd conformance-suite

   # Step A: Build the Java JAR using the builder container
   MAVEN_CACHE=./m2 docker-compose -f builder-compose.yml run builder

   # Step B: Launch MongoDB, Nginx, and the Conformance Server
   docker-compose up -d
   ```
   *Alternatively, use the official hosted test service at [https://www.certification.openid.net/](https://www.certification.openid.net/).*

### Test Suite Configuration

1. Access the web interface using **HTTPS**: `https://localhost:8443/` *(Note: ensure you use `https://`, as plain `http://` will trigger Nginx 400 Bad Request error. Accept the self-signed TLS certificate warning).*
2. Create a new test plan for **Basic OP** or **Config OP**.
3. Set JSON Configuration:
   ```json
   {
     "alias": "saaskit-test",
     "description": "SaaSKit OIDC Provider Conformance Test",
     "server": {
       "discoveryUrl": "http://saaskit:8080/.well-known/openid-configuration"
     },
     "client": {
       "client_id": "test_client_id",
       "client_secret": "test_client_secret",
       "redirect_uri": "https://localhost:8443/test/a/saaskit-test/callback"
     }
   }
   ```
4. Execute all test cases for **Basic OP** and ensure all status indicators show green (`PASSED`).

---

## 🔒 Security Specifications

- **PKCE Requirement:** `S256` code challenge is enforced for public clients.
- **Asymmetric Key Rotation:** RS256/ES256/EdDSA keys are exposed via `/keys` and automatically rotated via the KMS KeyRing engine.
- **Token Security:** Short-lived access tokens (5 minutes) with refresh token rotation and 10-second concurrency grace periods.
