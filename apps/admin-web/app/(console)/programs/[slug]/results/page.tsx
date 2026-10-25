import { AdminShell } from '@/components/admin/AdminShell';
import { ProgramTabs } from '@/components/programs/ProgramTabs';
import { ProgramRelease } from '@/components/programs/ProgramRelease';
import { adminServer } from '@/lib/admin-server';
import type { AdminProgram } from '@hackways/api-client';
export default async function ProgramResults({ params }: { params: { slug: string } }) { const { data: program } = await adminServer<{ data: AdminProgram }>('global', `/programs/${params.slug}`); return <AdminShell active="/programs"><div className="admin-page-heading"><h1>{program.title}</h1></div><ProgramTabs slug={program.slug} active="results" /><ProgramRelease key={program.review_revision} program={program} /></AdminShell>; }
