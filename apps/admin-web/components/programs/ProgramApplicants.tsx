'use client';
import Link from 'next/link';
import { useState } from 'react';
import { Checkbox } from '@hackways/ui/checkbox';
import type { ProgramApplicationSummary } from '@hackways/api-client';
import { AdminAuthenticationError } from '@/lib/admin-request';
import { ProgramAuthRecovery } from './ProgramAuthRecovery';

export function ProgramApplicants({ slug, applications, search, decision }: { slug: string; applications: ProgramApplicationSummary[]; search: string; decision: string }) {
  const [selected, setSelected] = useState<string[]>([]), [busy, setBusy] = useState(false), [error, setError] = useState<Error | null>(null);
  const base = `/programs/${slug}/communications`;
  async function download() {
    if (busy) return; setBusy(true); setError(null);
    const controller = new AbortController(), timeout = setTimeout(() => controller.abort(), 35000);
    try {
      const response = await fetch(`/api/v1/admin/global/programs/${slug}/applications/export.xlsx?q=${encodeURIComponent(search)}&decision=${encodeURIComponent(decision)}`, { credentials: 'include', cache: 'no-store', signal: controller.signal });
      if (response.status === 401) throw new AdminAuthenticationError(false);
      if (!response.ok) {
        const data: { error?: { code?: string; message?: string } } = await response.json();
        if (data.error?.code === 'ADMIN_VERIFICATION_REQUIRED') throw new AdminAuthenticationError(true);
        throw new Error(data.error?.message || 'The Excel export could not be prepared. Try again.');
      }
      if (!response.headers.get('Content-Type')?.includes('spreadsheetml.sheet')) throw new Error('The export response was not an Excel workbook. Try again.');
      const blob = await response.blob(), url = URL.createObjectURL(blob), anchor = document.createElement('a');
      anchor.href = url; anchor.download = response.headers.get('Content-Disposition')?.match(/filename="?([^";]+)"?/)?.[1] || `${slug}-applicants.xlsx`;
      anchor.click(); setTimeout(() => URL.revokeObjectURL(url), 1500);
    } catch (cause) { setError(cause instanceof Error ? cause : new Error('The export could not be downloaded. Try again.')); }
    finally { clearTimeout(timeout); setBusy(false); }
  }
  return <>
    <div className="program-applicant-actions">
      <button className="app-button app-button-outline" type="button" disabled={busy} onClick={() => void download()}>{busy ? 'Preparing Excel…' : 'Download Excel (.xlsx)'}</button>
      <Link className="app-button app-button-outline" href={`${base}?scope=filtered&q=${encodeURIComponent(search)}&decision=${encodeURIComponent(decision)}`}>Message this group</Link>
      <Link className="app-button app-button-primary" href={selected.length ? `${base}?ids=${selected.join(',')}` : base}>{selected.length ? `Message ${selected.length} selected` : 'Message applicants'}</Link>
    </div>
    <p className="program-admin-help">Excel includes all submitted applicants matching these filters. Select applicants below to message a specific group.</p>
    {error && <p className="program-admin-error" role="alert">{error.message}</p>}<ProgramAuthRecovery error={error} />
    {applications.length > 0 && <label className="program-page-select"><Checkbox checked={selected.length === applications.length} onCheckedChange={checked => setSelected(checked ? applications.map(a => a.id) : [])} />Select all on this page <span aria-live="polite">({selected.length} selected)</span></label>}
    <div className="program-admin-applications">{applications.map(a => <div key={a.id} className="program-selectable-applicant">
      <Checkbox aria-label={`Select ${a.full_name}`} checked={selected.includes(a.id)} onCheckedChange={checked => setSelected(value => checked ? [...value, a.id] : value.filter(id => id !== a.id))} />
      <Link className="program-admin-application-row" href={`/programs/${slug}/applications/${a.id}`}><div><h2>{a.full_name}</h2><p>{a.email} · {a.city}</p></div><div><span className="program-admin-badge">{a.decision === 'pending' ? 'Needs review' : a.decision}</span><small>{a.published_at ? 'Result released' : 'Private'}</small></div></Link>
    </div>)}{!applications.length && <div className="admin-result-state">No submitted applications match this view.</div>}</div>
  </>;
}
