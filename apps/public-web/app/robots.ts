import type { MetadataRoute } from "next";
import { site } from "@/lib/site";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: process.env.NODE_ENV === "production"
      ? { userAgent: "*", allow: "/" }
      : { userAgent: "*", disallow: "/" },
    sitemap: [`${site.url}/sitemap.xml`, `${site.url}/catalog-sitemap.xml`],
  };
}
