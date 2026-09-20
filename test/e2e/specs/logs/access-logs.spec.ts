import { test, expect } from '@playwright/test';
import { ROUTES, TIMEOUTS } from '../../fixtures/test-data';

test.describe('Access Logs', () => {
  test('should navigate to access logs page', async ({ page }) => {
    await page.goto(ROUTES.logsAccess);
    await expect(page).toHaveURL(/\/logs\/access/);
  });

  test('should display access logs interface', async ({ page }) => {
    await page.goto(ROUTES.logsAccess);
    await page.waitForLoadState('domcontentloaded');

    // Should have main content area
    await expect(page.locator('main')).toBeVisible();
  });

  test('should show log entries or empty state', async ({ page }) => {
    await page.goto(ROUTES.logsAccess);
    await page.waitForLoadState('domcontentloaded');

    // Either logs are displayed or an empty state. Asserted with expect(),
    // which retries, rather than with count(), which is a snapshot: after
    // domcontentloaded React has not necessarily rendered yet, so counting
    // right there was a race that lost whenever the machine was busy — it
    // passed alone and failed inside a parallel run.
    await expect(
      page.locator('table, [class*="log"], [class*="row"]')
        .or(page.locator('text=/no.*log|empty|no.*data/i'))
        .first()
    ).toBeVisible();
  });

  test('should have log filter options', async ({ page }) => {
    await page.goto(ROUTES.logsAccess);
    await page.waitForLoadState('networkidle');

    // Wait for React to render the log toolbar (Filters button lives in LogToolbar.tsx)
    await page.locator('button').filter({ hasText: /filter/i }).first().waitFor({ state: 'visible' });

    // Should have filter/search capabilities
    const hasFilters = await page.locator('input[type="search"], select, button').filter({
      has: page.locator('text=/filter|search|date/i'),
    }).count() > 0;

    // Filter section might be collapsed or in a panel
    expect(hasFilters || (await page.locator('text=/filter/i').count()) > 0).toBeTruthy();
  });

  test('should navigate between log sub-tabs', async ({ page }) => {
    await page.goto(ROUTES.logsAccess);
    await page.waitForLoadState('domcontentloaded');

    // Click on WAF Events tab - use link/tab within the main content area
    const wafTab = page.locator('main a, main button, [role="tab"]').filter({ hasText: /waf.*event/i }).first();
    if (await wafTab.isVisible()) {
      await wafTab.click();
      await expect(page).toHaveURL(/\/logs\/waf-events/);
    }

    // Go back to access logs - navigate directly to avoid ambiguous "access" text matching sidebar
    await page.goto(ROUTES.logsAccess);
    await expect(page).toHaveURL(/\/logs\/access/);
  });
});

test.describe('WAF Event Logs', () => {
  test('should navigate to WAF events page', async ({ page }) => {
    await page.goto(ROUTES.logsWafEvents);
    await expect(page).toHaveURL(/\/logs\/waf-events/);
  });

  test('should display WAF events interface', async ({ page }) => {
    await page.goto(ROUTES.logsWafEvents);
    await page.waitForLoadState('domcontentloaded');

    await expect(page.locator('main')).toBeVisible();
  });
});

test.describe('System Logs', () => {
  test('should navigate to system logs page', async ({ page }) => {
    await page.goto(ROUTES.logsSystem);
    await expect(page).toHaveURL(/\/logs\/system/);
  });

  test('should display system logs interface', async ({ page }) => {
    await page.goto(ROUTES.logsSystem);
    await page.waitForLoadState('domcontentloaded');

    await expect(page.locator('main')).toBeVisible();
  });
});

test.describe('Audit Logs', () => {
  test('should navigate to audit logs page', async ({ page }) => {
    await page.goto(ROUTES.logsAudit);
    await expect(page).toHaveURL(/\/logs\/audit/);
  });

  test('should display audit logs interface', async ({ page }) => {
    await page.goto(ROUTES.logsAudit);
    await page.waitForLoadState('domcontentloaded');

    await expect(page.locator('main')).toBeVisible();
  });

  test('should show audit entries for recent actions', async ({ page }) => {
    // Audit logs should capture user actions
    await page.goto(ROUTES.logsAudit);
    await page.waitForLoadState('domcontentloaded');

    // Retried, not counted once after a 1s cushion. The cushion was the same
    // bet as the fixed sleeps in the block-reason spec: usually enough, and
    // nothing when the machine is loaded.
    await expect(
      page.locator('table, [class*="log"], [class*="entry"]')
        .or(page.locator('text=/no.*log|empty/i'))
        .first()
    ).toBeVisible();
  });
});

test.describe('Raw Log Files', () => {
  test('should navigate to raw log files page', async ({ page }) => {
    await page.goto(ROUTES.logsRawFiles);
    await expect(page).toHaveURL(/\/logs\/raw-files/);
  });

  test('should display raw log files interface', async ({ page }) => {
    await page.goto(ROUTES.logsRawFiles);
    await page.waitForLoadState('domcontentloaded');

    await expect(page.locator('main')).toBeVisible();
  });

  test('should list available log files', async ({ page }) => {
    await page.goto(ROUTES.logsRawFiles);
    await page.waitForLoadState('domcontentloaded');

    // Same reasoning as "should show log entries or empty state": retry the
    // assertion instead of counting once, or this races the first render.
    await expect(
      page.locator('text=/\\.log|\\.gz|access|error/i')
        .or(page.locator('text=/no.*file|empty/i'))
        .first()
    ).toBeVisible();
  });
});

test.describe('Exploit Block Logs', () => {
  test('should navigate to exploit block logs page', async ({ page }) => {
    await page.goto(ROUTES.logsExploitBlocks);
    await expect(page).toHaveURL(/\/logs\/exploit-blocks/);
  });

  test('should display exploit block logs interface', async ({ page }) => {
    await page.goto(ROUTES.logsExploitBlocks);
    await page.waitForLoadState('domcontentloaded');

    await expect(page.locator('main')).toBeVisible();
  });
});
