import { requireUser, privateApi } from '@/lib/account';
import { ApiClientError } from '@/lib/api';
import { CreateEventForm } from '@/components/hosting/CreateEventForm';
import type { Space } from '@/types/api';
import Link from 'next/link';
export const metadata = { title: 'Create an event', robots: { index: false, follow: false } };
export default async function CreatePage({ searchParams }: { searchParams: { space?: string } }) {
  const spaceSlug = typeof searchParams.space === 'string' ? searchParams.space : '';
  const returnPath = spaceSlug ? `/create?space=${encodeURIComponent(spaceSlug)}` : '/create';
  const user = await requireUser(returnPath);
  let organizer: { display_name: string } | null = null;
  const [profile, list] = await Promise.allSettled([privateApi<{ data: { display_name: string } }>('/events/organizer/profile'), privateApi<{ data: Space[] }>('/spaces/mine?view=owned&limit=100')]);
  if (profile.status === 'fulfilled') organizer = profile.value.data;
  else if (!(profile.reason instanceof ApiClientError && profile.reason.code === 'FORBIDDEN')) return <Unavailable path={returnPath} message={profile.reason instanceof Error ? profile.reason.message : 'The host profile service is unavailable.'} />;
  if (list.status === 'rejected' || !Array.isArray(list.value.data)) return <Unavailable path={returnPath} message="Your Spaces could not be loaded. Please retry when the event service is available." />;
  const spaces = list.value.data.filter(space => space.status === 'active');
  let initialSpace = spaces.find(space => space.slug === spaceSlug)?.slug || '';
  if (spaceSlug && !initialSpace) {
    if (!/^[a-z0-9-]{1,230}$/.test(spaceSlug)) return <Unavailable path="/spaces?view=owned" message="Choose an active Space you own, or start a personal event." />;
    try {
      const candidate = (await privateApi<{ data: Space }>(`/spaces/by-slug/${encodeURIComponent(spaceSlug)}`)).data;
      const membership = (await privateApi<{ data: { owned: boolean } }>(`/spaces/${candidate.id}/membership`)).data;
      if (!membership.owned || candidate.status !== 'active') return <Unavailable path="/spaces?view=owned" message="This Space is archived or belongs to another host. Choose an active Space you own." />;
      spaces.push(candidate); initialSpace = candidate.slug;
    } catch (error) { return <Unavailable path="/spaces?view=owned" message={error instanceof ApiClientError && error.code === 'NOT_FOUND' ? 'This Space is no longer available. Choose another Space or start a personal event.' : 'The Space could not be loaded. Please retry.'} />; }
  }
  return <div className="compose-page"><CreateEventForm organizer={organizer} name={user.display_name} spaces={spaces} initialSpace={initialSpace} /></div>;
}
function Unavailable({ path, message }: { path: string; message: string }) {
  return <section className="desk-page desk-empty"><h1>Let’s get your event started.</h1><p role="alert">{message}</p><div className="journey-actions"><Link href={path} className="app-button app-button-primary">Try again</Link><Link href="/create" className="quiet-link">Create a personal event</Link></div></section>;
}
