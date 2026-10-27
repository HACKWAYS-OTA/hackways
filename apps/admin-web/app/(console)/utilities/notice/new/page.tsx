import { AdminShell } from '@/components/admin/AdminShell';
import { NoticeEditor } from '@/components/admin/NoticeEditor';
import { UtilityNavigation } from '@/components/admin/UtilityNavigation';
import '../../utilities.css';
export default function CreateNotice() { return <AdminShell active="/utilities/notice"><UtilityNavigation active="notice" /><header className="utility-heading"><h1>Create notice</h1></header><NoticeEditor /></AdminShell>; }
