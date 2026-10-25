'use client';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
export function ApplicationWithdraw({ slug, revision }: { slug: string; revision: number }) {
 const router = useRouter(), [confirming, setConfirming] = useState(false), [busy, setBusy] = useState(false), [error, setError] = useState('');
 async function withdraw() { setBusy(true); setError(''); try { await fetchApi(`/programs/${slug}/application/withdraw`, { method: 'POST', body: JSON.stringify({ revision }) }); router.refresh(); } catch (cause) { setError(cause instanceof Error ? cause.message : 'Withdrawal failed. Please try again.'); } finally { setBusy(false); } }
 return <div className="program-withdraw">{confirming ? <><p>Withdraw this application? It will no longer be considered and cannot be resubmitted.</p><button className="program-button program-button-danger" disabled={busy} type="button" onClick={() => void withdraw()}>{busy ? 'Withdrawing…' : 'Confirm withdrawal'}</button><button className="program-text-link" type="button" disabled={busy} onClick={() => setConfirming(false)}>Keep my application</button></> : <button className="program-text-link" type="button" onClick={() => setConfirming(true)}>Withdraw application</button>}{error && <p className="program-error" role="alert">{error}</p>}</div>;
}
