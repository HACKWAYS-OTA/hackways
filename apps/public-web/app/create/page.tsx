import { pageSocialMetadata } from '@/lib/social-metadata';
import { redirect } from 'next/navigation';
import { site } from '@/lib/site';

export const metadata = { ...pageSocialMetadata('create'), title: 'Create an event', robots: { index: false, follow: false } };

// Creation remains owned by core-web; this keeps the public entry path clean
// without moving the host-only session cookie to the public domain.
export default function CreateEntry() {
  redirect(`${site.app}/create`);
}
