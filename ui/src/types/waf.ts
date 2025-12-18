export interface WAFRule {
  id: number;
  category: string;
  description?: string;
  severity?: string;
  tags?: string[];
  enabled: boolean;
  globally_disabled?: boolean;
  exclusion?: WAFRuleExclusion;
  global_exclusion?: GlobalWAFRuleExclusion;
}

export interface WAFRuleCategory {
  id: string;
  name: string;
  description: string;
  file_name: string;
  rule_count: number;
  rules?: WAFRule[];
}

export interface WAFRulesResponse {
  categories: WAFRuleCategory[];
  total_rules: number;
}

export interface WAFRuleExclusion {
  id: string;
  proxy_host_id: string;
  rule_id: number;
  rule_category?: string;
  rule_description?: string;
  reason?: string;
  disabled_by?: string;
  created_at: string;
}

export interface WAFHostConfig {
  proxy_host_id: string;
  proxy_host_name: string;
  waf_enabled: boolean;
  waf_mode: string;
  exclusions?: WAFRuleExclusion[];
  exclusion_count: number;
}

export interface WAFHostConfigListResponse {
  hosts: WAFHostConfig[];
  total: number;
}

export interface CreateWAFRuleExclusionRequest {
  rule_id: number;
  rule_category?: string;
  rule_description?: string;
  reason?: string;
}

export interface WAFPolicyHistory {
  id: string;
  proxy_host_id: string;
  rule_id: number;
  rule_category?: string;
  rule_description?: string;
  action: 'disabled' | 'enabled';
  reason?: string;
  changed_by?: string;
  created_at: string;
}

export interface WAFPolicyHistoryResponse {
  history: WAFPolicyHistory[];
  total: number;
}

// Global WAF Rule Exclusions
export interface GlobalWAFRuleExclusion {
  id: string;
  rule_id: number;
  rule_category?: string;
  rule_description?: string;
  reason?: string;
  disabled_by?: string;
  created_at: string;
  updated_at: string;
}

export interface GlobalWAFRule extends Omit<WAFRule, 'exclusion'> {
  globally_disabled: boolean;
  global_exclusion?: GlobalWAFRuleExclusion;
}

export interface GlobalWAFRuleCategory {
  id: string;
  name: string;
  description: string;
  file_name: string;
  rule_count: number;
  rules?: GlobalWAFRule[];
}

export interface GlobalWAFRulesResponse {
  categories: GlobalWAFRuleCategory[];
  total_rules: number;
  global_exclusions: GlobalWAFRuleExclusion[];
}

export interface CreateGlobalWAFRuleExclusionRequest {
  rule_id: number;
  rule_category?: string;
  rule_description?: string;
  reason?: string;
}

export interface GlobalWAFPolicyHistory {
  id: string;
  rule_id: number;
  rule_category?: string;
  rule_description?: string;
  action: 'disabled' | 'enabled';
  reason?: string;
  changed_by?: string;
  created_at: string;
}

export interface GlobalWAFPolicyHistoryResponse {
  history: GlobalWAFPolicyHistory[];
  total: number;
}
