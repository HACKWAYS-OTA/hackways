'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/Button';

export function AdminAccount({ label = 'Sign out' }: { label?: string }) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const base = (process.env.NEXT_PUBLIC_API_URL || '/api/v1').replace(/\/$/, '');
  async function signOut() {
    setPending(true); setError('');
    const controller = new AbortController();
    const timer = window.setTimeout(() => controller.abort(), 15000);
    try {
      const response = await fetch(`${base}/auth/admin/logout`, { method: 'POST', credentials: 'include', signal: controller.signal });
      if (!response.ok) { setError('Sign out failed. Try again.'); return; }
      window.location.replace('/login');
    } catch { setError('Sign out failed. Try again.'); }
    finally { window.clearTimeout(timer); setPending(false); }
  }
  return <div className="admin-account">
    <Button onClick={signOut} variant="ghost" size="sm" disabled={pending}>{label}</Button>
    {error && <span className="field-error" role="alert">{error}</span>}
  </div>;
}
