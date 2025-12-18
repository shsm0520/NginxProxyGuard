import { apiGet, apiPut, apiDelete } from './client';

export interface ChallengeConfig {
  id: string;
  proxy_host_id?: string;
  enabled: boolean;
  challenge_type: 'recaptcha_v2' | 'recaptcha_v3' | 'turnstile';
  site_key: string;
  has_secret_key: boolean;
  token_validity: number;
  min_score: number;
  apply_to: string;
  page_title: string;
  page_message: string;
  theme: 'light' | 'dark';
  created_at: string;
  updated_at: string;
}

export interface ChallengeConfigRequest {
  enabled?: boolean;
  challenge_type?: string;
  site_key?: string;
  secret_key?: string;
  token_validity?: number;
  min_score?: number;
  apply_to?: string;
  page_title?: string;
  page_message?: string;
  theme?: string;
}

export interface ChallengeStats {
  total_challenges: number;
  passed_challenges: number;
  failed_challenges: number;
  active_tokens: number;
  average_score?: number;
  average_solve_time?: number;
}

// Global challenge config
export async function getGlobalChallengeConfig(): Promise<ChallengeConfig> {
  return apiGet<ChallengeConfig>('/api/v1/challenge-config');
}

export async function updateGlobalChallengeConfig(data: ChallengeConfigRequest): Promise<ChallengeConfig> {
  return apiPut<ChallengeConfig>('/api/v1/challenge-config', data);
}

// Per-host challenge config
export async function getHostChallengeConfig(proxyHostId: string): Promise<ChallengeConfig> {
  return apiGet<ChallengeConfig>(`/api/v1/proxy-hosts/${proxyHostId}/challenge-config`);
}

export async function updateHostChallengeConfig(proxyHostId: string, data: ChallengeConfigRequest): Promise<ChallengeConfig> {
  return apiPut<ChallengeConfig>(`/api/v1/proxy-hosts/${proxyHostId}/challenge-config`, data);
}

export async function deleteHostChallengeConfig(proxyHostId: string): Promise<void> {
  return apiDelete(`/api/v1/proxy-hosts/${proxyHostId}/challenge-config`);
}

// Challenge stats
export async function getChallengeStats(proxyHostId?: string): Promise<ChallengeStats> {
  const params = proxyHostId ? `?proxy_host_id=${proxyHostId}` : '';
  return apiGet<ChallengeStats>(`/api/v1/challenge-config/stats${params}`);
}
