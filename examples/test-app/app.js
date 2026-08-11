// State Management
const state = {
  accessToken: localStorage.getItem('saaskit_access_token') || '',
  refreshToken: localStorage.getItem('saaskit_refresh_token') || '',
  currentUser: null,
  activeTenant: null,
  tenants: []
};

// DOM Elements
const elements = {
  serverUrl: document.getElementById('serverUrl'),
  statusBadge: document.getElementById('statusBadge'),
  statusText: document.getElementById('statusText'),
  btnHealth: document.getElementById('btnHealth'),
  
  currentUserEmail: document.getElementById('currentUserEmail'),
  tokenPreview: document.getElementById('tokenPreview'),
  activeTenantName: document.getElementById('activeTenantName'),

  consoleOutput: document.getElementById('consoleOutput'),
  lastStatusCode: document.getElementById('lastStatusCode'),
  lastLatency: document.getElementById('lastLatency'),
  btnClearConsole: document.getElementById('btnClearConsole'),

  // Auth Inputs
  regEmail: document.getElementById('regEmail'),
  regPassword: document.getElementById('regPassword'),
  btnRegister: document.getElementById('btnRegister'),
  loginEmail: document.getElementById('loginEmail'),
  loginPassword: document.getElementById('loginPassword'),
  btnLogin: document.getElementById('btnLogin'),
  btnRefresh: document.getElementById('btnRefresh'),
  btnLogout: document.getElementById('btnLogout'),

  // Profile
  btnGetMe: document.getElementById('btnGetMe'),
  profileName: document.getElementById('profileName'),
  btnUpdateProfile: document.getElementById('btnUpdateProfile'),
  publicMeta: document.getElementById('publicMeta'),
  privateMeta: document.getElementById('privateMeta'),
  btnGetMeta: document.getElementById('btnGetMeta'),
  btnUpdateMeta: document.getElementById('btnUpdateMeta'),

  // Tenants
  tenantName: document.getElementById('tenantName'),
  tenantSlug: document.getElementById('tenantSlug'),
  btnCreateTenant: document.getElementById('btnCreateTenant'),
  btnListTenants: document.getElementById('btnListTenants'),
  tenantSelect: document.getElementById('tenantSelect'),
  btnSwitchTenant: document.getElementById('btnSwitchTenant'),
  inviteEmail: document.getElementById('inviteEmail'),
  inviteRole: document.getElementById('inviteRole'),
  btnInviteMember: document.getElementById('btnInviteMember'),
  btnListMembers: document.getElementById('btnListMembers'),

  // API Keys
  apiKeyName: document.getElementById('apiKeyName'),
  apiKeyType: document.getElementById('apiKeyType'),
  apiKeyScopes: document.getElementById('apiKeyScopes'),
  btnCreateApiKey: document.getElementById('btnCreateApiKey'),
  btnListApiKeys: document.getElementById('btnListApiKeys'),
  testApiKeySecret: document.getElementById('testApiKeySecret'),
  btnTestApiKey: document.getElementById('btnTestApiKey'),

  // OIDC
  btnOidcDiscovery: document.getElementById('btnOidcDiscovery'),
  btnJwks: document.getElementById('btnJwks'),
  oidcClientId: document.getElementById('oidcClientId'),
  oidcRedirectUri: document.getElementById('oidcRedirectUri'),
  btnLaunchOidc: document.getElementById('btnLaunchOidc'),

  // OIDC Certification Suite
  btnRunCertAudit: document.getElementById('btnRunCertAudit'),
  certScoreVal: document.getElementById('certScoreVal'),
  certProgressBar: document.getElementById('certProgressBar'),
  certCountPass: document.getElementById('certCountPass'),
  certCountWarn: document.getElementById('certCountWarn'),
  certCountFail: document.getElementById('certCountFail'),
  certCountTotal: document.getElementById('certCountTotal'),
  certTableBody: document.getElementById('certTableBody')
};

