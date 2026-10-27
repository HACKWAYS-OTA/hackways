import { pageSocialMetadata } from '@/lib/social-metadata';
import { brandIcons } from '@hackways/ui/brand-metadata';
import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import React from "react";
import type { Metadata } from "next";
import { Header } from "@/components/site/Header";
import { Footer } from "@/components/site/Footer";
import { SkipToContent } from "@/components/site/SkipToContent";
import { SiteMotion } from "@/components/site/SiteMotion";
import { VisitorAnalytics } from '@/components/site/VisitorAnalytics';
import { site } from "@/lib/site";
import "./globals.css";

export const metadata: Metadata = { ...pageSocialMetadata('platform'),
  icons: brandIcons(),
  metadataBase: new URL(site.url),
  title: { default: "Hackathons, Tech Meetups & Developer Events | Hackways", template: "%s | Hackways" },
  description: "Discover hackathons, developer meetups, workshops and conferences. Explore AI, software, data, cloud, cybersecurity, startups and design with Hackways.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <head>
        <link rel="preload" href="/fonts/google-sans-flex-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />
        <link rel="preload" href="/fonts/fg-futurist-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />
      </head>
      <body><UIProvider>
        <SkipToContent />
        <Header />
        <VisitorAnalytics />
        <main id="main-content" tabIndex={-1}><SiteMotion />{children}</main>
        <Footer />
      </UIProvider></body>
    </html>
  );
}
