// End-to-end regression for the block_reason ingestion pipeline.
//
// Why this file exists
// --------------------
// Issues #130, #133, #134, #137 plus the recent fixes in 38802ae/f0be478 all
// share a single failure mode: the security layer correctly *blocks* a request,
// but the resulting access-log row arrives in the DB with block_reason="none"
// (or wrong reason). Unit tests covering the nginx template (M3.5) and the
// log_collector parser (M3.6) catch their own slice — this spec is the seam
// that closes the loop:
//
//     request → nginx → access log → log_collector → DB → /api/v1/logs
//
// For each security feature we (1) provision a fresh single-purpose proxy
// host, (2) activate the feature via the dedicated APIHelper method, (3) fire
// the exact request that should trip it, and (4) poll /api/v1/logs until a row
// appears, asserting the row carries the expected block_reason plus any
// feature-specific metadata (bot_category, exploit_rule, geo_country_code).
//
// Per-test isolation
// ------------------
// Each case creates a unique host domain. State that survives host deletion
// (banned IPs scoped per-host, custom exploit rules which are global) is
// tracked on APIHelper and rolled back in afterEach. We rely on the test mmdb
// vendored at fixtures/geoip-test.mmdb — see global-setup.ts.
//
// GeoIP IPs are taken from the MaxMind GeoLite2-Country-Test database. We
// verified at fixture-build time which IPs resolve to which country code:
//   81.2.69.144   → GB
//   89.160.20.112 → SE
//   67.43.156.1   → BT
// KR is *not* present in the synthetic mmdb (that's why the original plan's
// 203.243.0.1 falls back to "--"). Tests use GB instead.

import { test, expect } from '@playwright/test';
import { APIHelper } from '../../utils/api-helper';
import { pollForLog, triggerRequest } from '../../utils/log-helper';
import { TestDataFactory } from '../../utils/test-data-factory';

// IPs that the synthetic GeoLite2-Country-Test mmdb actually resolves.
const GEO_IP_GB = '81.2.69.144';
const GEO_IP_SE = '89.160.20.112';

// Helpers below build a fresh host for each test so cases run in parallel
// without colliding on domain, geo restriction state, or exploit rule scope.
async function createIsolatedHost(api: APIHelper, prefix: string) {
  const host = await api.createProxyHost({
    domain_names: [TestDataFactory.generateDomain(`br-${prefix}`)],
    forward_scheme: 'http',
    // Forward to an unreachable port — we *want* the request to fall through to
    // 502/403/etc from the security layer, never reach an upstream that could
    // mask the block.
    forward_host: '127.0.0.1',
    forward_port: 1,
    enabled: true,
  });
  return host;
}

// Wait for nginx to reload after a configuration change.
// The backend uses debounced reload (NginxReloaderDebounce = 2s) plus `nginx -t`
// and reload execution (~1s). A fixed 800ms wait is insufficient and can lead to
// race conditions where requests hit nginx before the new rules or reloads land.
// To ensure host configuration is active and ready before final verification,
// we poll until the expected block status/reason is observed or timeout.
async function waitForReload(
  api: APIHelper,
  opts: {
    host: string;
    path: string;
    expectedStatus: number;
    expectedBlockReason: string;
    xForwardedFor?: string;
    userAgent?: string;
    method?: string;
    timeoutMs?: number;
    intervalMs?: number;
  }
) {
  const timeoutMs = opts.timeoutMs ?? 10000;
  const intervalMs = opts.intervalMs ?? 300;
  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    const res = triggerRequest({
      host: opts.host,
      path: opts.path,
      method: opts.method,
      userAgent: opts.userAgent,
      xForwardedFor: opts.xForwardedFor,
    });

    if (res.status === opts.expectedStatus) {
      // Status matches. Now verify if a matching access log entry with expected
      // block_reason has been ingested into DB.
      try {
        const rows = await api.getLogs({
          host: opts.host,
          limit: 10,
        });
        const match = rows.find(r =>
          r.status_code === opts.expectedStatus &&
          (!opts.expectedBlockReason || r.block_reason === opts.expectedBlockReason) &&
          (!opts.path || (r.request_uri ?? '').includes(opts.path.split('?')[0]))
        );
        if (match) {
          return match;
        }
      } catch {
        // Transient API query issue during polling; retry until deadline.
      }
    }

    await new Promise(res => setTimeout(res, intervalMs));
  }

  throw new Error(
    `waitForReload timed out after ${timeoutMs}ms for host=${opts.host} path=${opts.path} expectedStatus=${opts.expectedStatus} expectedBlockReason=${opts.expectedBlockReason}`
  );
}

