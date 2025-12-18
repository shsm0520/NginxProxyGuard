import { getAuthHeaders } from './auth';

const API_BASE = '/api/v1';

export interface AttackPattern {
  id: string;
  category: string;
  description: string;
}

export interface WAFTestResult {
  attack_type: string;
  test_url: string;
  status_code: number;
  blocked: boolean;
  response_time_ms: number;
  description: string;
}

export async function fetchAttackPatterns(): Promise<AttackPattern[]> {
  const res = await fetch(`${API_BASE}/waf-test/patterns`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch attack patterns');
  return res.json();
}

export async function testAttack(
  targetUrl: string,
  attackType: string,
  hostHeader?: string
): Promise<WAFTestResult> {
  const body: Record<string, string> = {
    target_url: targetUrl,
    attack_type: attackType
  };
  if (hostHeader) {
    body.host_header = hostHeader;
  }
  const res = await fetch(`${API_BASE}/waf-test/test`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error('Failed to execute test');
  return res.json();
}

export async function testAllAttacks(
  targetUrl: string,
  hostHeader?: string
): Promise<WAFTestResult[]> {
  const body: Record<string, string> = { target_url: targetUrl };
  if (hostHeader) {
    body.host_header = hostHeader;
  }
  const res = await fetch(`${API_BASE}/waf-test/test-all`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error('Failed to execute tests');
  return res.json();
}
