import { notFound } from 'next/navigation';
import { readSpace } from '@/lib/spaces';
import { requireUser, privateApi } from '@/lib/account';
import { SpaceForm } from '@/components/spaces/SpaceForm';
import '@/components/spaces/spaces.css';
export const metadata = { title: 'Space settings', robots: { index: false, follow: false } };
export default async function Settings({ params }: { params: { slug: string } }) {
  const user = await requireUser(`/spaces/${params.slug}/settings`), space = await readSpace(params.slug);
  const membership = (await privateApi<{ data: { owned: boolean } }>(`/spaces/${space.id}/membership`)).data;
  if (!membership.owned) notFound();
  return <SpaceForm space={space} hostName={user.display_name} hasProfile />;
}
