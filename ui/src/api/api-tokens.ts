import { getToken } from './auth';

const API_BASE = '/api/v1';

export interface APIToken {
  id: string;
  name: string;
  token_prefix: string;
  permissions: string[];
  allowed_ips?: string[];
  rate_limit?: number;
  expires_at?: string;
  last_used_at?: string;
  last_used_ip?: string;
  use_count: number;
  is_active: boolean;
  is_expired: boolean;
  created_at: string;
  username?: string;
}

export interface APITokenWithSecret extends APIToken {
  token: string; // Only returned once on creation
}

export interface CreateAPITokenRequest {
  name: string;
  permissions: string[];
  allowed_ips?: string[];
  rate_limit?: number;
  expires_in?: string; // e.g., "30d", "1y", "never"
}

export interface UpdateAPITokenRequest {
  name?: string;
  permissions?: string[];
  allowed_ips?: string[];
  rate_limit?: number;
  is_active?: boolean;
}

export interface PermissionsResponse {
  permissions: string[];
  groups: Record<string, string[]>;
}

export interface TokenUsage {
  id: string;
  token_id: string;
  endpoint: string;
  method: string;
  status_code: number;
  client_ip: string;
  user_agent: string;
  request_body_size: number;
  response_time_ms: number;
  created_at: string;
}

export interface TokenUsageResponse {
  token: APIToken;
  usages: TokenUsage[];
}

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = getToken();
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const response = await fetch(url, { ...options, headers });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(error.error || 'Request failed');
  }

  return response.json();
}

// Get all API tokens for current user
export async function getAPITokens(all = false): Promise<APIToken[]> {
  const url = `${API_BASE}/api-tokens${all ? '?all=true' : ''}`;
  return fetchWithAuth(url);
}

// Create a new API token
export async function createAPIToken(req: CreateAPITokenRequest): Promise<APITokenWithSecret> {
  return fetchWithAuth(`${API_BASE}/api-tokens`, {
    method: 'POST',
    body: JSON.stringify(req),
  });
}

// Get a specific token
export async function getAPIToken(id: string): Promise<APIToken> {
  return fetchWithAuth(`${API_BASE}/api-tokens/${id}`);
}

// Update a token
export async function updateAPIToken(id: string, req: UpdateAPITokenRequest): Promise<APIToken> {
  return fetchWithAuth(`${API_BASE}/api-tokens/${id}`, {
    method: 'PUT',
    body: JSON.stringify(req),
  });
}

// Revoke a token
export async function revokeAPIToken(id: string, reason?: string): Promise<void> {
  await fetchWithAuth(`${API_BASE}/api-tokens/${id}/revoke`, {
    method: 'POST',
    body: JSON.stringify({ reason }),
  });
}

// Delete a token
export async function deleteAPIToken(id: string): Promise<void> {
  await fetchWithAuth(`${API_BASE}/api-tokens/${id}`, {
    method: 'DELETE',
  });
}

// Get token usage statistics
export async function getTokenUsage(id: string): Promise<TokenUsageResponse> {
  return fetchWithAuth(`${API_BASE}/api-tokens/${id}/usage`);
}

// Get available permissions
export async function getPermissions(): Promise<PermissionsResponse> {
  return fetchWithAuth(`${API_BASE}/api-tokens/permissions`);
}

// Permission labels for display
export const permissionLabels: Record<string, string> = {
  '*': 'All Permissions',
  'proxy:read': 'Proxy Hosts (Read)',
  'proxy:write': 'Proxy Hosts (Write)',
  'proxy:delete': 'Proxy Hosts (Delete)',
  'certificate:read': 'Certificates (Read)',
  'certificate:write': 'Certificates (Write)',
  'certificate:delete': 'Certificates (Delete)',
  'waf:read': 'WAF (Read)',
  'waf:write': 'WAF (Write)',
  'logs:read': 'Logs (Read)',
  'settings:read': 'Settings (Read)',
  'settings:write': 'Settings (Write)',
  'backup:read': 'Backups (Read)',
  'backup:create': 'Backups (Create)',
  'backup:restore': 'Backups (Restore)',
  'user:read': 'Users (Read)',
};

// Permission group labels
export const permissionGroupLabels: Record<string, { label: string; description: string }> = {
  read_only: {
    label: 'Read Only',
    description: '모든 리소스를 조회할 수 있지만 변경할 수 없습니다.',
  },
  operator: {
    label: 'Operator',
    description: 'Proxy Host, Certificate, WAF 등을 관리할 수 있습니다.',
  },
  admin: {
    label: 'Full Access',
    description: '모든 권한을 가집니다.',
  },
};
