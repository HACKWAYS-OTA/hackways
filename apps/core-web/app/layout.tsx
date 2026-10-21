import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata, Viewport } from 'next';
import type { ReactNode } from 'react';
import Link from 'next/link';
import { Icon } from '@/components/ui/Icon';
import { SkipToContent } from '@/components/ui/SkipToContent';
import { AppNavigation } from '@/components/ui/AppNavigation';
import { site } from '@/lib/site';
import './globals.css';
import './dashboard.css';

export const metadata: Metadata = { metadataBase: new URL(site.app), title: { default: 'Discover events | Hackways', template: '%s | Hackways' }, description: 'Explore published tech events and reserve your place with Hackways.' };
export const viewport: Viewport = { width: 'device-width', initialScale: 1, viewportFit: 'cover', themeColor: '#ffffff' };

export default function RootLayout({ children }: { children: ReactNode }) {
  return <html lang="en"><head>
    <link rel="preload" href="/fonts/google-sans-flex-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />
    <link rel="preload" href="/fonts/fg-futurist-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />
  </head><body><UIProvider>
    <SkipToContent />
    <header className="app-header"><div className="app-header-inner">
      <Link className="wordmark" href="/discover" aria-label="Hackways events">hackways.</Link>
      <AppNavigation />
      <div className="header-actions"><Link className="header-create" href="/create"><Icon name="add" size={17} />Create event</Link><Link className="header-account" href="/account">Account</Link></div>
    </div></header>
    <main className="app-main" id="main-content" tabIndex={-1}>{children}</main>
    <footer className="app-footer app-container"><p>© {new Date().getFullYear()} Hackways</p><div><a href={`${site.public}/about`}>About</a><a href={`${site.public}/privacy`}>Privacy</a><a href={`${site.public}/help`}>Help & FAQs</a></div></footer>
  </UIProvider></body></html>;
}
