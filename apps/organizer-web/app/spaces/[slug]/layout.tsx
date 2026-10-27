import Image from 'next/image';
import { SpaceTabs } from '@/components/management/SpaceTabs';
import type { ReactNode } from 'react';
import { Avatar } from '@hackways/ui/avatar';
import { managedSpace, roleLabel } from '@/lib/workspace';
import { site } from '@/lib/site';

export default async function SpaceLayout({ params, children }: { params: { slug: string }; children: ReactNode }) {
  const { data: space, access } = await managedSpace(params.slug);
  return <div className="workspace-page"><header className="workspace-space-heading">{space.logo_url ? <Image className="workspace-space-logo" unoptimized src={space.logo_url} alt="" width={64} height={64} /> : <Avatar name={space.name} />}<div><h1>{space.name}</h1><p>{roleLabel[access.role]}{space.status === 'archived' ? ' · Archived' : ''}</p></div><a className="app-button app-button-outline app-button-sm" href={`${site.app}/spaces/${space.slug}`}>View Space</a></header>
    <SpaceTabs slug={space.slug} team={access.permissions.includes('access.manage')} />
    {children}
  </div>;
}
