import { readSpace } from '@/lib/spaces';
import { renderSocialImage, socialImageMetadata } from '@/lib/social-image';
export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';
export function generateImageMetadata() { return socialImageMetadata('Hackways Space'); }
export default async function SpaceImage({ params, id }: { params: { slug: string }; id: string }) {
  const space = await readSpace(params.slug);
  return renderSocialImage({ title: space.name, description: space.description, label: space.status === 'archived' ? 'Archived Space' : 'A community on Hackways' }, id);
}
