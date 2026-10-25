import { readEvent } from '@/lib/event-details';
import { eventDate, formatLabels } from '@/lib/events';
import { renderSocialImage, socialImageMetadata } from '@/lib/social-image';
export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';
export function generateImageMetadata() { return socialImageMetadata('Hackways event'); }
export default async function EventImage({ params, id }: { params: { slug: string }; id: string }) {
  const event = await readEvent(params.slug);
  if (event.visibility !== 'public' || event.status !== 'published') return renderSocialImage({ title: 'Find your next event.', description: 'Discover events and communities on Hackways.', label: 'Hackways' }, id);
  return renderSocialImage({ title: event.title, description: event.short_description || event.description, detail: eventDate(event, { day: 'numeric', month: 'short', year: 'numeric', hour: 'numeric', minute: '2-digit' }), label: formatLabels[event.format] }, id);
}
