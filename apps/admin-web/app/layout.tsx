import { staticSocialMetadata } from '@hackways/ui/social-image';
import { brandIcons } from '@hackways/ui/brand-metadata';
import '@hackways/ui/mcp.css';
import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import './globals.css';
import '@hackways/ui/auth.css';

const publicOrigin = process.env.NEXT_PUBLIC_PUBLIC_URL || 'https://hackways.in';
export const metadata: Metadata = { ...staticSocialMetadata(publicOrigin, 'admin', 'Hackways Admin', 'Hackways administration.'), icons: brandIcons(publicOrigin, 'admin-auth'), title: 'Hackways Admin', description: 'Hackways administration.', robots: { index: false, follow: false } };

export default function RootLayout({ children }: { children: ReactNode }) {
  return <html lang="en"><body><UIProvider>{children}</UIProvider></body></html>;
}
