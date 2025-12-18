import { getAuthHeaders } from './auth';

const API_BASE = '/api/v1';

export interface SystemLog {
  id: string;
  source: string;
  level: string;
  message: string;
  details?: Record<string, unknown>;
  container_name?: string;
  component?: string;
  created_at: string;
}

export interface SystemLogFilter {
  source?: string;
  level?: string;
  container?: string;
  component?: string;
  search?: string;
  start_time?: string;
  end_time?: string;
}

export interface SystemLogListResponse {
  logs: SystemLog[];
  total: number;
  limit: number;
  offset: number;
}

export interface SystemLogStats {
  by_source: Record<string, number>;
  by_level: Record<string, number>;
  total: number;
  last_24h: number;
}

export interface LogSource {
  value: string;
  label: string;
}

export interface LogLevel {
  value: string;
  label: string;
  color: string;
}

export async function fetchSystemLogs(
  limit: number = 100,
  offset: number = 0,
  filter?: SystemLogFilter
): Promise<SystemLogListResponse> {
  const params = new URLSearchParams();
  params.set('limit', limit.toString());
  params.set('offset', offset.toString());

  if (filter?.source) params.set('source', filter.source);
  if (filter?.level) params.set('level', filter.level);
  if (filter?.container) params.set('container', filter.container);
  if (filter?.search) params.set('search', filter.search);

  const res = await fetch(`${API_BASE}/system-logs?${params}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch system logs');
  return res.json();
}

export async function fetchSystemLogStats(): Promise<SystemLogStats> {
  const res = await fetch(`${API_BASE}/system-logs/stats`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch system log stats');
  return res.json();
}

export async function fetchLogSources(): Promise<LogSource[]> {
  const res = await fetch(`${API_BASE}/system-logs/sources`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch log sources');
  return res.json();
}

export async function fetchLogLevels(): Promise<LogLevel[]> {
  const res = await fetch(`${API_BASE}/system-logs/levels`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch log levels');
  return res.json();
}

export async function cleanupSystemLogs(retentionDays: number = 7): Promise<{ deleted: number }> {
  const res = await fetch(`${API_BASE}/system-logs/cleanup?retention_days=${retentionDays}`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to cleanup system logs');
  return res.json();
}
