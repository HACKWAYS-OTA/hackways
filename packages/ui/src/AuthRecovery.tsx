'use client';

import { useEffect, useRef, useState, type FormEvent } from 'react';
import { AuthLayout, AuthPasswordInput, AuthSubmit } from './Auth';

export type AuthFlowRequest = (path: string, body: Record<string, string>) => Promise<unknown>;

function useCooldown(initial = 0) {
  const [seconds, setSeconds] = useState(initial);
  useEffect(() => { if (!seconds) return; const timer = window.setTimeout(() => setSeconds(value => value - 1), 1000); return () => window.clearTimeout(timer); }, [seconds]);
  return { seconds, start: () => setSeconds(60), clear: () => setSeconds(0) };
}

export function AuthCodeInput({ value, onChange, disabled }: { value: string; onChange: (value: string) => void; disabled: boolean }) {
  return <div className="field"><label htmlFor="auth-code">Verification code</label><input id="auth-code" className="app-input hw-auth-code" name="code" inputMode="numeric" autoComplete="one-time-code" pattern="[0-9]{6}" maxLength={6} minLength={6} required value={value} onChange={event => onChange(event.target.value.replace(/\D/g, '').slice(0, 6))} disabled={disabled} aria-describedby="auth-code-help" /><p className="field-help" id="auth-code-help">Enter the six-digit code from your email. Codes expire after 15 minutes.</p></div>;
}

