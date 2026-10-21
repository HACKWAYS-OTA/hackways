import type { Metadata } from 'next';
import { EventBrowser } from '@/components/events/EventBrowser';
import { listEvents, PAGE_SIZE, type InitialEvents } from '@/lib/events';
import { ApiClientError } from '@/lib/api';
export const metadata: Metadata = { title: 'Discover events', alternates: { canonical: '/discover' } };
export default async function DiscoverPage() {
  let initial: InitialEvents;
  try { const page = await listEvents(); initial = { events: page.data, error: null, hasMore: page.data.length === PAGE_SIZE }; }
  catch (error) { initial = { events: [], error: error instanceof ApiClientError ? error.message : 'Events could not be loaded. Please try again.', hasMore: false }; }
  return <EventBrowser initial={initial} />;
}
