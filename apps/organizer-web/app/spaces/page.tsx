import Image from 'next/image';
import Link from 'next/link';
import type { WorkspaceSpace } from '@/types/api';
import { Avatar } from '@hackways/ui/avatar';
import { pageOffset, privateApi, requireUser } from '@/lib/account';
import { roleLabel } from '@/lib/workspace';
import { site } from '@/lib/site';
import { Icon } from '@/components/ui/Icon';

export const metadata = { title: 'Managed Spaces' };
export default async function Spaces({ searchParams }: { searchParams: { q?: string; offset?: string } }) {
  await requireUser('/spaces');
  const q = (searchParams.q || '').trim().slice(0, 200), offset = pageOffset(searchParams.offset);
  let spaces: WorkspaceSpace[] = [], error = '';
  try { spaces = (await privateApi<{ data: WorkspaceSpace[] }>(`/workspace/spaces?limit=21&offset=${offset}&q=${encodeURIComponent(q)}`)).data; }
  catch (failure) { error = failure instanceof Error ? failure.message : 'Spaces could not be loaded.'; }
  const path = (next: number) => `/spaces?${new URLSearchParams({ q, offset: String(next) })}`;
  return <div className="workspace-page"><header className="workspace-section-heading"><h1>Spaces</h1><a className="app-button app-button-primary" href={`${site.app}/spaces/new`}><Icon name="add" size={17} />Create Space</a></header>
    <form className="host-search" method="get" role="search"><Icon name="search" size={18} /><input name="q" type="search" aria-label="Search managed Spaces" placeholder="Find a Space" maxLength={200} defaultValue={q} /><button type="submit">Search</button></form>
    {error ? <div className="desk-empty"><p className="journey-error" role="alert">{error}</p><Link className="app-button app-button-outline" href={path(offset)}>Try again</Link></div> : spaces.length ? <div className="workspace-space-list">{spaces.slice(0, 20).map(space => <Link className="workspace-space-card" key={space.id} href={`/spaces/${space.slug}`}>{space.logo_url ? <Image className="workspace-space-logo" unoptimized src={space.logo_url} alt="" width={64} height={64} /> : <Avatar name={space.name} />}<div><h2>{space.name}</h2><p>{space.description}</p><small>{roleLabel[space.access_role]}{space.status === 'archived' ? ' · Archived' : ''}</small></div><Icon name="arrow" size={20} /></Link>)}</div> : <div className="desk-empty"><h2>{q ? 'No matching Spaces.' : 'No managed Spaces yet.'}</h2>{q && <Link className="quiet-link" href="/spaces">Clear search</Link>}</div>}
    {!error && (offset > 0 || spaces.length > 20) && <nav className="journey-pagination" aria-label="Managed Space pages">{offset > 0 && <Link className="app-button app-button-outline" href={path(Math.max(0, offset - 20))}>Previous</Link>}{spaces.length > 20 && <Link className="app-button app-button-outline" href={path(offset + 20)}>Next</Link>}</nav>}
  </div>;
}
