import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata, Viewport } from 'next';
import type { ReactNode } from 'react';
import Link from 'next/link';
import { site } from '@/lib/site';
import { SkipToContent } from '@/components/ui/SkipToContent';
import './globals.css';
import './dashboard.css';
import './host-dashboard.css';
export const metadata: Metadata = { metadataBase: new URL(site.organizer), title: { default: 'Manage events | Hackways', template: '%s | Hackways' }, description: 'Welcome your guests and manage the details of your events.', robots: { index: false, follow: false } };
export const viewport: Viewport = { width: 'device-width', initialScale: 1, themeColor: '#ffffff' };
export default function Layout({ children }: { children: ReactNode }) {
  return <html lang="en"><head><link rel="preload" href="/fonts/google-sans-flex-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" /><link rel="preload" href="/fonts/fg-futurist-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" /></head><body><UIProvider><SkipToContent /><header className="app-header"><div className="app-header-inner"><a className="wordmark" href={`${site.app}/my-events`}>hackways.</a><nav className="app-nav" aria-label="Event management"><Link href="/">Manage events</Link><a href={`${site.app}/my-events`}>My events</a></nav><div className="header-actions"><a className="header-create" href={`${site.app}/create`}>Create event</a><a className="header-account" href={`${site.app}/account`}>Account</a></div></div></header><main className="app-main" id="main-content" tabIndex={-1}>{children}</main><footer className="app-footer app-container"><p>Hackways</p><a href={`${site.app}/discover`}>Discover something new</a></footer></UIProvider></body></html>;
}
