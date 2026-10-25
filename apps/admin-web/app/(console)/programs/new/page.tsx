import { AdminShell } from '@/components/admin/AdminShell';
import { ProgramEditor } from '@/components/programs/ProgramEditor';
export const metadata = { title: 'Create program | Hackways Admin' };
export default function NewProgram() { return <AdminShell active="/programs"><div className="admin-page-heading"><h1>Create a program</h1></div><ProgramEditor /></AdminShell>; }