// Helper: Log to console UI
function logConsole(method, path, status, latency, data) {
  elements.lastStatusCode.textContent = `${status} ${status >= 200 && status < 300 ? 'OK' : 'ERROR'}`;
  elements.lastStatusCode.style.background = status >= 200 && status < 300 ? 'rgba(16, 185, 129, 0.2)' : 'rgba(239, 68, 68, 0.2)';
  elements.lastStatusCode.style.color = status >= 200 && status < 300 ? '#34d399' : '#f87171';
  elements.lastLatency.textContent = `${latency}ms`;

  const timestamp = new Date().toLocaleTimeString();
  const formattedJSON = typeof data === 'object' ? JSON.stringify(data, null, 2) : data;
  const logMessage = `[${timestamp}] ${method} ${path} (${status}) - ${latency}ms\n${formattedJSON}\n\n`;

  elements.consoleOutput.textContent = logMessage + elements.consoleOutput.textContent;
}

// HTTP Client wrapper
async function apiRequest(endpoint, options = {}) {
  const baseUrl = elements.serverUrl.value.replace(/\/$/, '');
  const url = `${baseUrl}${endpoint}`;
  const method = (options.method || 'GET').toUpperCase();
  
  const headers = {
    ...(options.headers || {})
  };

  if (method !== 'GET' && method !== 'HEAD' && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json';
  }

  if (options.useAuth && state.accessToken) {
    headers['Authorization'] = `Bearer ${state.accessToken}`;
  }

  if (options.useTenant && state.activeTenant) {
    headers['X-Tenant-ID'] = state.activeTenant.id;
  }

  if (options.apiKey) {
    headers['X-API-Key'] = options.apiKey;
  }

  const startTime = performance.now();
  try {
    const res = await fetch(url, {
      method,
      headers,
      body: options.body ? JSON.stringify(options.body) : undefined
    });
    
    const latency = Math.round(performance.now() - startTime);
    let data;
    const contentType = res.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      data = await res.json();
    } else {
      data = await res.text();
    }

    logConsole(method, endpoint, res.status, latency, data);
    return { ok: res.ok, status: res.status, data };
  } catch (err) {
    const latency = Math.round(performance.now() - startTime);
    logConsole(method, endpoint, 0, latency, { error: `Failed to connect to ${url}: ${err.message}` });
    return { ok: false, status: 0, data: { error: err.message } };
  }
}

// Save Session Tokens
function setSessionTokens(accessToken, refreshToken) {
  state.accessToken = accessToken || '';
  state.refreshToken = refreshToken || '';
  if (accessToken) localStorage.setItem('saaskit_access_token', accessToken);
  else localStorage.removeItem('saaskit_access_token');
  if (refreshToken) localStorage.setItem('saaskit_refresh_token', refreshToken);
  else localStorage.removeItem('saaskit_refresh_token');

  updateUIContext();
}

// Update Context Bar UI
function updateUIContext() {
  elements.currentUserEmail.textContent = state.currentUser ? state.currentUser.email : 'Not Logged In';
  elements.tokenPreview.textContent = state.accessToken ? `${state.accessToken.substring(0, 24)}...` : 'None';
  elements.activeTenantName.textContent = state.activeTenant ? `${state.activeTenant.name} (${state.activeTenant.slug})` : 'None';
}

// Check Health
async function checkHealth() {
  const res = await apiRequest('/health');
  if (res.ok) {
    elements.statusBadge.className = 'status-badge online';
    elements.statusText.textContent = 'Connected';
  } else {
    elements.statusBadge.className = 'status-badge offline';
    elements.statusText.textContent = 'Disconnected';
  }
}

