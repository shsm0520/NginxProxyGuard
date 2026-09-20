// E2E verification that ordinary access log lines survive the full pipeline:
// HTTP request → nginx writes /etc/nginx/logs/access_raw.log → LogCollector
// file-tail reads the line → parser populates fields → DB insert →
// /api/v1/logs returns the row.
//
// This is the missing companion to waf-audit-format.spec.ts. Previous silent
// failures (issues #141 / #144 / #145 / #146) were specifically about access
// logs not landing in the DB despite the WAF audit path still working — so
// the WAF spec alone was insufficient regression cover. This spec exercises
// the path the WAF spec doesn't: a clean 2xx GET that should produce exactly
// one access-log row with block_reason='none'.
//
// Field shape (verified against api/internal/model/log.go + repository/log.go):
//   log_type      ('access')
//   status_code   (200 for the GET below)
//   request_uri   (matches the spoofed path)
//   client_ip     (from spoofed X-Forwarded-For; set_real_ip_from trusts 127.0.0.1)
//   host          (from spoofed Host header)
//   block_reason  ('none' — this is the explicit invariant the silent-failure
//                  pattern would break, returning either no row at all OR a
//                  row with an unexpected block_reason)

import { test, expect } from '@playwright/test';
import { APIHelper } from '../../utils/api-helper';
import { pollForLog, triggerUntil } from '../../utils/log-helper';
import { TestDataFactory } from '../../utils/test-data-factory';

test.describe('Access log file-tail pipeline ingestion', () => {
  let api: APIHelper;

  test.beforeEach(async ({ request }) => {
    api = new APIHelper(request);
    await api.login();
  });

  test.afterEach(async () => {
    await api.cleanupTestHosts();
  });

  test('plain GET lands as log_type=access with block_reason=none', async () => {
    const host = TestDataFactory.generateDomain('access-pipeline');
    await api.createProxyHost({
      domain_names: [host],
      forward_host: '127.0.0.1',
      forward_port: 19080, // api container — guaranteed reachable inside e2e net
      forward_scheme: 'http',
      block_exploits: false,
      waf_enabled: false,
      enabled: true,
    });
    // Status itself isn't the invariant — we just need nginx to have processed
    // the request enough to write an access log line. 502 is also acceptable
    // (upstream may not actually answer); what matters is the log row.
    //
    // Waiting on that same set is what tells us the reload landed: until the
    // new server block is live the proxy answers from the catch-all instead,
    // and the request never reaches this host's access log at all.
    const servedByThisHost = (status: number) => [200, 502, 504, 404].includes(status);

    const probePath = `/_npg_pipeline_${Date.now()}`;
    const result = await triggerUntil(
      { host, path: probePath, xForwardedFor: '203.0.113.42' },
      r => servedByThisHost(r.status),
      { describe: `${host} to be served by its own server block` },
    );
    expect([200, 502, 504, 404]).toContain(result.status);

    // Matched on the domain, not on the host id. The id is not available yet
    // and the test was asking for something the collector does not promise:
    // it resolves domain -> host_id from a map rebuilt on a 30s ticker
    // (log_collector.go), and a plain miss returns "" rather than looking the
    // domain up. So a host created just after a tick is genuinely unattributed
    // for up to 30s, and polling its id for 15s could only ever fail-then-pass
    // on retry — which is what it did, on this spec, before and after the
    // trigger rework. The access-log line always carries the host name, which
    // is what this test is actually about.
    const row = await pollForLog(api, {
      host,
      expectedLogType: 'access',
      uriContains: probePath,
      timeoutMs: 15000,
    });

    expect(row.block_reason ?? 'none').toBe('none');
    expect(row.request_uri).toContain(probePath);
    expect(row.client_ip).toBe('203.0.113.42');
  });

  test('/api/v1/health/detailed surfaces collector state', async ({ request }) => {
    const token = await api.getToken();
    const resp = await request.get('/api/v1/health/detailed', {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();

    // Schema invariants — these fields are the operator's only window into
    // silent failures. If we ever drop them by mistake, future regressions
    // (issues #141/#144/#145/#146) become invisible again.
    expect(body.log_collector).toBeDefined();
    expect(typeof body.log_collector.access_log_path_configured).toBe('string');
    expect(typeof body.log_collector.access_log_path_actual).toBe('string');
    expect(typeof body.log_collector.fallback_active).toBe('boolean');
    expect(typeof body.log_collector.last_flush_seconds_ago).toBe('number');
    // The boot probe field added in v2.18.1 — guarantees the silent-failure
    // signal is wired all the way to the JSON response.
    expect(typeof body.log_collector.no_flush_since_start).toBe('boolean');
  });
});
