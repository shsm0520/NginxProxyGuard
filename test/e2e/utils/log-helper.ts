// Helpers for the block_reason end-to-end regression spec.
//
// The block_reason pipeline spans nginx (template renders `set $block_reason_var ...`)
// → access log → log_collector parser → DB → /api/v1/logs response. This file gives
// the spec two primitives:
//
//   triggerRequest()  — fire a curl request against the e2e proxy (HTTP port),
//                       spoofing Host/XFF/UA/method so we can exercise individual
//                       security paths in isolation.
//   pollForLog()      — repeatedly poll /api/v1/logs until a row matching the
//                       supplied predicate appears, or fail with a diagnostic.
//
// We shell out to curl via execFileSync (NOT exec) — the harness security hook
// rejects exec, and execFileSync also avoids shell expansion on the spoofed
// values (Host headers and X-Forwarded-For are user-controlled in this context).

import { execFileSync } from 'child_process';
import { readFileSync, unlinkSync } from 'fs';
import type { APIHelper, LogRow } from './api-helper';
import { NGINX_HTTP_PORT } from '../fixtures/test-data';

export interface TriggerRequestOptions {
  /** Virtual host the proxy should route to. */
  host: string;
  /** Path + query string starting with "/". */
  path?: string;
  /** HTTP method. Defaults to GET. */
  method?: string;
  /** User-Agent header. Defaults to a benign curl identifier. */
  userAgent?: string;
  /** Spoofed client IP. nginx's set_real_ip_from trusts 127.0.0.1 so this becomes $remote_addr. */
  xForwardedFor?: string;
  /** Additional headers to set, in `Header: value` form. */
  extraHeaders?: string[];
  /** Override the host:port the curl connects to. Defaults to 127.0.0.1:<NGINX_HTTP_PORT>. */
  origin?: string;
  /** Timeout in seconds passed to curl --max-time. Defaults to 5. */
  timeoutSec?: number;
}

export interface TriggerRequestResult {
  status: number;
  body: string;
}

let bodyFileCounter = 0;

const DEFAULT_ORIGIN = `127.0.0.1:${NGINX_HTTP_PORT}`;
const DEFAULT_USER_AGENT = 'npg-e2e/block-reason-spec';

/**
 * Fire a single HTTP request against the e2e proxy and return the response status + body.
 *
 * The combination of curl(127.0.0.1) + X-Forwarded-For makes nginx's real_ip module
 * substitute $remote_addr — which is what the security template inspects. This is the
 * only practical way to test per-IP rules (geo, banned-ip, access-list) from the host.
 */
export function triggerRequest(opts: TriggerRequestOptions): TriggerRequestResult {
  const origin = opts.origin ?? DEFAULT_ORIGIN;
  const path = opts.path ?? '/';
  const method = opts.method ?? 'GET';
  const ua = opts.userAgent ?? DEFAULT_USER_AGENT;
  const timeoutSec = opts.timeoutSec ?? 5;

  // A per-call body file. This used to be a single fixed path, which two
  // workers firing at the same moment would overwrite for each other — the
  // status was still right, so an assertion on the BODY would fail for a
  // reason nothing in the test pointed at.
  const bodyFile = `/tmp/npg-e2e-body-${process.pid}-${bodyFileCounter++}.txt`;

  const args: string[] = [
    '-sk',
    '-o', bodyFile,
    '-w', '%{http_code}',
    '--max-time', String(timeoutSec),
    '-X', method,
    '-H', `Host: ${opts.host}`,
    '-H', `User-Agent: ${ua}`,
  ];

  if (opts.xForwardedFor) {
    args.push('-H', `X-Forwarded-For: ${opts.xForwardedFor}`);
  }
  if (opts.extraHeaders) {
    for (const h of opts.extraHeaders) {
      args.push('-H', h);
    }
  }

  // Build URL — leave path encoding alone, the caller controls it. Use http:// because
  // the test proxy's default origin is the HTTP port. Tests that need TLS use the
  // HTTPS port + ssl.
  args.push(`http://${origin}${path}`);

  let statusStr: string;
  try {
    statusStr = execFileSync('curl', args, { encoding: 'utf-8', stdio: ['pipe', 'pipe', 'pipe'] });
  } catch (err: unknown) {
    const e = err as { status?: number; stdout?: string; stderr?: string };
    // curl exits non-zero on connection failure; surface what we have for diagnostics
    throw new Error(
      `triggerRequest curl failed (exit ${e.status ?? '?'}): ${e.stderr ?? ''} ${e.stdout ?? ''}`.trim()
    );
  }

  const status = parseInt(statusStr.trim(), 10);
  let body = '';
  try {
    body = readFileSync(bodyFile, 'utf-8');
  } catch {
    body = '';
  }
  try {
    unlinkSync(bodyFile);
  } catch {
    // Best effort: a leftover file in the container's /tmp is harmless.
  }

  return { status, body };
}

