import { UIProvider } from '@hackways/ui/provider';
import '@hackways/ui/styles.css';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import './globals.css';
import '@hackways/ui/auth.css';

export const metadata: Metadata = { title: 'Hackways Admin', description: 'Hackways administration.' };

export default function RootLayout({ children }: { children: ReactNode }) {
  return <html lang="en"><body><UIProvider>{children}</UIProvider></body></html>;
}
