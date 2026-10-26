import { staticSocialMetadata } from '@hackways/ui/social-image';
import { brandIcons } from '@hackways/ui/brand-metadata';
import { SessionRefresh } from '@hackways/ui/session-refresh';
import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata, Viewport } from 'next';
import type { ReactNode } from 'react';
import { AppChrome } from '@/components/ui/AppChrome';
import { site } from '@/lib/site';
import { currentUser } from '@/lib/account';
import './globals.css';
import './dashboard.css';
import './host-dashboard.css';
import '@hackways/ui/auth.css';

export const metadata: Metadata = { ...staticSocialMetadata(site.app, 'manage', 'Manage events | Hackways', 'Welcome your guests and manage the details of your events.'), icons: brandIcons('/manage'), metadataBase: new URL(site.organizer), title: { default: 'Manage events | Hackways', template: '%s | Hackways' }, description: 'Welcome your guests and manage the details of your events.', robots: { index: false, follow: false } };
export const viewport: Viewport = { width: 'device-width', initialScale: 1, themeColor: '#ffffff' };

export default async function Layout({ children }: { children: ReactNode }) {
  const user = await currentUser();
  return <html lang="en"><head><link rel="preload" href="/manage/fonts/google-sans-flex-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" /><link rel="preload" href="/manage/fonts/fg-futurist-regular.woff2" as="font" type="font/woff2" crossOrigin="anonymous" /></head><body><UIProvider><SessionRefresh enabled={Boolean(user)} apiBase={process.env.NEXT_PUBLIC_GLOBAL_API_URL || (process.env.NODE_ENV === 'production' ? '/api/v1' : 'http://localhost:8082/api/v1')} /><AppChrome user={user}>{children}</AppChrome></UIProvider></body></html>;
}
