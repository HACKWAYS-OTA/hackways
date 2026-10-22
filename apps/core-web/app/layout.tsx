import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata, Viewport } from 'next';
import type { ReactNode } from 'react';
import { AppChrome } from '@/components/ui/AppChrome';
import { site } from '@/lib/site';
import { currentUser } from '@/lib/account';
import './globals.css';
import './dashboard.css';
import '@hackways/ui/auth.css';

export const metadata: Metadata = { metadataBase: new URL(site.app), title: { default: 'Discover events | Hackways', template: '%s | Hackways' }, description: 'Explore published tech events and reserve your place with Hackways.' };
export const viewport: Viewport = { width: 'device-width', initialScale: 1, viewportFit: 'cover', themeColor: '#ffffff' };

export default async function RootLayout({ children }: { children: ReactNode }) {
  const user = await currentUser();
  return <html lang="en"><head>
    <link rel="preload" href="/fonts/google-sans-flex-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />
    <link rel="preload" href="/fonts/fg-futurist-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />
  </head><body><UIProvider><AppChrome user={user}>{children}</AppChrome></UIProvider></body></html>;
}
