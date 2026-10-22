"use client";
import { useRouter } from 'next/navigation';
import { AuthLayout } from '@hackways/ui/auth';
import { AuthRecoveryForm, AuthResetPasswordForm, AuthPromotionUnsubscribe, AuthVerificationForm, type AuthFlowRequest } from '@hackways/ui/auth-recovery';
import { fetchApi } from '@/lib/api';
import { site } from '@/lib/site';
const request: AuthFlowRequest = (path, body) => fetchApi(path, { method: 'POST', body: JSON.stringify(body) });
export function PasswordRecovery({ next }: { next: string }) {
  return <AuthRecoveryForm request={request} signInHref={`/login?next=${encodeURIComponent(next)}`} homeHref={`${site.app}/discover`} />;
}
export function EmailVerification({ email, next, queued, unavailable }: { email: string; next: string; queued: boolean; unavailable: boolean }) {
 const router=useRouter();
 return <AuthLayout title="Verify your email" homeHref={`${site.app}/discover`}>
  {unavailable && <p className="hw-auth-note" role="status">Email delivery is currently unavailable. Request another code when delivery is restored.</p>}
  <AuthVerificationForm email={email} request={request} initiallyQueued={queued} onVerified={()=>{router.replace(next);router.refresh();}} onSignOut={()=>{router.replace('/login');router.refresh();}} />
 </AuthLayout>;
}

export function PasswordReset() {
 return <AuthResetPasswordForm request={request} signInHref="/login" recoveryHref="/auth/forgot-password" homeHref={`${site.app}/discover`} />;
}

export function PromotionUnsubscribe() { return <AuthPromotionUnsubscribe request={request} homeHref={`${site.app}/discover`} />; }
