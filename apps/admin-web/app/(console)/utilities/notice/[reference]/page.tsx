import { NoticeEditor } from '@/components/admin/NoticeEditor';
import { adminServer } from '@/lib/admin-server';
import type { Notice } from '@/lib/utilities';
import '../../utilities.css';
export default async function EditNotice({ params }: { params: { reference: string } }) {
 const result = await adminServer<{ data: Notice }>('global', `/utilities/notices/${encodeURIComponent(params.reference)}`);
 return <NoticeEditor initial={result.data} />;
}
