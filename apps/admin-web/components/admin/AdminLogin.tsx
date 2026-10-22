"use client";

import { useState, type FormEvent } from 'react';
import { Input } from '@/components/ui/Input';
import { AuthPasswordInput, AuthSubmit } from '@hackways/ui/auth';

export function AdminLogin({ next }: { next: string }) {
  const [error,setError]=useState('');
  const [pending,setPending]=useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();setPending(true);setError('');
    const data=new FormData(event.currentTarget);
    const controller=new AbortController();const timer=window.setTimeout(()=>controller.abort(),15000);
    try {
      const base=(process.env.NEXT_PUBLIC_API_URL || '/api/v1').replace(/\/$/,'');
      const response=await fetch(`${base}/auth/admin/login`,{method:'POST',credentials:'include',headers:{'Content-Type':'application/json'},body:JSON.stringify({email:data.get('email'),password:data.get('password')}),signal:controller.signal});
      if(response.status===403){window.location.replace('/access-denied');return;}
      if(!response.ok){setError(response.status===401?'Check your email and password.':response.status===429?'Too many attempts. Try again in a minute.':'Sign in could not be completed. Try again.');return;}
      window.location.assign(next);
    } catch {setError('Sign in could not be completed. Try again.');}
    finally {window.clearTimeout(timer);setPending(false);}
  }
  return <form onSubmit={submit} className="hw-auth-form" aria-busy={pending}>
    <Input name="email" type="email" label="Email address" autoComplete="username" disabled={pending} required maxLength={320} />
    <AuthPasswordInput name="password" autoComplete="current-password" disabled={pending} required maxLength={72} />
    <a className="hw-auth-forgot" href={`${process.env.NEXT_PUBLIC_APP_URL || 'https://hackways.in'}/auth/forgot-password`}>Forgot password?</a>
    {error && <p className="hw-auth-error" role="alert">{error}</p>}
    <AuthSubmit disabled={pending}>{pending?'Signing in...':'Sign in'}</AuthSubmit>
  </form>;
}
