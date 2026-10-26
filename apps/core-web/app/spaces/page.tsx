import Image from 'next/image';
import { pageSocialMetadata } from '@/lib/social-metadata';
import type { Metadata } from 'next';
import Link from 'next/link';
import type { Space } from '@/types/api';
import { fetchApi } from '@/lib/api';
import { requireUser, privateApi, pageOffset } from '@/lib/account';
import { Icon } from '@/components/ui/Icon';
import '@/components/spaces/spaces.css';
export function generateMetadata({ searchParams }: { searchParams: { view?: string; q?: string; offset?: string } }): Metadata {
  const privateView = ['owned', 'joined'].includes(searchParams.view || '');
  const filtered = Boolean((searchParams.q || '').slice(0, 200));
  const offset = pageOffset(searchParams.offset);
  return {
    ...pageSocialMetadata('spaces'),
    title: 'Spaces',
    description: 'Find a community around what you care about. Join a Space and discover its events.',
    ...(privateView || filtered
      ? { robots: { index: false, follow: !privateView } }
      : { alternates: { canonical: offset ? `/spaces?offset=${offset}` : '/spaces' } }),
  };
}
export default async function Spaces({ searchParams }: { searchParams: { view?: string; q?: string; offset?: string } }) {
  const view = ['owned', 'joined'].includes(searchParams.view || '') ? searchParams.view! : 'all';
  const offset = pageOffset(searchParams.offset), query = (searchParams.q || '').slice(0, 200);
  if (view !== 'all') await requireUser(`/spaces?view=${view}`);
  let spaces: Space[] = [], error = '';
  try { spaces = (await (view === 'all' ? fetchApi<{ data: Space[] }>(`/spaces?limit=20&offset=${offset}&q=${encodeURIComponent(query)}`, { cache: 'no-store' }) : privateApi<{ data: Space[] }>(`/spaces/mine?view=${view}&limit=20&offset=${offset}`))).data; }
  catch (err) { error = err instanceof Error ? err.message : 'Spaces could not be loaded.'; }
  const base = `/spaces?view=${view}&q=${encodeURIComponent(query)}`;
  return <div className="spaces-page"><header className="desk-heading"><div><h1>Find your people.</h1><p>A Space for shared interests. Events that bring you together.</p></div><Link className="app-button app-button-primary" href="/spaces/new"><Icon name="add" size={16} />Create a Space</Link></header><nav className="desk-tabs" aria-label="Browse Spaces">{[['all', 'Explore Spaces'], ['joined', 'My communities'], ['owned', 'Spaces I host']].map(([key, title]) => <Link key={key} href={`/spaces?view=${key}`} aria-current={view === key ? 'page' : undefined}>{title}</Link>)}</nav>{view === 'all' && <form className="spaces-search" action="/spaces"><Icon name="search" size={18} /><label htmlFor="space-search" className="visually-hidden">Find a Space</label><input id="space-search" type="search" name="q" defaultValue={query} maxLength={200} /><button className="quiet-link">Search</button></form>}{error ? <section className="desk-empty"><h2>We couldn’t load Spaces.</h2><p role="alert">{error}</p><Link className="app-button app-button-outline" href={base}>Try again</Link></section> : spaces.length === 0 ? <section className="desk-empty"><Icon name="people" size={30} /><h2>{query ? 'No matching Spaces.' : view === 'owned' ? 'Bring your community together.' : view === 'joined' ? 'Your people are out there.' : 'Make room for a new community.'}</h2><p>{view === 'owned' ? 'Create a Space, tell people what it’s about, and host your first event under it.' : view === 'joined' ? 'Join a Space that shares your interests. Its upcoming events are always a visit away.' : 'Explore another interest, or start a Space for the people you want to bring together.'}</p><Link className="app-button app-button-outline" href={view === 'joined' ? '/spaces' : '/spaces/new'}>{view === 'joined' ? 'Explore Spaces' : 'Create a Space'}</Link></section> : <div className="space-directory">{spaces.map(space => <Link className="space-directory-row" href={`/spaces/${space.slug}`} key={space.slug}><span className="space-mark" aria-hidden="true">{space.logo_url ? <Image unoptimized src={space.logo_url} alt="" width={64} height={64} /> : space.name.trim().slice(0, 1)}</span><div><span className="space-member-count">{space.member_count} {space.member_count === 1 ? 'member' : 'members'}{space.status === 'archived' ? ' · Archived' : ''}</span><h2>{space.name}</h2><p>{space.description}</p></div><span className="space-open" aria-hidden="true">Visit Space<Icon name="arrow" size={17} /></span></Link>)}</div>}{!error && (offset > 0 || spaces.length === 20) && <nav className="journey-pagination" aria-label="More Spaces">{offset > 0 && <Link className="app-button app-button-outline" href={`${base}&offset=${Math.max(0, offset - 20)}`}>Previous</Link>}{spaces.length === 20 && <Link className="app-button app-button-outline" href={`${base}&offset=${offset + 20}`}>Next</Link>}</nav>}</div>;
}
