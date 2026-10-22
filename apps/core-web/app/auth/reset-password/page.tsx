import { PasswordReset } from '@/components/account/AuthRecovery';
export const metadata = { title: 'Reset password', robots: { index: false, follow: false }, referrer: 'no-referrer' };
export default function ResetPassword() { return <PasswordReset />; }
