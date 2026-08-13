# OpenID Connect Certification Profile

## Overview

This document specifies the exact OpenID Foundation certification profile and deployment parameters targeted by SaaSKit for OpenID Provider (OP) conformance.

---

## Target Profile Details

| Parameter | Value / Setting |
| :--- | :--- |
| **Certification Profile** | **Basic OpenID Provider Certification Profile (Basic OP)** |
| **Specification** | OpenID Connect Core 1.0 (Authorization Code Flow) |
| **Test Plan Alias** | `saaskit-basic` |
| **Test Plan ID** | `Ag2Eo0agEY0i6` |
| **Test Execution Date** | August 12, 2026 |
| **SaaSKit Version** | `0.1.0-dev` |
| **Git Commit** | `HEAD` *(Branch: `main`)* |
| **Conformance Suite Version** | `5.2.3` |
| **Tested Issuer URL** | `https://fast-readers-post.loca.lt` |
| **Local Target URL** | `http://localhost:8080` |
| **Conformance Suite URL** | `https://localhost:8443` |

---

## Supported Features

- **Grant Types**: `authorization_code`, `refresh_token`
- **Response Types**: `code`
- **Response Modes**: `query`, `fragment`
- **Subject Types**: `public`
- **Signing Algorithms**: `RS256`, `ES256`, `EdDSA`
- **Token Endpoint Auth Methods**: `client_secret_basic`, `client_secret_post`, `none`
- **PKCE Support**: `S256` code challenge method enforced
- **Scopes Supported**: `openid`, `profile`, `email`, `address`, `phone`, `offline_access`
- **Standard Claims**: `sub`, `iss`, `aud`, `exp`, `iat`, `auth_time`, `nonce`, `email`, `email_verified`, `name`, `given_name`, `family_name`, `middle_name`, `nickname`, `profile`, `picture`, `website`, `gender`, `birthdate`, `zoneinfo`, `locale`, `updated_at`, `preferred_username`, `phone_number`, `phone_number_verified`, `address`
- **Session Management & Prompt**: `prompt=none`, `prompt=login`, `max_age` enforcement

---

## Unsupported / Out-of-Scope Features

- **Implicit Flow** (`response_type=token` / `id_token token`): Not targeted in Basic OP profile.
- **Hybrid Flow** (`response_type=code id_token`): Not targeted in Basic OP profile.
- **Dynamic Client Registration (OIDC Register)**: Manual/seeded client provisioning via database.
- **Pushed Authorization Requests (PAR / RFC 9126)**: Out of scope for Basic OP profile.
- **JWT Secured Authorization Requests (JAR / RFC 9101)**: Returns `request_not_supported` (permitted behavior).

---

## Client Configurations Tested

### Primary Client (`client`)
- **`client_id`**: `test_client_id`
- **`client_secret`**: `test_client_secret`
- **`redirect_uri`**: `https://localhost.emobix.co.uk:8443/test/a/saaskit-basic/callback`

### Second Client (`client2`)
- **`client_id`**: `test_client_id_2`
- **`client_secret`**: `test_client_secret_2`
- **`redirect_uri`**: `https://localhost.emobix.co.uk:8443/test/a/saaskit-basic/callback`
