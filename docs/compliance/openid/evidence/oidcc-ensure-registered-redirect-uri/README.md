# Evidence: oidcc-ensure-registered-redirect-uri

## Test Overview

- **Test Module**: `oidcc-ensure-registered-redirect-uri`
- **Specification**: OpenID Connect Core 1.0 Section 3.1.2.1
- **Goal**: Verify that authorization requests containing unregistered `redirect_uri` values are blocked at the authorization endpoint without redirecting back to the invalid URL.

---

## Expected & Actual Behavior

1. **Requested URL**:
   `https://fast-readers-post.loca.lt/authorize?client_id=test_client_id&redirect_uri=https://localhost.emobix.co.uk:8443/test/a/saaskit-basic/callback/INVALID_SUFFIX`
2. **Server Response**:
   HTTP 200/400 HTML Error Page rendered directly to the Resource Owner:
   > *"The requested redirect_uri is missing in the client configuration."*
3. **Redirection Status**:
   **No redirection** performed to the unapproved callback URL.
4. **Conformance Suite Verdict**:
   `PASS` *(Confirmed via manual verification button in UI)*.

---

## Code References

- Allowed Redirect URIs Whitelist: [models.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/models.go#L146)
- Unit Test: `TestClientAdapter_RedirectURIs` in [provider_test.go](file:///home/user/dev/perso/oidc-project/internal/oidc/provider/provider_test.go#L243)
