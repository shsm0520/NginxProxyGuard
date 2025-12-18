import type {
  Certificate,
  CreateCertificateRequest,
  UploadCertificateRequest,
  CertificateListResponse,
  CertificateLogResponse,
  CertificateHistoryListResponse,
} from '../types/certificate';
import { getAuthHeaders } from './auth';

const API_BASE = '/api/v1';

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }
  return response.json();
}

export async function listCertificates(page = 1, perPage = 20): Promise<CertificateListResponse> {
  const response = await fetch(`${API_BASE}/certificates?page=${page}&per_page=${perPage}`, {
    headers: getAuthHeaders(),
  });
  return handleResponse<CertificateListResponse>(response);
}

export async function getCertificate(id: string): Promise<Certificate> {
  const response = await fetch(`${API_BASE}/certificates/${id}`, {
    headers: getAuthHeaders(),
  });
  return handleResponse<Certificate>(response);
}

export async function createCertificate(data: CreateCertificateRequest): Promise<Certificate> {
  const response = await fetch(`${API_BASE}/certificates`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  return handleResponse<Certificate>(response);
}

export async function uploadCertificate(data: UploadCertificateRequest): Promise<Certificate> {
  const response = await fetch(`${API_BASE}/certificates/upload`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  return handleResponse<Certificate>(response);
}

export async function deleteCertificate(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/certificates/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }
}

export async function renewCertificate(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/certificates/${id}/renew`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }
}

export async function getExpiringCertificates(days = 30): Promise<Certificate[]> {
  const response = await fetch(`${API_BASE}/certificates/expiring?days=${days}`, {
    headers: getAuthHeaders(),
  });
  const data = await handleResponse<{ data: Certificate[] }>(response);
  return data.data;
}

export async function getCertificateLogs(id: string): Promise<CertificateLogResponse> {
  const response = await fetch(`${API_BASE}/certificates/${id}/logs`, {
    headers: getAuthHeaders(),
  });
  return handleResponse<CertificateLogResponse>(response);
}

export async function getCertificateHistory(page = 1, perPage = 20, certificateId?: string): Promise<CertificateHistoryListResponse> {
  let url = `${API_BASE}/certificates/history?page=${page}&per_page=${perPage}`;
  if (certificateId) {
    url += `&certificate_id=${certificateId}`;
  }
  const response = await fetch(url, {
    headers: getAuthHeaders(),
  });
  return handleResponse<CertificateHistoryListResponse>(response);
}
