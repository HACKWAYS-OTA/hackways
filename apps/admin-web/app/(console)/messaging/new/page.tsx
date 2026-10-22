import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { MessageComposer } from '@/components/admin/MessageComposer';
import { Icon } from '@/components/ui/Icon';
export const metadata = { title: 'Compose message | Hackways Admin' };
export default function NewMessage() {
  return <AdminShell active="/messaging"><Link href="/messaging" className="app-button app-button-outline campaign-back"><Icon name="back" />Back</Link><div className="admin-page-heading"><h1>Compose message</h1></div><MessageComposer /></AdminShell>;
}
