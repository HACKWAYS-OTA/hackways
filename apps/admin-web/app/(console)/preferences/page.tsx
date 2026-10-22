import { cookies } from 'next/headers';
import { AdminShell } from '@/components/admin/AdminShell';
import { AdminPreferences } from '@/components/admin/AdminPreferences';
import { parsePreferences } from '@/lib/admin-preferences';
export const metadata = { title: 'Display & accessibility | Hackways Admin' };
export default function Preferences() { return <AdminShell active="/preferences"><div className="admin-page-heading"><h1>Display & accessibility</h1></div><AdminPreferences initial={parsePreferences(cookies().get('hackways_admin_display')?.value)} /></AdminShell>; }
