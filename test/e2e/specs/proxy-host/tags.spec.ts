import { test, expect } from '@playwright/test';
import { ProxyHostListPage } from '../../pages/proxy-host-list.page';
import { TestDataFactory } from '../../utils/test-data-factory';
import { APIHelper } from '../../utils/api-helper';

// Tags + auto groups. The API leg proves the tag filter is AND (a host must
// carry *every* listed tag, not any of them) and that it still holds past the
// 20-per-page boundary; the UI legs prove the group panel and the row chips
// write the filter into the URL and read it back out.
//
// Serial: every assertion here is a count — group buckets, filtered totals,
// rendered rows — so two of these tests in different workers would count each
// other's seed.
test.describe.configure({ mode: 'serial' });

const SEED_UPSTREAM = '192.0.2.9';
// A second documentation address, used only by the row-chip test so its host
// can be isolated with an upstream filter no matter what else is in the list.
const ROW_UPSTREAM = '192.0.2.33';

test.describe('Proxy host tags and groups', () => {
  let listPage: ProxyHostListPage;
  let api: APIHelper;
  const created: string[] = [];

  test.beforeEach(async ({ page, request }) => {
    // Every seeded host is a real create — config generation plus an nginx
    // reload, about a second each — and the teardown pays it again. The
    // 45s default is not enough for a seed that has to cross a page boundary.
    test.setTimeout(240_000);
    listPage = new ProxyHostListPage(page);
    api = new APIHelper(request);
    await api.login();
  });

  test.afterEach(async () => {
    // Leftover hosts push later specs' new rows off page 1 of the list, which
    // reads exactly like a regression (`Expected: > 20, Received: 20`).
    for (const id of created.splice(0)) {
      await api.deleteProxyHost(id).catch(() => undefined);
    }
    await api.cleanupTestHosts();
  });

  const seed = async (n: number, tags: string[], domainPrefix: string, upstream = SEED_UPSTREAM) => {
    for (let i = 0; i < n; i++) {
      const host = await api.createProxyHost({
        domain_names: [TestDataFactory.generateDomain(`${domainPrefix}-${i}`)],
        forward_scheme: 'http',
        forward_host: upstream,
        forward_port: 8080,
        // Omitting this creates the host *disabled* (Go zero value), which
        // would make the status buckets and the rendered rows lie.
        enabled: true,
        tags,
      });
      created.push(host.id);
    }
  };

  test('tag filter is AND and survives the page boundary', async ({ request }) => {
    await seed(22, ['e2e-media', 'e2e-shared'], 'tagm'); // > one page
    await seed(3, ['e2e-family', 'e2e-shared'], 'tagf');

    const token = await api.getToken();
    const auth = { Authorization: `Bearer ${token}` };
    // `data` is null when nothing matches, so totals are the assertion and any
    // iteration guards the array.
    const q = async (query: string) => {
      const r = await request.get(`/api/v1/proxy-hosts?${query}`, { headers: auth });
      expect(r.ok()).toBeTruthy();
      return (await r.json()) as { total: number; data: { tags: string[] }[] | null };
    };

    expect((await q('tag=e2e-media&per_page=20')).total).toBe(22);
    expect((await q('tag=e2e-shared&per_page=20')).total).toBe(25);
    // AND, not overlap: no host carries both of these, so the answer is none.
    expect((await q('tag=e2e-media&tag=e2e-family')).total).toBe(0);
    expect((await q('tag=e2e-shared&tag=e2e-family')).total).toBe(3);

    // The filter has to reach the page query too, not only the count.
    const page2 = await q('tag=e2e-media&per_page=20&page=2');
    expect(page2.data ?? []).toHaveLength(2);
    for (const h of page2.data ?? []) expect(h.tags).toContain('e2e-media');

    const groupsRes = await request.get('/api/v1/proxy-hosts/groups', { headers: auth });
    expect(groupsRes.ok()).toBeTruthy();
    const groups = (await groupsRes.json()) as { tags: { name: string; count: number }[] };
    const count = (name: string) => groups.tags.find((g) => g.name === name)?.count;
    // The panel's counts must agree with what the filter actually returns.
    expect(count('e2e-media')).toBe(22);
    expect(count('e2e-family')).toBe(3);
    expect(count('e2e-shared')).toBe(25);

    // A malformed tag is rejected, not silently matched against nothing.
    const bad = await request.get('/api/v1/proxy-hosts?tag=-bad', { headers: auth });
    expect(bad.status()).toBe(400);
  });

  test('a group chip filters the table and the URL keeps it', async ({ page }) => {
    await seed(2, ['e2e-ui-tag'], 'tagui');
    await seed(1, [], 'taguinone');
    await listPage.goto();

    // The seed guarantees at least one untagged host, so an unfiltered list is
    // always wider than the filtered one. Polled: another worker's spec may be
    // adding or removing its own hosts while this one runs.
    await expect.poll(() => listPage.tableRows.count()).toBeGreaterThan(2);

    await listPage.groupTagChip('e2e-ui-tag').click();
    await expect(page).toHaveURL(/tag=e2e-ui-tag/);
    await expect(listPage.tableRows).toHaveCount(2);
    await expect(listPage.rowTagGroups).toHaveCount(2);

    // The URL is the state, so a reload restores the same view.
    await page.reload();
    await expect(page).toHaveURL(/tag=e2e-ui-tag/);
    await expect(listPage.tableRows).toHaveCount(2);

    await listPage.groupClearButton.click();
    await expect(page).not.toHaveURL(/tag=/);
    await expect.poll(() => listPage.tableRows.count()).toBeGreaterThan(2);
  });

  test('a row chip adds that tag to the filter', async ({ page }) => {
    await seed(1, ['e2e-row-tag'], 'tagrow', ROW_UPSTREAM);

    // Arrive on an upstream filter so the seeded host is the only row whatever
    // else the suite has left in the list — and so the click has to *add* to an
    // existing filter rather than replace it.
    await listPage.gotoWithFilter(`upstream=${ROW_UPSTREAM}`);
    await expect(listPage.tableRows).toHaveCount(1);

    await listPage.rowTagChip('e2e-row-tag').click();
    await expect(page).toHaveURL(/tag=e2e-row-tag/);
    expect(page.url()).toContain(`upstream=${ROW_UPSTREAM}`);
    await expect(listPage.tableRows).toHaveCount(1);
  });
});
