'use client';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';

export function ManagementChrome({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  return pathname.endsWith('/edit') ? null : <>{children}</>;
}
