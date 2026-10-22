"use client";
import { useRef, useState, type FormEvent } from 'react';
import { Checkbox } from '@hackways/ui/checkbox';
import { fetchApi } from '@/lib/api';
export function PromotionPreference({ subscribed, verified }: { subscribed: boolean; verified: boolean }) {
 const lock = useRef(false), [selected,setSelected] = useState(subscribed), [saved,setSaved] = useState(subscribed), [busy,setBusy] = useState(false), [error,setError] = useState(''), [notice,setNotice] = useState('');
 async function submit(event: FormEvent<HTMLFormElement>) {
  event.preventDefault(); if (lock.current) return;
  lock.current=true;setBusy(true);setError('');setNotice('');
  try { const result=await fetchApi<{subscribed:boolean}>('/auth/promotions/preference',{method:'POST',body:JSON.stringify({subscribed:selected})});setSaved(result.subscribed);setSelected(result.subscribed);setNotice('Email preference saved.'); }
  catch(err) {setError(err instanceof Error ? err.message : 'Your preference could not be saved.');}
  finally {lock.current=false;setBusy(false);}
 }
 return <section className="promotion-preference"><h2>Hackways news and offers</h2><form onSubmit={submit}><label className="promotion-choice"><Checkbox checked={selected} onCheckedChange={setSelected} disabled={busy} />Email me Hackways promotional updates.</label><p>Account security and event emails are separate. You can unsubscribe from promotions at any time.</p>{!verified && <p>Verify your email before turning promotional updates on.</p>}{error && <p className="journey-error" role="alert">{error}</p>}{notice && <p className="journey-notice" role="status">{notice}</p>}<button type="submit" className="app-button app-button-outline" disabled={busy || selected===saved || selected && !verified}>{busy ? 'Saving...' : 'Save email preference'}</button></form></section>;
}
