import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { ProgramReview } from '@/components/programs/ProgramReview';
import { adminServer } from '@/lib/admin-server';
import type { Program, AdminProgramApplication } from '@hackways/api-client';
export const metadata = { title: 'Review application | Hackways Admin' };
export default async function ReviewApplication({ params }: { params: { slug: string; id: string } }) { const data = await adminServer<{ program: Program; application: AdminProgramApplication }>('global', `/programs/${params.slug}/applications/${params.id}`); return <AdminShell active="/programs"><Link className="program-admin-back" href={`/programs/${params.slug}`}>Back to {data.program.title}</Link><ProgramReview program={data.program} application={data.application} /></AdminShell>; }
