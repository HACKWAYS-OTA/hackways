import { requireUser, privateApi } from '@/lib/account';
import { ApiClientError } from '@/lib/api';
import { SpaceForm } from '@/components/spaces/SpaceForm';
import '@/components/spaces/spaces.css';
export const metadata = { title: 'Create a Space', robots: { index: false, follow: false } };
export default async function NewSpace() {
  const user = await requireUser('/spaces/new');
  let profile = true;
  try { await privateApi('/events/organizer/profile'); } catch (error) { if (error instanceof ApiClientError && error.code === 'FORBIDDEN') profile = false; else throw error; }
  return <SpaceForm hostName={user.display_name} hasProfile={profile} />;
}
