import { AccessList, CreateAccessListRequest, RedirectHost, CreateRedirectHostRequest, GeoRestriction, CreateGeoRestrictionRequest } from '../types/access';
import { getAuthHeaders } from './auth';

const API_BASE = '/api/v1';

// Access Lists
export async function getAccessLists(page = 1, perPage = 20): Promise<{ data: AccessList[]; total: number }> {
  const res = await fetch(`${API_BASE}/access-lists?page=${page}&per_page=${perPage}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch access lists');
  return res.json();
}

export async function getAccessList(id: string): Promise<AccessList> {
  const res = await fetch(`${API_BASE}/access-lists/${id}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch access list');
  return res.json();
}

export async function createAccessList(data: CreateAccessListRequest): Promise<AccessList> {
  const res = await fetch(`${API_BASE}/access-lists`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || 'Failed to create access list');
  }
  return res.json();
}

export async function updateAccessList(id: string, data: Partial<CreateAccessListRequest>): Promise<AccessList> {
  const res = await fetch(`${API_BASE}/access-lists/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || 'Failed to update access list');
  }
  return res.json();
}

export async function deleteAccessList(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/access-lists/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to delete access list');
}

// Redirect Hosts
export async function getRedirectHosts(page = 1, perPage = 20): Promise<{ data: RedirectHost[]; total: number }> {
  const res = await fetch(`${API_BASE}/redirect-hosts?page=${page}&per_page=${perPage}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch redirect hosts');
  return res.json();
}

export async function getRedirectHost(id: string): Promise<RedirectHost> {
  const res = await fetch(`${API_BASE}/redirect-hosts/${id}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch redirect host');
  return res.json();
}

export async function createRedirectHost(data: CreateRedirectHostRequest): Promise<RedirectHost> {
  const res = await fetch(`${API_BASE}/redirect-hosts`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || 'Failed to create redirect host');
  }
  return res.json();
}

export async function updateRedirectHost(id: string, data: Partial<CreateRedirectHostRequest>): Promise<RedirectHost> {
  const res = await fetch(`${API_BASE}/redirect-hosts/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || 'Failed to update redirect host');
  }
  return res.json();
}

export async function deleteRedirectHost(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/redirect-hosts/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to delete redirect host');
}

// Geo Restrictions
export async function getGeoRestriction(proxyHostId: string): Promise<GeoRestriction> {
  const res = await fetch(`${API_BASE}/proxy-hosts/${proxyHostId}/geo`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch geo restriction');
  return res.json();
}

export async function setGeoRestriction(proxyHostId: string, data: CreateGeoRestrictionRequest, skipReload = false): Promise<GeoRestriction> {
  const url = skipReload
    ? `${API_BASE}/proxy-hosts/${proxyHostId}/geo?skip_reload=true`
    : `${API_BASE}/proxy-hosts/${proxyHostId}/geo`;
  const res = await fetch(url, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || 'Failed to set geo restriction');
  }
  return res.json();
}

export async function deleteGeoRestriction(proxyHostId: string, skipReload = false): Promise<void> {
  const url = skipReload
    ? `${API_BASE}/proxy-hosts/${proxyHostId}/geo?skip_reload=true`
    : `${API_BASE}/proxy-hosts/${proxyHostId}/geo`;
  const res = await fetch(url, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to delete geo restriction');
}

export async function getCountryCodes(): Promise<Record<string, string>> {
  const res = await fetch(`${API_BASE}/geo/countries`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch country codes');
  return res.json();
}
