import { requireEventPermission } from '@/lib/management';
import { privateApi } from '@/lib/account';
import type { Space } from '@/types/api';
import { SpaceAssignment } from '@/components/management/SpaceAssignment';
import { EventStatusControls } from '@/components/management/EventStatusControls';
export const metadata = { title: 'Event settings' };
export default async function Settings({ params }: { params: { id: string } }) {
  const { data: event, access } = await requireEventPermission(params.id, 'lifecycle');
  const association = access.permissions.includes('association');
  const [spaces, current] = association ? await Promise.allSettled([privateApi<{ data: Space[] }>('/spaces/mine?view=owned&limit=100'), event.space_id ? privateApi<{ data: Space }>(`/spaces/${event.space_id}/manage`) : Promise.resolve(null)]) : [];
  return <div className="host-settings-layout">{association && <section><h2>Event settings</h2><div className="host-settings-space">{spaces?.status === 'fulfilled' && current?.status === 'fulfilled' ? <SpaceAssignment key={`space-${event.revision}`} event={event} spaces={spaces.value.data} current={current.value?.data || null} /> : <section><h2>Event Space</h2><p className="journey-error">Your Space settings could not be loaded. Refresh to try again.</p></section>}</div></section>}<aside className="host-settings-aside"><EventStatusControls key={event.revision} event={event} canDelete={access.permissions.includes('delete')} /></aside></div>;
}