// OIDC Certification Compliance Audit Runner
async function runOidcCertificationAudit() {
  elements.certTableBody.innerHTML = '<tr><td colspan="4" class="text-center">Running automated OIDC Certification compliance audit...</td></tr>';
  
  const results = [];
  
  // Test 1: Discovery Endpoint Status
  const discoRes = await apiRequest('/.well-known/openid-configuration');
  if (discoRes.ok && typeof discoRes.data === 'object') {
    results.push({
      id: 'OP-DISCO-01',
      spec: 'OIDC Discovery 1.0 (Clause 4)',
      status: 'PASS',
      note: 'Discovery document returned valid JSON (HTTP 200)'
    });

    const d = discoRes.data;

    // Test 2: Required Discovery Metadata Fields
    const requiredFields = ['issuer', 'authorization_endpoint', 'token_endpoint', 'userinfo_endpoint', 'jwks_uri', 'response_types_supported', 'subject_types_supported', 'id_token_signing_alg_values_supported'];
    const missing = requiredFields.filter(f => !d[f]);
    if (missing.length === 0) {
      results.push({
        id: 'OP-DISCO-02',
        spec: 'OIDC Core 1.0 Section 3.1.2.1',
        status: 'PASS',
        note: 'All required OpenID Discovery metadata attributes present'
      });
    } else {
      results.push({
        id: 'OP-DISCO-02',
        spec: 'OIDC Core 1.0 Section 3.1.2.1',
        status: 'FAIL',
        note: `Missing required metadata fields: ${missing.join(', ')}`
      });
    }

    // Test 3: OpenID Scope Support
    const scopes = d.scopes_supported || [];
    if (scopes.includes('openid')) {
      results.push({
        id: 'OP-SCOPE-01',
        spec: 'OIDC Core 1.0 Section 3.1.2.1',
        status: 'PASS',
        note: `Required scope "openid" supported (found: ${scopes.join(', ')})`
      });
    } else {
      results.push({
        id: 'OP-SCOPE-01',
        spec: 'OIDC Core 1.0 Section 3.1.2.1',
        status: 'FAIL',
        note: 'Mandatory "openid" scope missing from scopes_supported'
      });
    }

    // Test 4: Subject Types
    const subTypes = d.subject_types_supported || [];
    if (subTypes.includes('public') || subTypes.includes('pairwise')) {
      results.push({
        id: 'OP-SUB-01',
        spec: 'OIDC Core 1.0 Section 8',
        status: 'PASS',
        note: `Subject type support: ${subTypes.join(', ')}`
      });
    } else {
      results.push({
        id: 'OP-SUB-01',
        spec: 'OIDC Core 1.0 Section 8',
        status: 'FAIL',
        note: 'No valid subject_types_supported found'
      });
    }

    // Test 5: ID Token Signing Algorithms
    const algs = d.id_token_signing_alg_values_supported || [];
    if (algs.includes('RS256')) {
      results.push({
        id: 'OP-SIG-01',
        spec: 'OIDC Core 1.0 Section 15.1',
        status: 'PASS',
        note: `Mandatory signing algorithm RS256 supported (configured: ${algs.join(', ')})`
      });
    } else {
      results.push({
        id: 'OP-SIG-01',
        spec: 'OIDC Core 1.0 Section 15.1',
        status: 'FAIL',
        note: 'Mandatory RS256 algorithm not declared in id_token_signing_alg_values_supported'
      });
    }

    // Test 6: PKCE Support (RFC 7636)
    const pkceMethods = d.code_challenge_methods_supported || [];
    if (pkceMethods.includes('S256')) {
      results.push({
        id: 'OP-PKCE-01',
        spec: 'RFC 7636 / OIDF FAPI Profile',
        status: 'PASS',
        note: 'PKCE code challenge method S256 supported'
      });
    } else {
      results.push({
        id: 'OP-PKCE-01',
        spec: 'RFC 7636 / OIDF FAPI Profile',
        status: 'WARN',
        note: 'S256 PKCE method not declared in discovery'
      });
    }

    // Test 7: Grant Types & Response Types
    const grants = d.grant_types_supported || [];
    if (grants.includes('authorization_code')) {
      results.push({
        id: 'OP-GRANT-01',
        spec: 'OIDC Core 1.0 Section 3.1',
        status: 'PASS',
        note: `Supported grant types: ${grants.join(', ')}`
      });
    } else {
      results.push({
        id: 'OP-GRANT-01',
        spec: 'OIDC Core 1.0 Section 3.1',
        status: 'FAIL',
        note: 'authorization_code grant type not supported'
      });
    }

  } else {
    results.push({
      id: 'OP-DISCO-01',
      spec: 'OIDC Discovery 1.0',
      status: 'FAIL',
      note: 'Failed to fetch OIDC Discovery document from /.well-known/openid-configuration'
    });
  }

  // Test 8: JWKS Keys Endpoint
  let jwksRes = await apiRequest('/keys');
  if (!jwksRes.ok) jwksRes = await apiRequest('/oidc/keys');
  
  if (jwksRes.ok && jwksRes.data && Array.isArray(jwksRes.data.keys)) {
    if (jwksRes.data.keys.length > 0) {
      const k = jwksRes.data.keys[0];
      results.push({
        id: 'OP-JWKS-01',
        spec: 'OIDC Core 1.0 Section 10.1 (JWKS)',
        status: 'PASS',
        note: `JWKS keys served with key_id "${k.kid || 'active'}", kty "${k.kty || 'RSA'}"`
      });
    } else {
      results.push({
        id: 'OP-JWKS-01',
        spec: 'OIDC Core 1.0 Section 10.1 (JWKS)',
        status: 'WARN',
        note: 'JWKS endpoint returned 200 OK but keys array is empty'
      });
    }
  } else {
    results.push({
      id: 'OP-JWKS-01',
      spec: 'OIDC Core 1.0 Section 10.1 (JWKS)',
      status: 'FAIL',
      note: 'JWKS endpoint unreachable or returned invalid JSON format'
    });
  }

  // Test 9: Endpoints reachability
  const healthRes = await apiRequest('/health');
  if (healthRes.ok) {
    results.push({
      id: 'OP-SYSTEM-01',
      spec: 'SaaS Platform Operations',
      status: 'PASS',
      note: 'Health check endpoint /health returned HTTP 200 OK'
    });
  }

  // Render Table
  let passCount = 0, warnCount = 0, failCount = 0;
  elements.certTableBody.innerHTML = '';
  results.forEach(r => {
    if (r.status === 'PASS') passCount++;
    else if (r.status === 'WARN') warnCount++;
    else failCount++;

    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><strong>${r.id}</strong></td>
      <td>${r.spec}</td>
      <td><span class="status-pill ${r.status.toLowerCase()}">${r.status}</span></td>
      <td>${r.note}</td>
    `;
    elements.certTableBody.appendChild(tr);
  });

  const total = results.length;
  const score = total > 0 ? Math.round((passCount / total) * 100) : 0;

  elements.certScoreVal.textContent = `${score}%`;
  elements.certProgressBar.style.width = `${score}%`;
  elements.certCountPass.textContent = passCount;
  elements.certCountWarn.textContent = warnCount;
  elements.certCountFail.textContent = failCount;
  elements.certCountTotal.textContent = total;
}

// Setup Event Listeners
function setupEvents() {
  // Navigation Tabs
  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
      document.querySelectorAll('.tab-pane').forEach(p => p.classList.remove('active'));
      
      btn.classList.add('active');
      const tabId = `tab-${btn.dataset.tab}`;
      document.getElementById(tabId).classList.add('active');
    });
  });

  // Health
  elements.btnHealth.addEventListener('click', checkHealth);
  elements.btnClearConsole.addEventListener('click', () => {
    elements.consoleOutput.textContent = '';
    elements.lastStatusCode.textContent = 'Ready';
    elements.lastStatusCode.style.background = '#1e293b';
    elements.lastStatusCode.style.color = '#a5b4fc';
    elements.lastLatency.textContent = '0ms';
  });

  // Auth: Register
  elements.btnRegister.addEventListener('click', async () => {
    const email = elements.regEmail.value;
    const password = elements.regPassword.value;
    await apiRequest('/api/v1/auth/register', {
      method: 'POST',
      body: { email, password }
    });
  });

  // Auth: Login
  elements.btnLogin.addEventListener('click', async () => {
    const email = elements.loginEmail.value;
    const password = elements.loginPassword.value;
    const res = await apiRequest('/api/v1/auth/login', {
      method: 'POST',
      body: { email, password }
    });

    if (res.ok && res.data) {
      const tokens = res.data.tokens || res.data;
      const accessToken = tokens.access_token || tokens.accessToken;
      const refreshToken = tokens.refresh_token || tokens.refreshToken;
      if (accessToken) {
        setSessionTokens(accessToken, refreshToken);
        await fetchMe();
      }
    }
  });

  // Auth: Refresh Token
  elements.btnRefresh.addEventListener('click', async () => {
    if (!state.refreshToken) {
      alert('No refresh token available');
      return;
    }
    const res = await apiRequest('/api/v1/auth/refresh', {
      method: 'POST',
      body: { refresh_token: state.refreshToken, refreshToken: state.refreshToken }
    });

    if (res.ok && res.data) {
      const tokens = res.data.tokens || res.data;
      const accessToken = tokens.access_token || tokens.accessToken;
      const refreshToken = tokens.refresh_token || tokens.refreshToken || state.refreshToken;
      if (accessToken) {
        setSessionTokens(accessToken, refreshToken);
      }
    }
  });

  // Auth: Logout
  elements.btnLogout.addEventListener('click', async () => {
    if (state.accessToken) {
      await apiRequest('/api/v1/auth/logout', {
        method: 'POST',
        useAuth: true,
        body: { refresh_token: state.refreshToken }
      });
    }
    setSessionTokens('', '');
    state.currentUser = null;
    state.activeTenant = null;
    updateUIContext();
  });

  // Profile: Get Me
  async function fetchMe() {
    const res = await apiRequest('/api/v1/users/me', { useAuth: true });
    if (res.ok) {
      state.currentUser = res.data;
      if (res.data.name) elements.profileName.value = res.data.name;
      updateUIContext();
    }
    return res;
  }
  elements.btnGetMe.addEventListener('click', fetchMe);

  // Profile: Update Profile Name
  elements.btnUpdateProfile.addEventListener('click', async () => {
    const name = elements.profileName.value;
    await apiRequest('/api/v1/users/me', {
      method: 'PATCH',
      useAuth: true,
      body: { name }
    });
    await fetchMe();
  });

  // Profile: Get Metadata
  elements.btnGetMeta.addEventListener('click', async () => {
    const res = await apiRequest('/api/v1/users/me/metadata', { useAuth: true });
    if (res.ok) {
      if (res.data.metadata_public) elements.publicMeta.value = JSON.stringify(res.data.metadata_public, null, 2);
      if (res.data.metadata_private) elements.privateMeta.value = JSON.stringify(res.data.metadata_private, null, 2);
    }
  });

  // Profile: Save Metadata
  elements.btnUpdateMeta.addEventListener('click', async () => {
    let public_meta = {}, private_meta = {};
    try {
      if (elements.publicMeta.value) public_meta = JSON.parse(elements.publicMeta.value);
      if (elements.privateMeta.value) private_meta = JSON.parse(elements.privateMeta.value);
    } catch (e) {
      alert('Invalid JSON in metadata fields: ' + e.message);
      return;
    }

    await apiRequest('/api/v1/users/me/metadata', {
      method: 'PATCH',
      useAuth: true,
      body: {
        metadata_public: public_meta,
        metadata_private: private_meta
      }
    });
  });

  // Tenants: Create Tenant
  elements.btnCreateTenant.addEventListener('click', async () => {
    const name = elements.tenantName.value;
    const slug = elements.tenantSlug.value;
    const res = await apiRequest('/api/v1/tenants', {
      method: 'POST',
      useAuth: true,
      body: { name, slug }
    });
    if (res.ok) {
      await fetchTenants();
    }
  });

  // Tenants: List Tenants
  async function fetchTenants() {
    const res = await apiRequest('/api/v1/tenants', { useAuth: true });
    if (res.ok) {
      state.tenants = Array.isArray(res.data) ? res.data : (res.data.tenants || []);
      elements.tenantSelect.innerHTML = '<option value="">(Select a tenant)</option>';
      state.tenants.forEach(t => {
        const opt = document.createElement('option');
        opt.value = t.id;
        opt.textContent = `${t.name} (${t.slug})`;
        elements.tenantSelect.appendChild(opt);
      });
      if (state.tenants.length > 0 && !state.activeTenant) {
        state.activeTenant = state.tenants[0];
        elements.tenantSelect.value = state.activeTenant.id;
        updateUIContext();
      }
    }
  }
  elements.btnListTenants.addEventListener('click', fetchTenants);

  // Tenants: Switch Tenant
  elements.btnSwitchTenant.addEventListener('click', async () => {
    const tenantId = elements.tenantSelect.value;
    if (!tenantId) {
      alert('Select a tenant from the dropdown first');
      return;
    }
    const found = state.tenants.find(t => t.id === tenantId);
    if (found) {
      state.activeTenant = found;
      updateUIContext();
      logConsole('CLIENT', 'switch-tenant', 200, 0, { message: 'Switched active tenant context', activeTenant: state.activeTenant });
    }
  });

  // Tenants: Invite Member
  elements.btnInviteMember.addEventListener('click', async () => {
    if (!state.activeTenant) {
      alert('Please select/switch to an active tenant first');
      return;
    }
    const email = elements.inviteEmail.value;
    const role = elements.inviteRole.value;
    await apiRequest(`/api/v1/tenants/${state.activeTenant.id}/members`, {
      method: 'POST',
      useAuth: true,
      body: { email, role }
    });
  });

  // Tenants: List Members
  elements.btnListMembers.addEventListener('click', async () => {
    if (!state.activeTenant) {
      alert('Please select/switch to an active tenant first');
      return;
    }
    await apiRequest(`/api/v1/tenants/${state.activeTenant.id}/members`, { useAuth: true });
  });

  // API Keys: Create
  elements.btnCreateApiKey.addEventListener('click', async () => {
    if (!state.activeTenant) {
      alert('Please select/switch to an active tenant first');
      return;
    }
    const name = elements.apiKeyName.value;
    const type = elements.apiKeyType.value;
    const scopes = elements.apiKeyScopes.value.split(',').map(s => s.trim()).filter(Boolean);
    const res = await apiRequest(`/api/v1/tenants/${state.activeTenant.id}/api-keys`, {
      method: 'POST',
      useAuth: true,
      body: { name, type, scopes }
    });
    if (res.ok && res.data) {
      const fullKey = res.data.full_key || res.data.key;
      elements.testApiKeySecret.value = fullKey;
    }
  });

  // API Keys: List
  elements.btnListApiKeys.addEventListener('click', async () => {
    if (!state.activeTenant) {
      alert('Please select/switch to an active tenant first');
      return;
    }
    await apiRequest(`/api/v1/tenants/${state.activeTenant.id}/api-keys`, { useAuth: true });
  });

  // API Keys: Authenticate via API Key
  elements.btnTestApiKey.addEventListener('click', async () => {
    const key = elements.testApiKeySecret.value;
    if (!key) {
      alert('Enter an API key secret to test');
      return;
    }
    await apiRequest('/api/v1/users/me', { apiKey: key });
  });

  // OIDC: Discovery
  elements.btnOidcDiscovery.addEventListener('click', async () => {
    await apiRequest('/.well-known/openid-configuration');
  });

  // OIDC: JWKS
  elements.btnJwks.addEventListener('click', async () => {
    let res = await apiRequest('/keys');
    if (!res.ok) await apiRequest('/oidc/keys');
  });

  // OIDC: Launch Authorize
  elements.btnLaunchOidc.addEventListener('click', () => {
    const baseUrl = elements.serverUrl.value.replace(/\/$/, '');
    const clientId = elements.oidcClientId.value;
    const redirectUri = elements.oidcRedirectUri.value;
    const url = `${baseUrl}/oidc/authorize?response_type=code&client_id=${encodeURIComponent(clientId)}&redirect_uri=${encodeURIComponent(redirectUri)}&scope=openid%20profile%20email`;
    window.open(url, '_blank');
  });

  // OIDC Certification Suite
  elements.btnRunCertAudit.addEventListener('click', runOidcCertificationAudit);

  // Initial check
  checkHealth();
  updateUIContext();
}

document.addEventListener('DOMContentLoaded', setupEvents);
