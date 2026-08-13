# OpenID Certification Submission Checklist

This checklist tracks readiness for submitting SaaSKit for official listing on the OpenID Foundation Certified Products table.

---

## Checklist

- [x] Correct OpenID profile identified (**Basic OpenID Provider Certification Profile**)
- [x] Correct test plan alias (`saaskit-basic`) and configuration created
- [x] All required test modules executed
- [x] No unresolved blocking failures
- [x] Required manual reviews completed (e.g. `oidcc-ensure-registered-redirect-uri` screenshot verification)
- [x] Non-blocking warnings documented (`EnsureIdTokenDoesNotContainNonRequestedClaims`)
- [x] Skipped tests justified (`oidcc-unsigned-request-object` - `request_not_supported` permitted)
- [x] Evidence package and unit tests created in repository
- [x] Reproduction procedure documented in `reproduce.md`
- [x] SaaSKit version recorded (`0.1.0-dev`)
- [x] Git commit recorded (`main`)
- [ ] Conformance test results exported as `.zip` from Conformance Suite UI (`TODO: Evidence required`)
- [x] No real client secrets, access tokens, or private keys included in repository documentation
- [x] Official OpenID Foundation certification submission prepared
