'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

export function SpaceTabs({ slug, team }: { slug: string; team: boolean }) {
  const pathname = usePathname(), base = `/spaces/${slug}`;
  return <nav className="workspace-resource-nav" aria-label="Manage this Space"><Link href={base} aria-current={pathname === base ? 'page' : undefined}>Events</Link>{team && <Link href={`${base}/team`} aria-current={pathname === `${base}/team` ? 'page' : undefined}>Team access</Link>}</nav>;
}
