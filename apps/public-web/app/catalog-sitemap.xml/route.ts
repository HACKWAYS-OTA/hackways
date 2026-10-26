import { catalogSitemap } from '@/lib/catalog-sitemap';

export const dynamic = 'force-dynamic';
const escapes: Record<string, string> = { '<': '&lt;', '>': '&gt;', '&': '&amp;', '"': '&quot;', "'": '&apos;' };
const xml = (value: string) => value.replace(/[<>&"']/g, char => escapes[char]);

/** Public SEO metadata adapter; all record eligibility stays in the Go APIs. */
export async function GET() {
  try {
    const entries = await catalogSitemap();
    const body = `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">${entries.map(entry => `<url><loc>${xml(entry.url)}</loc>${entry.lastModified ? `<lastmod>${xml(entry.lastModified)}</lastmod>` : ''}</url>`).join('')}</urlset>\n`;
    return new Response(body, { headers: { 'Content-Type': 'application/xml; charset=utf-8', 'Cache-Control': 'public, max-age=0, must-revalidate' } });
  } catch (error) {
    console.error('Public catalog sitemap unavailable', error instanceof Error ? error.message : 'Unknown error');
    return new Response('Catalog sitemap temporarily unavailable.', { status: 503, headers: { 'Content-Type': 'text/plain; charset=utf-8', 'Cache-Control': 'no-store', 'Retry-After': '300' } });
  }
}
