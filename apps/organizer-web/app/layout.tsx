import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata, Viewport } from 'next';
import type { ReactNode } from 'react';
import { AppChrome } from '@/components/ui/AppChrome';
import { site } from '@/lib/site';
import './globals.css';
import './dashboard.css';
import './host-dashboard.css';

export const metadata: Metadata = { metadataBase: new URL(site.organizer), title: { default: 'Manage events | Hackways', template: '%s | Hackways' }, description: 'Welcome your guests and manage the details of your events.', robots: { index: false, follow: false } };
export const viewport: Viewport = { width: 'device-width', initialScale: 1, themeColor: '#ffffff' };

export default function Layout({ children }: { children: ReactNode }) {
  return <html lang="en"><head><link rel="preload" href="/fonts/google-sans-flex-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" /><link rel="preload" href="/fonts/fg-futurist-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" /></head><body><UIProvider><AppChrome>{children}</AppChrome></UIProvider></body></html>;
}
