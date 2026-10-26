import { AdminShell } from '@/components/admin/AdminShell';
import { ProgramTabs } from '@/components/programs/ProgramTabs';
import { ProgramCommunication } from '@/components/programs/ProgramCommunication';
import { adminServer } from '@/lib/admin-server';
import type { AdminProgram } from '@hackways/api-client';
export const metadata = { title: 'Applicant communications | Hackways Admin' };
export default async function Communications({ params, searchParams }: { params: { slug: string }; searchParams: { ids?: string; scope?: string; q?: string; decision?: string } }) {
  const { data: program } = await adminServer<{ data: AdminProgram }>('global', `/programs/${encodeURIComponent(params.slug)}`);
  const ids = (searchParams.ids || '').split(',').filter(Boolean).slice(0, 200);
  return <AdminShell active="/programs"><div className="admin-page-heading"><div><h1>{program.title}</h1><p>Applicant communications</p></div></div><ProgramTabs slug={program.slug} active="communications" /><ProgramCommunication slug={program.slug} ids={ids} scope={searchParams.scope || 'all'} search={(searchParams.q || '').slice(0, 200)} decision={searchParams.decision || ''} /></AdminShell>;
}