/**
 * Wait until the proxy answers the way the just-applied configuration says it
 * should, then return that response.
 *
 * Replaces the fixed `waitForReload()` sleep each spec used to carry. That
 * sleep was betting on a number, and the bet was structurally unwinnable:
 * NginxReloader debounces on the TRAILING edge (constants.go,
 * NginxReloaderDebounce = 2s) and every new request RESTARTS the timer. So
 * when specs run in parallel — each creating hosts and toggling security
 * settings — the reload that makes this test's rule live can be pushed back
 * indefinitely by other workers. An 800ms sleep was short even for the quiet
 * case. That is #294: block-reason case 10 failed alongside its siblings and
 * passed alone, which is the signature of a timing assumption, not of a broken
 * uri_block rule.
 *
 * Polling the real condition removes the guess in both directions: it returns
 * as soon as the rule is live (usually first try) and it still fails, with a
 * useful message, when the rule never becomes live.
 *
 * The probe is the test's own request, fired repeatedly. Every caller here
 * targets a freshly created single-purpose host on a unique path, so the extra
 * attempts only add rows that pollForLog would match anyway. Do NOT use this
 * for a rule whose own trigger is cumulative — rate limits, fail2ban counters,
 * auto-ban thresholds — because there the repeats ARE the state under test.
 */
export async function triggerUntil(
  opts: TriggerRequestOptions,
  expected: (result: TriggerRequestResult) => boolean,
  options: { timeoutMs?: number; intervalMs?: number; describe?: string } = {},
): Promise<TriggerRequestResult> {
  const timeoutMs = options.timeoutMs ?? 20000;
  const interval = options.intervalMs ?? 250;
  const deadline = Date.now() + timeoutMs;

  let attempts = 0;
  let last: TriggerRequestResult | undefined;
  let lastError: unknown;

  for (;;) {
    attempts++;
    try {
      last = triggerRequest(opts);
      lastError = undefined;
      if (expected(last)) {
        return last;
      }
    } catch (err) {
      // curl can fail outright while nginx is mid-reload; that is a state to
      // wait through, not to fail on — unless it is still happening at the
      // deadline, in which case it is reported.
      lastError = err;
    }

    if (Date.now() >= deadline) {
      const what = options.describe ?? `${opts.method ?? 'GET'} ${opts.host}${opts.path ?? '/'}`;
      const seen = lastError
        ? `the request kept failing: ${lastError instanceof Error ? lastError.message : String(lastError)}`
        : `the last response was status ${last?.status}`;
      throw new Error(
        `Timed out after ${timeoutMs}ms (${attempts} attempts) waiting for ${what} to match the expected response; ${seen}. ` +
        `If the rule is correct, the reload never landed — NginxReloader debounces on the trailing edge and restarts its timer on every request.`,
      );
    }

    await new Promise(res => setTimeout(res, interval));
  }
}

export interface LogMatchCriteria {
  /**
   * Filter by domain name (preferred — the log_collector caches host→ID lookups
   * for 60s, so a freshly-created host's rows may have NULL proxy_host_id.
   * The `host` field is always populated from the nginx access log line).
   */
  host?: string;
  /** Filter by proxy_host_id. Only reliable after the domain→ID cache refreshes. */
  hostId?: string;
  expectedBlockReason?: string;
  expectedStatus?: number;
  expectedLogType?: string;
  /** If provided, the row's request_uri must contain this substring. */
  uriContains?: string;
  /** Timeout in ms. Defaults to 12000. */
  timeoutMs?: number;
  /** Poll interval in ms. Defaults to 300. */
  intervalMs?: number;
}

export type { LogRow };

/**
 * Repeatedly query /api/v1/logs for the host until a row matching the predicate appears.
 *
 * The log_collector runs out-of-band — log lines may take 1-5s to land in DB after the
 * request was served. We poll instead of waiting a fixed timeout to keep tests snappy
 * on fast machines.
 */
export async function pollForLog(api: APIHelper, criteria: LogMatchCriteria): Promise<LogRow> {
  const deadline = Date.now() + (criteria.timeoutMs ?? 12000);
  const interval = criteria.intervalMs ?? 300;

  let lastSnapshot: LogRow[] = [];

  while (Date.now() < deadline) {
    let rows: LogRow[] = [];
    try {
      rows = await api.getLogs({
        host_id: criteria.hostId,
        host: criteria.host,
        limit: 50,
      });
    } catch {
      // Transient API hiccup — retry until deadline.
    }
    lastSnapshot = rows;

    for (const row of rows) {
      if (criteria.expectedBlockReason !== undefined && row.block_reason !== criteria.expectedBlockReason) {
        continue;
      }
      if (criteria.expectedStatus !== undefined && row.status_code !== criteria.expectedStatus) {
        continue;
      }
      if (criteria.expectedLogType !== undefined && row.log_type !== criteria.expectedLogType) {
        continue;
      }
      if (criteria.uriContains !== undefined) {
        const uri = row.request_uri ?? '';
        if (!uri.includes(criteria.uriContains)) {
          continue;
        }
      }
      return row;
    }

    await new Promise(res => setTimeout(res, interval));
  }

  // Build a diagnostic message that surfaces what we *did* see — the most common failure
  // mode is "block_reason is 'none' / wrong value" which is the regression we're guarding.
  const sample = lastSnapshot.slice(0, 5).map(r =>
    `  status=${r.status_code} block_reason=${r.block_reason ?? '∅'} uri=${r.request_uri} ua=${r.http_user_agent}`
  ).join('\n');
  throw new Error(
    `pollForLog timed out after ${criteria.timeoutMs ?? 12000}ms.\n` +
    `Expected: ${JSON.stringify({
      block_reason: criteria.expectedBlockReason,
      status: criteria.expectedStatus,
      log_type: criteria.expectedLogType,
      uri_contains: criteria.uriContains,
    })}\n` +
    `Most recent rows for host=${criteria.host ?? '∅'} hostId=${criteria.hostId ?? '∅'}:\n${sample || '  (none)'}`
  );
}
