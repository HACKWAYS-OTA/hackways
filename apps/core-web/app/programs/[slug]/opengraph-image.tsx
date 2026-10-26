import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { programRead } from '@/lib/programs';
import { renderSocialImage, socialImageMetadata } from '@/lib/social-image';
export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';
export function generateImageMetadata({ params }: { params: { slug: string } }) { return params.slug === 'ambassadors' ? [{ id: 'wide', size: { width: 1200, height: 630 }, alt: 'Hackways Ambassadors', contentType: 'image/png' }] : socialImageMetadata('Hackways program'); }
export default async function ProgramImage({ params, id }: { params: { slug: string }; id: string }) {
  const { data, accepting_applications } = await programRead(params.slug);
  if (data.slug === 'ambassadors') return new Response(new Uint8Array(await readFile(join(process.cwd(), 'public', 'programs', 'ambassadors-og.png'))), { headers: { 'Content-Type': 'image/png', 'Cache-Control': 'public, max-age=86400', 'X-Content-Type-Options': 'nosniff' } });
  return renderSocialImage({ title: data.title, description: data.summary.split(/(?<=\.)\s/)[0], label: accepting_applications ? 'Applications open' : 'Applications closed', campaign: data.slug === 'ambassadors' }, id);
}
