import { requireUser } from '@/lib/account';
import { AccountActions } from '@/components/account/AccountActions';
import { Avatar as ProfileAvatar } from '@hackways/ui/avatar';
import { ProfileForm } from '@/components/account/ProfileForm';
import Link from 'next/link';
import { Icon } from '@/components/ui/Icon';
import './profile.css';
export const metadata = { title: 'Account', robots: { index: false, follow: false } };
export default async function AccountPage() {
  const user = await requireUser('/account');
  return <div className="account-page profile-page"><div className="desk-heading"><div><h1>Your profile</h1><p>Manage your details and your Hackways account.</p></div></div><div className="profile-layout"><aside className="profile-summary"><ProfileAvatar name={user.display_name} image={user.avatar_url} large /><h2>{user.display_name}</h2><p>{user.email}</p><span className="profile-verification"><Icon name={user.email_verified ? 'check' : 'alert'} size={16} />{user.email_verified ? 'Email verified' : 'Email not verified'}</span><nav aria-label="Your events"><Link href="/my-events"><Icon name="calendar" size={18} />My events<Icon name="arrow" size={16} /></Link><Link href="/my-events?view=hosting"><Icon name="people" size={18} />Hosted events<Icon name="arrow" size={16} /></Link></nav></aside><div className="profile-content"><ProfileForm user={user} /><AccountActions verified={user.email_verified} /></div></div></div>;
}
