import type { MetadataRoute } from "next";
import { site } from "@/lib/site";
import { faqTopics } from '@/lib/help-topics';

export default function sitemap(): MetadataRoute.Sitemap {
  // These directories are served by core-web on the same public origin.
  // Keep private application, account and registration URLs out of this list.
  return ["", "/discover", "/spaces", "/cities", "/programs", "/ambassadors", "/mcp", "/mcp/setup", "/about", "/organizers", "/help", ...faqTopics.map(topic => `/help/${topic.id}`), "/contact", "/privacy", "/terms", "/cookies", "/acceptable-use"].map((path) => ({ url: `${site.url}${path}` }));
}
