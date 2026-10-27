import 'server-only';
import type { Metadata } from 'next';
import { staticSocialMetadata } from './SocialImage';

type SeoPage = { path: string; title: string; description: string; canonical_path: string; noindex: boolean };
export type SeoRule = Pick<SeoPage, 'path' | 'canonical_path' | 'noindex'>;
export function siteContentApi() {
  return (process.env.GLOBAL_API_URL || (process.env.NODE_ENV === 'production' ? 'http://global-backend:8082/api/v1' : 'http://localhost:8082/api/v1')).replace(/\/$/, '') + '/site';
}

/** Presentation metadata only. Go owns public-route eligibility and saved settings. */
export async function managedMetadata(path: string, base: Metadata): Promise<Metadata> {
  if (base.robots && typeof base.robots === 'object' && base.robots.index === false) return base;
  base = { ...base, robots: base.robots || { index: true, follow: true } };
  try {
    const response = await fetch(`${siteContentApi()}/seo?${new URLSearchParams({ path })}`, { next: { revalidate: 30 }, credentials: 'omit', signal: AbortSignal.timeout(3000) });
    if (response.status === 400) return { ...base, robots: { index: false, follow: false } };
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const { data } = await response.json() as { data: SeoPage | null };
    if (!data) return base;
    const social = staticSocialMetadata(process.env.NEXT_PUBLIC_SITE_URL || process.env.NEXT_PUBLIC_APP_URL || 'https://hackways.in', 'platform', typeof base.title === 'string' ? base.title : 'Hackways', base.description || '');
    return { ...base,
      ...(data.title ? { title: { absolute: data.title } } : {}),
      ...(data.description ? { description: data.description } : {}),
      alternates: { ...base.alternates, canonical: data.canonical_path || base.alternates?.canonical || path },
      ...(data.noindex ? { robots: { index: false, follow: true } } : {}),
      openGraph: { ...social.openGraph, ...base.openGraph, ...(data.title ? { title: data.title } : {}), ...(data.description ? { description: data.description } : {}), ...(data.canonical_path ? { url: data.canonical_path } : {}) },
      twitter: { ...social.twitter, ...base.twitter, ...(data.title ? { title: data.title } : {}), ...(data.description ? { description: data.description } : {}) },
    };
  } catch { console.warn('SEO settings unavailable; existing page metadata retained:', path); return base; }
}

export async function seoSitemapRules(): Promise<SeoRule[]> {
  const rules: SeoRule[] = [];
  const signal = AbortSignal.timeout(6000);
  for (let page = 0; page < 20; page++) {
    const response = await fetch(`${siteContentApi()}/seo-rules?limit=100&offset=${page * 100}`, { next: { revalidate: 30 }, credentials: 'omit', signal });
    if (!response.ok) throw new Error('SEO sitemap rules unavailable');
    const result = await response.json() as { data: SeoRule[]; has_next: boolean };
    rules.push(...result.data);
    if (!result.has_next) return rules;
  }
  throw new Error('SEO sitemap rules exceed the bounded page limit');
}

export function includeInSitemap(url: string, rules: SeoRule[]) {
  const path = new URL(url).pathname;
  const rule = rules.find(row => row.path === path);
  return !rule?.noindex && (!rule?.canonical_path || rule.canonical_path === path);
}
