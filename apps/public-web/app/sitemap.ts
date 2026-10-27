import type { MetadataRoute } from "next";
import { site } from "@/lib/site";
import { faqTopics } from '@/lib/help-topics';
import { seoSitemapRules, includeInSitemap } from '@hackways/ui/seo';

export const dynamic = 'force-dynamic';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const rules = await seoSitemapRules();
  // These directories are served by core-web on the same public origin.
  // Keep private application, account and registration URLs out of this list.
  return ["", "/discover", "/spaces", "/cities", "/programs", "/ambassadors", "/notices", "/mcp", "/mcp/setup", "/about", "/organizers", "/help", ...faqTopics.map(topic => `/help/${topic.id}`), "/contact", "/privacy", "/terms", "/cookies", "/acceptable-use"].map((path) => ({ url: `${site.url}${path}` })).filter(row => includeInSitemap(row.url, rules));
}
