import type { Metadata } from 'next';
import { EventBrowser } from '@/components/events/EventBrowser';
import { listEvents, PAGE_SIZE, type InitialEvents } from '@/lib/events';
import { ApiClientError } from '@/lib/api';
import { listPromotions } from '@/lib/promotions';
import { runtimeFeatures } from '@/lib/runtime';
export const metadata: Metadata = { title: 'Discover events', alternates: { canonical: '/discover' } };
export default async function DiscoverPage() {
  let initial: InitialEvents;
  const [events, promotions, flags] = await Promise.allSettled([listEvents(), listPromotions(), runtimeFeatures()]);
  if (events.status === 'fulfilled') initial = { events: events.value.data, error: null, hasMore: events.value.data.length === PAGE_SIZE };
  else initial = { events: [], error: events.reason instanceof ApiClientError ? events.reason.message : 'Events could not be loaded. Please try again.', hasMore: false };
  return <EventBrowser initial={initial} promotions={promotions.status === 'fulfilled' && flags.status === 'fulfilled' && flags.value['discover.promotions'] ? promotions.value : null} />;
}
