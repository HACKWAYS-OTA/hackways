import type { ReactNode } from 'react';
import { brandIcons } from '@hackways/ui/brand-metadata';
export const metadata = { icons: brandIcons('', 'admin-auth') };
export default function ConnectionLayout({ children }: { children: ReactNode }) { return <>{children}</>; }
