import 'server-only';
import { unstable_cache } from 'next/cache';
import { site } from './site';
import { seoSitemapRules } from '@hackways/ui/seo';

type Entry = { url: string; lastModified?: string };
const PAGE_SIZE = 50;
const MAX_PAGES = 20;
const MAX_RESPONSE_BYTES = 2 * 1024 * 1024;

function apiOrigin(configured: string | undefined, fallback: string): string {
  const url = new URL(configured || fallback);
  if (!['http:', 'https:'].includes(url.protocol) || url.username || url.password || url.search || url.hash) throw new Error('Invalid sitemap API origin');
  return url.href.replace(/\/$/, '');
}

async function readPage(response: Response): Promise<{ data: unknown[]; has_next?: boolean }> {
  if (!response.ok || !response.body) throw new Error(`Public catalog API returned HTTP ${response.status}`);
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let bytes = 0, text = '';
  try {
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;
      bytes += value.byteLength;
      if (bytes > MAX_RESPONSE_BYTES) throw new Error('Public catalog API page exceeds sitemap size limit');
      text += decoder.decode(value, { stream: true });
    }
    text += decoder.decode();
  } finally { await reader.cancel(); }
  const parsed: unknown = JSON.parse(text);
  if (!parsed || typeof parsed !== 'object' || !('data' in parsed) || !Array.isArray(parsed.data)) throw new Error('Invalid public catalog API page');
  const more = 'has_next' in parsed ? parsed.has_next : undefined;
  if (more !== undefined && typeof more !== 'boolean') throw new Error('Invalid public catalog pagination');
  return { data: parsed.data as unknown[], has_next: more };
}

async function readFeed(base: string, path: string, signal: AbortSignal, publicPath = path): Promise<Entry[]> {
  const entries: Entry[] = [];
  for (let page = 0; page < MAX_PAGES; page++) {
    const url = `${base}/${path}?limit=${PAGE_SIZE}&offset=${page * PAGE_SIZE}`;
    const result = await readPage(await fetch(url, { signal, cache: 'no-store', credentials: 'omit', headers: { Accept: 'application/json' } }));
    if (result.data.length > PAGE_SIZE) throw new Error('Public catalog API exceeded its page limit');
    for (const row of result.data) {
      if (!row || typeof row !== 'object' || !('slug' in row) || typeof row.slug !== 'string' || !row.slug) throw new Error('Invalid public catalog slug');
      const modified = 'updated_at' in row && typeof row.updated_at === 'string' && Number.isFinite(Date.parse(row.updated_at)) ? new Date(row.updated_at).toISOString() : undefined;
      entries.push({ url: `${site.url}/${publicPath}/${encodeURIComponent(row.slug)}`, ...(modified ? { lastModified: modified } : {}) });
    }
    if (result.has_next === false || (result.has_next === undefined && result.data.length < PAGE_SIZE)) return entries;
  }
  console.warn(`Public catalog sitemap reached its ${PAGE_SIZE * MAX_PAGES}-URL cap for ${path}`);
  return entries;
}

// Go's anonymous endpoints own publication/visibility filtering. Only public URLs
// are cached, so sitemap reads do not put database queries on every request.
const cachedCatalog = unstable_cache(async (core: string, global: string) => {
  const signal = AbortSignal.timeout(8000);
  const [feeds, rules] = await Promise.all([Promise.all([readFeed(core, 'events', signal), readFeed(core, 'spaces', signal), readFeed(core, 'cities', signal), readFeed(global, 'programs', signal), readFeed(global, 'site/notices', signal, 'notice')]), seoSitemapRules()]);
  const excluded = new Set(rules.filter(rule => rule.noindex || rule.canonical_path && rule.canonical_path !== rule.path).map(rule => rule.path));
  return Array.from(new Map(feeds.flat().filter(entry => !excluded.has(new URL(entry.url).pathname)).map(entry => [entry.url, entry])).values());
}, ['public-catalog-sitemap-v3'], { revalidate: 300 });

let pending: Promise<Entry[]> | undefined;
export async function catalogSitemap(): Promise<Entry[]> {
  const production = process.env.NODE_ENV === 'production';
  const core = apiOrigin(process.env.CORE_API_URL, production ? 'http://core-backend:8081/api/v1' : 'http://localhost:8081/api/v1');
  const global = apiOrigin(process.env.GLOBAL_API_URL, production ? 'http://global-backend:8082/api/v1' : 'http://localhost:8082/api/v1');
  if (pending) return pending;
  const request = cachedCatalog(core, global);
  pending = request;
  try { return await request; }
  finally { if (pending === request) pending = undefined; }
}