export function AuthRecoveryForm({ request, signInHref, homeHref }: { request: AuthFlowRequest; signInHref: string; homeHref: string }) {
  const lock = useRef(false), cooldown = useCooldown();
  const [sent, setSent] = useState(false), [email, setEmail] = useState('');
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function send(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault();
    if (lock.current || cooldown.seconds) return;
    lock.current = true; setBusy(true); setError('');
    try { await request('/auth/password-reset/request', { email: email.trim() }); setSent(true); cooldown.start(); }
    catch (err) { setError(err instanceof Error ? err.message : 'Recovery could not be started. Try again.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <AuthLayout title={sent ? 'Check your email' : 'Forgot your password?'} homeHref={homeHref}>
    {sent ? <><p className="hw-auth-note" role="status">If an account exists for <strong>{email}</strong>, check your inbox for a password reset link. The link expires in 15 minutes.</p>
      <div className="hw-auth-options"><button type="button" disabled={busy || cooldown.seconds > 0} onClick={() => void send()}>{cooldown.seconds ? `Resend in ${cooldown.seconds}s` : 'Resend reset link'}</button><button type="button" disabled={busy} onClick={() => { setSent(false); setError(''); cooldown.clear(); }}>Use another email</button></div></> :
      <form className="hw-auth-form" onSubmit={send} aria-busy={busy}><div className="field"><label htmlFor="recovery-email">Email address</label><input id="recovery-email" className="app-input" name="email" type="email" autoComplete="email" maxLength={320} required value={email} onChange={event => setEmail(event.target.value)} disabled={busy} /></div><AuthSubmit disabled={busy}>{busy ? 'Please wait...' : 'Send reset link'}</AuthSubmit></form>}
    {error && <p className="hw-auth-error" role="alert">{error}</p>}
    <p className="hw-auth-switch"><a href={signInHref}>Back to sign in</a></p>
  </AuthLayout>;
}

export function AuthResetPasswordForm({ request, signInHref, recoveryHref, homeHref }: { request: AuthFlowRequest; signInHref: string; recoveryHref: string; homeHref: string }) {
  const lock = useRef(false), captured = useRef(false);
  const [token, setToken] = useState<string | null>(null), [done, setDone] = useState(false);
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  useEffect(() => {
    if (captured.current) return;
    captured.current = true;
    const url = new URL(window.location.href), value = url.searchParams.get('key') || '';
    url.searchParams.delete('key');
    // Keep the secret only in component memory. Preserve Next's history state.
    window.history.replaceState(window.history.state, '', url.pathname + url.search);
    setToken(/^[a-f0-9]{64}$/.test(value) ? value : '');
  }, []);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current || !token) return;
    const data = new FormData(event.currentTarget), password = String(data.get('password') || '');
    if (password !== data.get('confirm_password')) { setError('The passwords do not match.'); return; }
    if (new TextEncoder().encode(password).length > 72) { setError('Your password must be no more than 72 bytes.'); return; }
    lock.current = true; setBusy(true); setError('');
    try { await request('/auth/password-reset/confirm', { token, password }); setDone(true); setToken(''); }
    catch (err) { setError(err instanceof Error ? err.message : 'The password could not be reset. Your entries are still here.'); }
    finally { lock.current = false; setBusy(false); }
  }
  return <AuthLayout title={done ? 'Password updated' : 'Set a new password'} homeHref={homeHref}>
    {done ? <><p className="hw-auth-note" role="status">Your password has changed. Sign in again on your devices.</p><a className="hw-auth-submit hw-auth-link-button" href={signInHref}>Back to sign in</a></> : token === null ? <p className="hw-auth-note" role="status">Opening your secure link...</p> : !token ? <><p className="hw-auth-error" role="alert">This page needs a valid reset link from your email. Request a new link if it has expired.</p><a className="hw-auth-submit hw-auth-link-button" href={recoveryHref}>Request reset link</a></> :
      <><form className="hw-auth-form" onSubmit={submit} aria-busy={busy}>
        <AuthPasswordInput name="password" label="New password" autoComplete="new-password" minLength={8} maxLength={72} required disabled={busy} helperText="Use at least 8 characters, up to 72 bytes." />
        <AuthPasswordInput name="confirm_password" label="Confirm new password" autoComplete="new-password" minLength={8} maxLength={72} required disabled={busy} />
        {error && <p className="hw-auth-error" role="alert">{error}</p>}
        <AuthSubmit disabled={busy}>{busy ? 'Please wait...' : 'Reset password'}</AuthSubmit>
      </form><p className="hw-auth-switch"><a href={recoveryHref}>Request a new link</a></p></>}
  </AuthLayout>;
}

export function AuthVerificationForm({ email, request, onVerified, onSignOut, initiallyQueued = false }: { email: string; request: AuthFlowRequest; onVerified: () => void; onSignOut: () => void; initiallyQueued?: boolean }) {
  const lock = useRef(false), cooldown = useCooldown(initiallyQueued ? 60 : 0);
  const [code, setCode] = useState(''), [busy, setBusy] = useState(false), [error, setError] = useState(''), [notice, setNotice] = useState(initiallyQueued ? 'Check your inbox for your verification code.' : '');
  async function act(confirm: boolean) {
    if (lock.current || !confirm && cooldown.seconds) return;
    lock.current = true; setBusy(true); setError(''); setNotice('');
    try { await request(confirm ? '/auth/email-verification/confirm' : '/auth/email-verification/request', confirm ? { code } : {}); if (confirm) onVerified(); else { cooldown.start(); setNotice('Check your inbox for a new code. Use the latest email.'); } }
    catch (err) { setError(err instanceof Error ? err.message : 'Email verification could not be completed. Try again.'); }
    finally { lock.current = false; setBusy(false); }
  }
  async function signOut() {
    if(lock.current) return;lock.current=true;setBusy(true);setError('');
    try {await request('/auth/logout',{});onSignOut();}catch(err) {setError(err instanceof Error ? err.message : 'Sign out failed. Try again.');}finally {lock.current=false;setBusy(false);}
  }
  return <><p className="hw-auth-note">Enter the code for <strong>{email}</strong>.</p><form className="hw-auth-form" onSubmit={event => { event.preventDefault(); void act(true); }} aria-busy={busy}>
    <AuthCodeInput value={code} onChange={setCode} disabled={busy} />
    {error && <p className="hw-auth-error" role="alert">{error}</p>}{notice && <p className="hw-auth-note" role="status">{notice}</p>}
    <AuthSubmit disabled={busy || code.length !== 6}>{busy ? 'Please wait...' : 'Verify email'}</AuthSubmit>
  </form><div className="hw-auth-options"><button type="button" disabled={busy || cooldown.seconds > 0} onClick={() => void act(false)}>{cooldown.seconds ? `Resend in ${cooldown.seconds}s` : 'Send another code'}</button><button type="button" disabled={busy} onClick={()=>void signOut()}>Use another account</button></div></>;
}

export function AuthPromotionUnsubscribe({ request, homeHref }: { request: AuthFlowRequest; homeHref: string }) {
 const captured=useRef(false),lock=useRef(false),[key,setKey]=useState<string|null>(null),[busy,setBusy]=useState(false),[done,setDone]=useState(false),[error,setError]=useState('');
 useEffect(()=>{if(captured.current)return;captured.current=true;const url=new URL(window.location.href),value=url.searchParams.get('key')||'';url.searchParams.delete('key');window.history.replaceState(window.history.state,'',url.pathname+url.search);setKey(/^[a-f0-9]{64}$/.test(value)?value:'');},[]);
 async function unsubscribe(){if(lock.current||!key)return;lock.current=true;setBusy(true);setError('');try{await request('/auth/promotions/unsubscribe',{key});setDone(true);setKey('');}catch(err){setError(err instanceof Error?err.message:'Your preference could not be saved. Try again.');}finally{lock.current=false;setBusy(false);}}
 return <AuthLayout title={done?'You are unsubscribed':'Email preferences'} homeHref={homeHref}>{done?<p className="hw-auth-note" role="status">Promotional email is turned off for this address. Account and event emails are unchanged.</p>:key===null?<p className="hw-auth-note" role="status">Opening your email preferences...</p>:!key?<p className="hw-auth-error" role="alert">Open the unsubscribe link from a Hackways promotional email.</p>:<><p className="hw-auth-note">Turn off Hackways news and promotional offers for the address that received this email.</p>{error&&<p className="hw-auth-error" role="alert">{error}</p>}<button type="button" className="hw-auth-submit" disabled={busy} onClick={()=>void unsubscribe()}>{busy?'Saving...':'Unsubscribe from promotions'}</button></>}<p className="hw-auth-switch"><a href={homeHref}>Back to Hackways</a></p></AuthLayout>;
}
