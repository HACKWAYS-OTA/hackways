import type { MetadataRoute } from "next";
import { site } from "@/lib/site";
import { faqTopics } from '@/lib/help-topics';

export default function sitemap(): MetadataRoute.Sitemap {
  return ["", "/mcp", "/mcp/setup", "/about", "/organizers", "/help", ...faqTopics.map(topic => `/help/${topic.id}`), "/contact", "/privacy", "/terms", "/cookies", "/acceptable-use"].map((path) => ({ url: `${site.url}${path}` }));
}
