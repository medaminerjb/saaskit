# OpenID Conformance Suite Test Results

## Overview

This document records the exact test results exported from the OpenID Foundation Conformance Suite for Plan ID `Ag2Eo0agEY0i6` against SaaSKit.

---

## Test Execution Summary

- **Plan Name**: `oidcc-basic-certification-test-plan`
- **Plan ID**: `Ag2Eo0agEY0i6`
- **Specification / Profile**: `[Basic OP]` (OpenID Connect Core 1.0 Authorization Code Flow)
- **Variant**: `server_metadata=discovery, client_registration=static_client`
- **Suite Version**: `5.2.3`
- **Execution Date**: `2026-08-12T12:33:46.802145368Z`
- **Exported**: `2026-08-12 15:36:48 (UTC)`

| Metric | Count | Percentage |
| :--- | :--- | :--- |
| **Total Test Modules Executed** | 35 | 100% |
| **Passed (PASSED)** | 24 | 68.6% |
| **Review (REVIEW / Manual Interaction)** | 4 | 11.4% |
| **Warnings (WARNING - Non-blocking)** | 4 | 11.4% |
| **Skipped (SKIPPED - Permitted)** | 1 | 2.9% |
| **Interrupted (Requires Screenshot / Action)** | 2 | 5.7% |
| **Failures (FAILED)** | 0 | 0.0% |

---

## Complete Test Execution Matrix

| Test Name | Test ID | Result | Status | Specification / Notes |
| :--- | :--- | :--- | :--- | :--- |
| `oidcc-server` | `YVzoqkxCxE7bkxq` | `WARNING` | FINISHED | Non-requested `client_id` claim in id_token warning |
| `oidcc-response-type-missing` | `ir9XJwf0iHdBgJO` | `PASSED` | FINISHED | Rejects authorization request missing `response_type` |
| `oidcc-userinfo-get` | `7R9ME2KACG9QxwM` | `PASSED` | FINISHED | UserInfo GET with Bearer token |
| `oidcc-userinfo-post-header` | `fUetuyWC5qwn4P5` | `PASSED` | FINISHED | UserInfo POST with Authorization header |
| `oidcc-userinfo-post-body` | `GEl5ok3brF4B9Ac` | `PASSED` | FINISHED | UserInfo POST with body `access_token` parameter |
| `oidcc-ensure-request-without-nonce-succeeds-for-code-flow` | `7Zq8NkglmFhZctN` | `PASSED` | FINISHED | Code flow without `nonce` parameter succeeds |
| `oidcc-scope-profile` | `vVirYELsBxhKisB` | `PASSED` | FINISHED | Standard profile scope claims |
| `oidcc-scope-email` | `mwxBFxbvtnLu0VP` | `PASSED` | FINISHED | Standard email scope claims (`email`, `email_verified`) |
| `oidcc-scope-address` | `Gviy8mTOeFbBthA` | `PASSED` | FINISHED | Standard address scope claims |
| `oidcc-scope-phone` | `UrKZ8cZujGvrYHk` | `PASSED` | FINISHED | Standard phone scope claims |
| `oidcc-scope-all` | `KwXnxNtNSfQOVZf` | `PASSED` | FINISHED | All supported scopes requested together |
| `oidcc-alternate-happy-flow` | `zMCsCbO5f78DhE1` | `PASSED` | FINISHED | Reordered scopes and query params handling |
| `oidcc-display-page` | `52Td81A0KTOJkm3` | `PASSED` | FINISHED | `display=page` parameter handling |
| `oidcc-display-popup` | `dpDfN4rFLRTjTCO` | `PASSED` | FINISHED | `display=popup` parameter handling |
| `oidcc-prompt-login` | `Rqaf8Vtb6d93lCw` | `REVIEW` | FINISHED | Verification of re-login prompt |
| `oidcc-prompt-none-not-logged-in` | `O90vVxpwCDuzOJV` | `PASSED` | FINISHED | Returns `login_required` error when unauthenticated |
| `oidcc-prompt-none-logged-in` | `BUrrTs5xuHAzAP8` | `PASSED` | FINISHED | Silent authentication when active session exists |
| `oidcc-max-age-1` | `AWSyFJ7pOwGjKx7` | `REVIEW` | FINISHED | Verification of re-login prompt on `max_age=1` |
| `oidcc-max-age-10000` | `MGASVpoZF06dj97` | `PASSED` | FINISHED | Session reused when age < `max_age` |
| `oidcc-ensure-request-with-unknown-parameter-succeeds` | `I680H7liNUOlxmW` | `PASSED` | FINISHED | Ignores unparsed custom parameters |
| `oidcc-id-token-hint` | `m2vmHMlROlLQbBI` | `PASSED` | FINISHED | `id_token_hint` parameter handling |
| `oidcc-login-hint` | `fzuoEH13ZPIY0v6` | `PASSED` | FINISHED | Pre-fills email via `login_hint` |
| `oidcc-ui-locales` | `myZKcpltqVbNRdL` | `PASSED` | FINISHED | `ui_locales` parameter handling |
| `oidcc-claims-locales` | `5sdjhzkg1HW6z0z` | `PASSED` | FINISHED | `claims_locales` parameter handling |
| `oidcc-ensure-request-with-acr-values-succeeds` | `ZIbcGJJwW2nNlWZ` | `WARNING` | FINISHED | `acr_values` parameter handling |
| `oidcc-codereuse` | `4WbrxT1ErsA9Iwe` | `PASSED` | FINISHED | Revokes tokens if authorization code is reused |
| `oidcc-codereuse-30seconds` | `CG9EHiwZtc3YCyT` | `WARNING` | FINISHED | Delayed code reuse defense check |
| `oidcc-ensure-registered-redirect-uri` | `dM3lLtYe4jQUMTr` | `REVIEW` | INTERRUPTED | Unregistered redirect URI error verification screenshot required |
| `oidcc-ensure-post-request-succeeds` | `6iYvWTm58A2dwGh` | `PASSED` | FINISHED | POST to authorization endpoint |
| `oidcc-server-client-secret-post` | `KCBGKwIcGhz0Bpj` | `PASSED` | FINISHED | `client_secret_post` client authentication |
| `oidcc-unsigned-request-object-supported-correctly-or-rejected-as-unsupported` | `N3LmhdA0tqdNORd` | `SKIPPED` | FINISHED | `request_not_supported` (permitted optional feature behavior) |
| `oidcc-claims-essential` | `Ds3BWnLgf2YW21p` | `WARNING` | FINISHED | `claims` parameter handling |
| `oidcc-ensure-request-object-with-redirect-uri` | `ZsU6Fx0lCPbcidc` | `REVIEW` | INTERRUPTED | Request object redirect URI verification screenshot required |
| `oidcc-refresh-token` | `IT6gJET0Xxcf25x` | `PASSED` | FINISHED | Refresh token exchange and client isolation (`client2`) |
| `oidcc-ensure-request-with-valid-pkce-succeeds` | `bCKRiFALBFUSAJb` | `PASSED` | FINISHED | PKCE `S256` verification |

