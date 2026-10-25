import { notFound, redirect } from 'next/navigation';
import { requireUser } from '@/lib/account';
import { programMine } from '@/lib/programs';
import { ProgramFrame } from '@/components/programs/ProgramFrame';
import { ProgramApplicationForm } from '@/components/programs/ProgramApplicationForm';
export const dynamic = 'force-dynamic';
export const metadata = { title: 'Apply to a program', robots: { index: false, follow: false } };
export default async function ApplyPage({ params }: { params: { slug: string; step: string } }) {
  if (!['profile', 'proposal', 'review'].includes(params.step)) notFound();
  const path = `/programs/${params.slug}/apply/${params.step}`;
  const user = await requireUser(path);
  if (!user.email_verified) redirect(`/auth/verify-email?next=${encodeURIComponent(path)}`);
  const data = await programMine(params.slug);
  if (!data.application && params.step !== 'profile') redirect(`/programs/${params.slug}/apply/profile`);
  if (data.application && data.application.status !== 'draft') redirect(`/programs/${params.slug}/application`);
  return <ProgramFrame title={data.program.title} slug={params.slug} section="Your application"><ProgramApplicationForm key={`${params.step}-${data.application?.revision || 0}`} data={data} user={user} step={params.step as 'profile' | 'proposal' | 'review'} /></ProgramFrame>;
}
