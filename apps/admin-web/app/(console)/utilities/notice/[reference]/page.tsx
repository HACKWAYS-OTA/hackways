import { AdminShell } from '@/components/admin/AdminShell';
import { NoticeEditor } from '@/components/admin/NoticeEditor';
import { UtilityNavigation } from '@/components/admin/UtilityNavigation';
import { adminServer } from '@/lib/admin-server';
import type { Notice } from '@/lib/utilities';
import '../../utilities.css';
export default async function EditNotice({ params }: { params: { reference: string } }) {
 const result = await adminServer<{ data: Notice }>('global', `/utilities/notices/${encodeURIComponent(params.reference)}`);
 return <AdminShell active="/utilities/notice"><UtilityNavigation active="notice" /><header className="utility-heading"><h1>Edit notice</h1></header><NoticeEditor initial={result.data} /></AdminShell>;
}
