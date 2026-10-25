import { AdminShell } from '@/components/admin/AdminShell';
import { ProgramTabs } from '@/components/programs/ProgramTabs';
import { ProgramEditor } from '@/components/programs/ProgramEditor';
import { adminServer } from '@/lib/admin-server';
import type { AdminProgram } from '@hackways/api-client';
export default async function ProgramSettings({ params }: { params: { slug: string } }) { const { data: program } = await adminServer<{ data: AdminProgram }>('global', `/programs/${params.slug}`); return <AdminShell active="/programs"><div className="admin-page-heading"><h1>{program.title}</h1></div><ProgramTabs slug={program.slug} active="settings" /><ProgramEditor initial={program} /></AdminShell>; }
