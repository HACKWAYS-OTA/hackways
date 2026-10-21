import { requireUser } from '@/lib/account';
import { AccountActions } from '@/components/account/AccountActions';
export const metadata = { title: 'Account', robots: { index: false, follow: false } };
export default async function AccountPage() {
  const user = await requireUser('/account');
  return <div className="account-page"><div className="desk-heading"><div><h1>Your account</h1><p>A familiar name. One place for your plans.</p></div></div><section className="account-identity"><h2>{user.display_name}</h2><p>{user.email}</p><span className="journey-status">{user.email_verified ? 'Email verified' : 'Email not verified yet'}</span></section><AccountActions verified={user.email_verified} /></div>;
}
