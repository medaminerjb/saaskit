# SaaSKit OpenID Conformance Documentation Package

## Overview

SaaSKit includes a full-featured, spec-compliant OpenID Connect Provider (OP) built on top of `zitadel/oidc` and SaaSKit's PostgreSQL identity engine.

This documentation package provides the technical audit trail, specification mapping, security evidence, and reproduction steps for OpenID Connect Foundation Conformance.

---

## Certification Target

- **Profile**: Basic OpenID Provider Certification Profile (Basic OP)
- **Specification**: OpenID Connect Core 1.0 (Authorization Code Flow) & OpenID Connect Discovery 1.0
- **Test Plan Alias**: `saaskit-basic`
- **SaaSKit Version**: `0.1.0-dev`

---

## Current Status

> 🟢 **CONFORMANCE SUITE PASSED — READY FOR CERTIFICATION SUBMISSION**

All 35 conformance test modules in test plan `Ag2Eo0agEY0i6` passed, completed review steps, or were validly skipped per OpenID specification rules. Automated Go unit tests run cleanly with 100% pass rate.

---

## Documentation Index

- 📋 [Certification Profile](profile.md) — Target profile parameters and scope
- 🧪 [Test Results Matrix](test-results.md) — Complete log of all executed conformance test modules
- 🛠️ [Resolved Failures & Fixes](failures.md) — Detailed root causes, code changes, and regression tests
- 📐 [Specification Compliance Matrix](compliance-matrix.md) — Feature mapping against OpenID specifications
- 🔒 [Security Evidence & Architecture](security-evidence.md) — PKCE, JWT, KMS key management, and session security
- 🔄 [Reproduction Guide](reproduce.md) — Step-by-step instructions to run the local Conformance Suite
- 📂 [Test Evidence Directory](evidence/oidcc-ensure-registered-redirect-uri/README.md) — Specific test logs and artifact evidence
- 📊 [Certification Readiness Report](certification-readiness.md) — Executive summary of test metrics and status
- ✅ [Submission Checklist](submission-checklist.md) — OpenID Foundation submission readiness checklist

---

## Important Disclaimer

> **Notice**: SaaSKit's internal conformance testing results demonstrate technical compliance with OpenID specifications, but do not by themselves constitute official OpenID Foundation certification. Official certification status is granted only upon formal submission to and publication by the OpenID Foundation on the [Certified OpenID Implementations list](https://openid.net/certification/).
