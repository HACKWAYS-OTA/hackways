import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import './globals.css';
import '@hackways/ui/auth.css';
import { cookies } from 'next/headers';
import { parsePreferences } from '@/lib/admin-preferences';

export const metadata: Metadata = { title: 'Hackways Admin', description: 'Hackways administration.' };

export default function RootLayout({ children }: { children: ReactNode }) {
  const preferences = parsePreferences(cookies().get('hackways_admin_display')?.value);
  return <html lang="en" data-admin-text={preferences.text} data-admin-density={preferences.density} data-admin-contrast={preferences.contrast} data-admin-motion={preferences.motion} data-admin-scrollbars={preferences.scrollbars}><body><UIProvider reducedMotion={preferences.motion === 'reduced' ? 'always' : 'user'}>{children}</UIProvider></body></html>;
}