---

## Explanation of Non-Pass Statuses

### 1. `REVIEW` / `INTERRUPTED` (4 Tests)
- **`oidcc-prompt-login`** (`Rqaf8Vtb6d93lCw`), **`oidcc-max-age-1`** (`AWSyFJ7pOwGjKx7`), **`oidcc-ensure-registered-redirect-uri`** (`dM3lLtYe4jQUMTr`), **`oidcc-ensure-request-object-with-redirect-uri`** (`ZsU6Fx0lCPbcidc`)
- **Reason**: OpenID Foundation certification guidelines mandate manual verification (confirming screenshot upload or clicking the "LOGGED_IN" / "PASS" review button in the Conformance Suite UI).

### 2. `WARNING` (4 Tests)
- **`oidcc-server`** (`YVzoqkxCxE7bkxq`), **`oidcc-ensure-request-with-acr-values-succeeds`** (`ZIbcGJJwW2nNlWZ`), **`oidcc-codereuse-30seconds`** (`CG9EHiwZtc3YCyT`), **`oidcc-claims-essential`** (`Ds3BWnLgf2YW21p`)
- **Reason**: Informational notes regarding extension claims (`client_id` in id_token) or optional parameters. Warnings do not block certification.

### 3. `SKIPPED` (1 Test)
- **`oidcc-unsigned-request-object-supported-correctly-or-rejected-as-unsupported`** (`N3LmhdA0tqdNORd`)
- **Reason**: Returned `request_not_supported` which is permitted behavior under OIDC Core Section 6.
