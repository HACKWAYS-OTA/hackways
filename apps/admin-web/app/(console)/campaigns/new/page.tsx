import Link from 'next/link';
import { AdminShell } from '@/components/admin/AdminShell';
import { CampaignEditor } from '@/components/admin/CampaignEditor';
import { Icon } from '@/components/ui/Icon';
export const metadata = { title: 'New campaign | Hackways Admin' };
export default function NewCampaign() { return <AdminShell active="/campaigns"><Link className="app-button app-button-outline campaign-back" href="/campaigns"><Icon name="back" />Back</Link><div className="admin-page-heading"><h1>New campaign</h1></div><div className="campaign-workbench"><section><p className="campaign-from">From updates@mail.hackways.in</p><CampaignEditor /></section><aside className="campaign-summary"><h2>Draft first</h2><p>Save your message to see the real email preview, review eligible contacts, and send a test to your verified admin email.</p><p>Campaigns include an unsubscribe link and use the existing India email service.</p></aside></div></AdminShell>; }
