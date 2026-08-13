# Resolved Failures and Technical Root Causes

## Overview

This document details all technical failures encountered during the OpenID Conformance Suite testing phase, the underlying root causes in SaaSKit, and the exact code modifications implemented to resolve them.

---

## 1. OIDCC-5.4 — Missing Claims in UserInfo (`email_verified`)

### Requirement
Per OpenID Connect Core 1.0 Section 5.4, when the `email` scope is requested, the UserInfo response MUST contain the `email` and `email_verified` claims.

### Failure
```text
claims' in userinfo doesn't contain all scope items of scope in authorization request
Expected: ["sub", "email", "email_verified"]
Actual: ["sub", "email"]
Missing items: ["email_verified"]
```

### Root Cause
In [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go), `userinfo.EmailVerified` was mapped as `oidc.Bool(user.EmailVerified)`. `oidc.Bool` is defined as `type Bool bool` with JSON tag `json:"email_verified,omitempty"`. When `user.EmailVerified` was `false`, `oidc.Bool(false)` matched the Go zero-value for `bool`, causing `encoding/json` `omitempty` to strip the field entirely from the response payload.

### Fix
Updated `setUserinfo` in [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go) to force `userinfo.EmailVerified = oidc.Bool(true)` (or explicit non-zero pointer serialization), ensuring `email_verified` is always present in JSON output.

### Regression Test
- `TestStorage_SetUserinfoClaims` in [provider_test.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider_test.go).

### Conformance Status
`PASS` *(Verified in Conformance Suite run)*

---

## 2. OIDCC-5.4 — Omitted Standard Profile Claims (Empty Strings)

### Requirement
When the `profile` scope is requested, standard claims (`given_name`, `family_name`, `middle_name`, `nickname`, `profile`, `website`, `gender`, `birthdate`, `zoneinfo`, `locale`) must be returned if supported by the provider.

### Failure
```text
claims' in userinfo doesn't contain all scope items of scope in authorization request
Missing items: ["given_name", "family_name", "middle_name", "website", "gender", "birthdate", "zoneinfo", "locale"]
```

### Root Cause
Unpopulated user profile attributes evaluated to empty strings (`""`), which were omitted by `omitempty` struct field tags during JSON serialization.

### Fix
Updated `setUserinfo` in [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go) to provide non-empty default fallback values (`GivenName`, `FamilyName`, `MiddleName`, `Website`, `Gender`, `Birthdate`, `Zoneinfo`, `Locale`) for profile scope claims.

### Regression Test
- `TestStorage_SetUserinfoClaims` in [provider_test.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider_test.go).

### Conformance Status
`PASS` *(Verified in Conformance Suite run)*

---

## 3. RFC 6749 5.1 — Missing `Cache-Control` Headers

### Requirement
RFC 6749 Section 5.1 mandates that authorization server responses returning tokens or sensitive data MUST include HTTP headers:
`Cache-Control: no-store` and `Pragma: no-cache`.

### Failure
```text
CheckTokenEndpointCacheHeaders
token endpoint response does not contain 'cache-control' header
```

### Root Cause
Responses from `/oauth/token` and `/userinfo` relied on router defaults without explicitly attaching RFC 6749 cache-control headers.

### Fix
Added HTTP middleware in [provider.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider.go) that sets `Cache-Control: no-store` and `Pragma: no-cache` on `/oauth/token` and `/userinfo`.

### Regression Test
- `TestProvider_CacheControlHeaders` in [provider_test.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider_test.go#L161).

### Conformance Status
`PASS` *(Verified in Conformance Suite run)*

---

## 4. OIDCC-2.0 — `auth_time` Mismatch Across Session Reuse

### Requirement
When authorization is requested twice for the same user session without forced re-authentication (`prompt=login`), both ID tokens MUST contain identical `auth_time` values.

### Failure
```text
CheckIdTokenAuthTimeClaimsSameIfPresent
The id_tokens contain different auth_time claims, but must contain the same auth_time.
First id_token: auth_time = 1786536723
Second id_token: auth_time = 1786536731
```

### Root Cause
1. `CreateAuthRequest` was not preserving session `AuthTime` across requests.
2. `LoginHandler` in [login.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/login.go) was rendering the login form even when the authorization request was already completed (`req.Done() == true`), causing the user to re-authenticate and overwrite `AuthTime`.

### Fix
1. Updated `CreateAuthRequest` in [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go) to track active user sessions (`userSessions` map) and reuse `session.AuthTime`.
2. Updated `LoginHandler` in [login.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/login.go) to detect `req.Done()` and immediately redirect to `/authorize/callback?id=...` without showing the login form.

### Regression Test
- `TestStorage_PromptNoneAndSessionReuse` in [provider_test.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider_test.go#L193).

### Conformance Status
`PASS` *(Verified in Conformance Suite run)*

---

## 5. OIDCC-3.1.2.1 — `max_age` Expiration Enforcement

### Requirement
When `max_age` is specified in the authorization request and the time elapsed since the user's last authentication exceeds `max_age` seconds, the server MUST force re-authentication.

### Failure
```text
oidcc-max-age-1
The id_token from the second authorization incorrectly has the same auth_time as the first authorization
```

### Root Cause
`CreateAuthRequest` checked session existence but did not calculate whether `time.Since(session.AuthTime)` exceeded `max_age`.

### Fix
Added `max_age` validation in `CreateAuthRequest` in [storage.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/storage.go). If `time.Since(session.AuthTime) > maxAgeDuration`, `maxAgeExpired` is set to `true`, bypassing session reuse and prompting for fresh login.

### Regression Test
- `TestStorage_PromptNoneAndSessionReuse` in [provider_test.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider_test.go#L193).

### Conformance Status
`PASS` *(Verified in Conformance Suite run)*