test.describe('block_reason regression — security layer to log pipeline', () => {
  let api: APIHelper;

  test.beforeEach(async ({ request }) => {
    api = new APIHelper(request);
    await api.login();
  });

  test.afterEach(async () => {
    // Roll back banned IPs + exploit rules (host-scoped state is cleaned up by
    // deleteProxyHost in each test). Order matters: rules before hosts so the
    // host's block_exploits flag stays consistent.
    await api.cleanupBlockReasonState();
    await api.cleanupTestHosts();
    await api.cleanupTestAccessLists();
  });

  test('case 1: geo_block via blacklist mode sets block_reason=geo_block', async () => {
    const host = await createIsolatedHost(api, 'geo-bl');
    await api.setGeoRestriction(host.id, { mode: 'blacklist', countries: ['GB'] });
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case1',
      expectedStatus: 403,
      expectedBlockReason: 'geo_block',
      xForwardedFor: GEO_IP_GB,
    });

    // Control request: unblocked path/IP must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case1-control',
      xForwardedFor: GEO_IP_SE,
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.geo_country_code).toBe('GB');
    expect(row.status_code).toBe(403);
  });

  test('case 2: geo_block via whitelist mode sets block_reason=geo_block', async () => {
    const host = await createIsolatedHost(api, 'geo-wl');
    // Whitelist SE only — GB request must be blocked.
    await api.setGeoRestriction(host.id, { mode: 'whitelist', countries: ['SE'] });
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case2',
      expectedStatus: 403,
      expectedBlockReason: 'geo_block',
      xForwardedFor: GEO_IP_GB,
    });

    // Control request: allowed country must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case2-control',
      xForwardedFor: GEO_IP_SE,
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.geo_country_code).toBe('GB');
  });

  test('case 3: geo_challenge_mode still records block_reason=geo_block', async () => {
    const host = await createIsolatedHost(api, 'geo-ch');
    // challenge_mode=true should redirect to CAPTCHA (302) instead of 403, but
    // block_reason MUST still be geo_block — that's the regression vector.
    await api.setGeoRestriction(host.id, {
      mode: 'blacklist',
      countries: ['GB'],
      challengeMode: true,
    });
    // Challenge mode issues 302 or 200 with CAPTCHA page.
    const deadline = Date.now() + 10000;
    let row: LogRow | undefined;
    while (Date.now() < deadline) {
      const res = triggerRequest({
        host: host.domain_names[0],
        path: '/case3',
        xForwardedFor: GEO_IP_GB,
      });
      if (res.status !== 502) {
        try {
          const rows = await api.getLogs({ host: host.domain_names[0], limit: 10 });
          const match = rows.find(r => r.block_reason === 'geo_block' && (r.request_uri ?? '').includes('/case3'));
          if (match) {
            row = match;
            break;
          }
        } catch {
          // retry
        }
      }
      await new Promise(res => setTimeout(res, 300));
    }
    if (!row) {
      throw new Error('case 3: timed out waiting for geo_block log row in challenge mode');
    }

    expect(row.geo_country_code).toBe('GB');
  });

  test('case 4: access_denied is set when access-list deny matches', async () => {
    const host = await createIsolatedHost(api, 'access');
    await api.setAccessList(host.id, [
      { directive: 'deny', address: '10.255.255.42' },
      { directive: 'allow', address: 'all' },
    ]);
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case4',
      expectedStatus: 403,
      expectedBlockReason: 'access_denied',
      xForwardedFor: '10.255.255.42',
    });

    // Control request: allowed IP must not be denied
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case4-control',
      xForwardedFor: '10.255.255.43',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.status_code).toBe(403);
    expect(row.client_ip).toBe('10.255.255.42');
  });

  test('case 5: exploit_block fires on path-traversal query string (LFI rule)', async () => {
    const host = await createIsolatedHost(api, 'exp-lfi');
    await api.enableBlockExploits(host.id);
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case5?file=../../etc/passwd',
      expectedStatus: 403,
      expectedBlockReason: 'exploit_block',
    });

    // Control request: benign query string must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case5-control?file=normal.txt',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.status_code).toBe(403);
    expect(row.exploit_rule).toBeTruthy();
  });

  test('case 6: exploit_block fires on dotenv request_uri rule', async () => {
    const host = await createIsolatedHost(api, 'exp-uri');
    await api.enableBlockExploits(host.id);
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case6/.env',
      expectedStatus: 403,
      expectedBlockReason: 'exploit_block',
    });

    // Control request: benign path must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case6/index.html',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.exploit_rule).toBeTruthy();
  });

  test('case 7: exploit_block fires on scanner user agent (sqlmap)', async () => {
    const host = await createIsolatedHost(api, 'exp-ua');
    await api.enableBlockExploits(host.id);
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case7',
      userAgent: 'sqlmap/1.7.2#stable (http://sqlmap.org)',
      expectedStatus: 403,
      expectedBlockReason: 'exploit_block',
    });

    // Control request: normal UA must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case7-control',
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.exploit_rule).toBeTruthy();
    expect(row.bot_category).toBe('scanner');
  });

  test('case 8: exploit_block fires on TRACE method (seeded Dangerous Methods rule)', async () => {
    const host = await createIsolatedHost(api, 'exp-trace');
    await api.enableBlockExploits(host.id);
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case8',
      method: 'TRACE',
      expectedStatus: 405,
      expectedBlockReason: 'exploit_block',
    });

    // Control request: GET method must not return 405
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case8-control',
      method: 'GET',
    });
    expect(controlRes.status).not.toBe(405);

    expect(row.status_code).toBe(405);
  });

  test('case 9: banned_ip blocks request and records block_reason=banned_ip', async () => {
    const host = await createIsolatedHost(api, 'banned');
    const bannedIp = '10.255.255.99';
    await api.setBannedIPs(host.id, [bannedIp]);
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case9',
      xForwardedFor: bannedIp,
      expectedStatus: 403,
      expectedBlockReason: 'banned_ip',
    });

    // Control request: unbanned IP must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case9-control',
      xForwardedFor: '10.255.255.100',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.client_ip).toBe(bannedIp);
  });

  test('case 10: uri_block prefix-matches blocked path', async () => {
    const host = await createIsolatedHost(api, 'uri');
    await api.setURIBlock(host.id, '/case10-admin', 'prefix');
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case10-admin/dashboard',
      expectedStatus: 403,
      expectedBlockReason: 'uri_block',
    });

    // Control request: unblocked URI path must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case10-public/dashboard',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.status_code).toBe(403);
  });

  test('case 11: bot_filter blocks bad-bot UA (AhrefsBot)', async () => {
    const host = await createIsolatedHost(api, 'bot-bad');
    await api.setBotFilter(host.id, { blockBadBots: true });
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case11',
      userAgent: 'AhrefsBot/7.0',
      expectedStatus: 403,
      expectedBlockReason: 'bot_filter',
    });

    // Control request: standard browser UA must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case11-control',
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.bot_category).toBe('bad_bot');
  });

  test('case 12: bot_filter blocks AI-bot UA (GPTBot)', async () => {
    const host = await createIsolatedHost(api, 'bot-ai');
    await api.setBotFilter(host.id, { blockAiBots: true });
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case12',
      userAgent: 'Mozilla/5.0 (compatible; GPTBot/1.0; +https://openai.com/gptbot)',
      expectedStatus: 403,
      expectedBlockReason: 'bot_filter',
    });

    // Control request: standard browser UA must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case12-control',
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.bot_category).toBe('ai_bot');
  });

  test('case 13: bot_filter blocks suspicious client (curl)', async () => {
    const host = await createIsolatedHost(api, 'bot-susp');
    await api.setBotFilter(host.id, { blockSuspiciousClients: true });
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case13',
      userAgent: 'curl/7.88.0',
      expectedStatus: 403,
      expectedBlockReason: 'bot_filter',
    });

    // Control request: standard browser UA must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case13-control',
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.bot_category).toBe('suspicious');
  });

  test('case 14: bot_filter blocks custom-blocked agent', async () => {
    const host = await createIsolatedHost(api, 'bot-custom');
    await api.setBotFilter(host.id, { customBlockedAgents: 'MyEvilBot' });
    const row = await waitForReload(api, {
      host: host.domain_names[0],
      path: '/case14',
      userAgent: 'MyEvilBot/2.0',
      expectedStatus: 403,
      expectedBlockReason: 'bot_filter',
    });

    // Control request: unblocked custom agent must not return 403
    const controlRes = triggerRequest({
      host: host.domain_names[0],
      path: '/case14-control',
      userAgent: 'GoodBot/1.0',
    });
    expect(controlRes.status).not.toBe(403);

    expect(row.bot_category).toBe('custom');
  });

  // ----- Intentionally skipped -----
  // The plan calls these out as deferred. They depend on data the e2e stack
  // doesn't seed yet; activating them here would either produce false negatives
  // (no rows) or require fixture additions that are out of scope for M3.7.

  // eslint-disable-next-line playwright/no-skipped-test
  test.skip('cloud_provider_block — needs seeded cloud_providers table (CIDRs for AWS/GCP/etc.)', () => {});

  // eslint-disable-next-line playwright/no-skipped-test
  test.skip('cloud_provider_challenge — same dependency as cloud_provider_block', () => {});

  // eslint-disable-next-line playwright/no-skipped-test
  test.skip('filter_subscription_* — needs an active filter subscription seeded with UA + IP entries', () => {});

  // eslint-disable-next-line playwright/no-skipped-test
  test.skip('rate_limit — template does not emit block_reason="rate_limit" by design; status-only check is covered elsewhere', () => {});
});
