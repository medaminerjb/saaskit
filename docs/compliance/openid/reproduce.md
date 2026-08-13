# Conformance Test Reproduction Guide

## Overview

This guide provides step-by-step instructions to reproduce the OpenID Foundation Conformance Suite testing against SaaSKit.

---

## 1. Prerequisites

- **Go**: 1.22 or higher
- **Docker & Docker Compose**: Installed and running
- **Git**: Installed
- **Local Tunnel**: `localtunnel` (`lt`) or `ngrok` (required if running Conformance Suite in Docker or using cloud suite)

---

## 2. Server Configuration & Setup

### Step A: Database & Dependencies
Start PostgreSQL database:
```bash
docker-compose up -d postgres
```

### Step B: Launch SaaSKit Server
Set environment variables and launch SaaSKit with the public Issuer URL:

```bash
SAASKIT_BASE_URL="https://YOUR_TUNNEL_URL.loca.lt" \
SAASKIT_JWT_ISSUER="https://YOUR_TUNNEL_URL.loca.lt" \
go run ./cmd/saaskit
```

---

## 3. Conformance Suite Deployment

### Option A: Local Docker Deployment
1. Clone the OpenID Conformance Suite repository:
   ```bash
   git clone https://gitlab.com/openid/conformance-suite.git
   cd conformance-suite
   ```
2. Build and launch:
   ```bash
   MAVEN_CACHE=./m2 docker-compose -f builder-compose.yml run builder
   docker-compose up -d
   ```
3. Open the UI: `https://localhost:8443` *(Accept self-signed TLS certificate)*.

### Option B: OpenID Foundation Hosted Service
Navigate to [https://www.certification.openid.net/](https://www.certification.openid.net/).

---

## 4. Test Plan Setup

1. Click **Create Test Plan**.
2. Select Profile: **Basic OpenID Provider Certification Profile (Basic OP)**.
3. Switch to the **JSON** tab and paste the following configuration template:

```json
{
  "alias": "saaskit-basic",
  "description": "SaaSKit OpenID Basic OP Conformance Test Plan",
  "server": {
    "discoveryUrl": "https://YOUR_TUNNEL_URL.loca.lt/.well-known/openid-configuration"
  },
  "client": {
    "client_id": "test_client_id",
    "client_secret": "test_client_secret",
    "redirect_uri": "https://localhost:8443/test/a/saaskit-basic/callback"
  },
  "client_secret_post": {
    "client_id": "test_client_id",
    "client_secret": "test_client_secret"
  },
  "client2": {
    "client_id": "test_client_id_2",
    "client_secret": "test_client_secret_2"
  }
}
```

---

## 5. Test Execution Workflow

1. **Start Test Plan**: Click **Create Test Plan**.
2. **Execute Modules**: Click **Run** on each test module.
3. **Manual Interaction Steps**:
   - **`oidcc-ensure-registered-redirect-uri`**: The browser lands on an error page ("The requested redirect_uri is missing"). In the Conformance Suite UI top status bar, click **PASS / Continue** to confirm error page display without redirect.
4. **Verification**:
   - Ensure all tests complete with status **PASS** 🟢 or **SKIPPED (Permitted)** ⚪.
