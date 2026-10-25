import { programRead } from '@/lib/programs';
import { renderSocialImage, socialImageMetadata } from '@/lib/social-image';
export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';
export function generateImageMetadata() { return socialImageMetadata('Hackways program'); }
export default async function ProgramImage({ params, id }: { params: { slug: string }; id: string }) {
  const { data, accepting_applications } = await programRead(params.slug);
  return renderSocialImage({ title: data.title, description: data.summary.split(/(?<=\.)\s/)[0], label: accepting_applications ? 'Applications open' : 'Applications closed', campaign: data.slug === 'ambassadors' }, id);
}
