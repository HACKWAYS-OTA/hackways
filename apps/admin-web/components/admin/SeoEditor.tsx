'use client';
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Checkbox } from '@hackways/ui/checkbox';
import { adminRequest, mutationHeaders } from '@/lib/admin-request';
import type { SeoPage, SeoTopic } from '@/lib/utilities';
import { SeoKeywordLibrary } from './SeoKeywordLibrary';

export function SeoEditor({ initial, path = '/', topics }: { initial?: SeoPage; path?: string; topics:SeoTopic[]|null }) {
 const router = useRouter(); const [saved, setSaved] = useState(initial), [noindex, setNoindex] = useState(initial?.noindex || false), [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
 const receipt = useRef<{ body: string; reason: string; key: string }>();
 const formRef = useRef<HTMLFormElement>(null);
 async function reloadSaved(){
  if(busy||!formRef.current)return;setBusy(true);
  try{const selected=String(new FormData(formRef.current).get('path')).trim();const result=await adminRequest<{data:SeoPage|null}>(`global/utilities/seo/page?path=${encodeURIComponent(selected)}`,{},true);setSaved(result.data||undefined);setError('');setMessage('Latest saved version loaded. Your unsaved entries are preserved; review them before saving.');}
  catch(failure){setError(failure instanceof Error?failure.message:'The saved version could not be loaded.');}finally{setBusy(false);}
 }
 function useDefaults(){
  if(!formRef.current||busy)return;
  for(const name of ['title','description','canonical_path']){const field=formRef.current.elements.namedItem(name);if(field instanceof HTMLInputElement||field instanceof HTMLTextAreaElement)field.value='';}
  setNoindex(false);setMessage('Page defaults selected. Save to apply.');
 }
 function addKeywords(words:string[]){
  if(!formRef.current||busy)return;const field=formRef.current.elements.namedItem('planning_keywords');if(!(field instanceof HTMLTextAreaElement))return;
  const existing=field.value.split(/\r?\n/).map(value=>value.trim()).filter(Boolean),seen=new Set(existing.map(value=>value.toLowerCase()));const added=words.filter(word=>{if(seen.has(word.toLowerCase()))return false;seen.add(word.toLowerCase());return true;});
  if(existing.length+added.length>2000){setError('This plan holds up to 2,000 keywords. Remove some terms before adding more.');return}
  field.value=[...existing,...added].join('\n');setError('');setMessage(added.length?`${added.length} keyword ideas added to your draft. Save to apply.`:'These terms are already in your plan.');
 }
 async function save(event: FormEvent<HTMLFormElement>) {
  event.preventDefault(); if (busy) return;
  const form = new FormData(event.currentTarget), fields = { revision: saved?.revision || 0, path: String(form.get('path')).trim(), title: String(form.get('title')).trim(), description: String(form.get('description')).trim(), canonical_path: String(form.get('canonical_path')).trim(), noindex, planning_keywords: String(form.get('planning_keywords')).split(/\r?\n/).map(value => value.trim()).filter(Boolean), notes: String(form.get('notes')).trim() };
  const body = JSON.stringify(fields), reason = String(form.get('reason')).trim(); if (receipt.current?.body !== body || receipt.current.reason !== reason) receipt.current = { body, reason, key: crypto.randomUUID() };
  setBusy(true); setError(''); setMessage('');
  try { const result = await adminRequest<{ data: Pick<SeoPage, 'revision' | 'reference' | 'path'> }>('global/utilities/seo', { method: 'POST', headers: { ...mutationHeaders(receipt.current.key), 'X-Admin-Reason': encodeURIComponent(reason) }, body }, true); setSaved({ ...fields, ...result.data }); setMessage('SEO settings saved. Public metadata refreshes within 30 seconds.'); if (!saved) router.replace(`/utilities/seo-manager/edit?path=${encodeURIComponent(result.data.path)}`); else router.refresh(); }
  catch (failure) { setError(failure instanceof Error ? failure.message : 'Settings could not be saved. Your entries are still here.'); } finally { setBusy(false); }
 }
 return <form ref={formRef} className="utility-editor" onSubmit={save}><fieldset className="utility-editor-fields" disabled={busy}><div className="utility-editor-main">
  <label className="field">Public page path<input name="path" className="app-input" defaultValue={initial?.path || path} readOnly={Boolean(saved)} required maxLength={240} placeholder="/" /></label>
  <label className="field">Search title<input name="title" className="app-input" defaultValue={initial?.title} maxLength={90} /><small>Leave blank to use the page title.</small></label>
  <label className="field">Search description<textarea name="description" className="app-input" defaultValue={initial?.description} maxLength={320} rows={4} /><small>Describe what this page offers. Leave blank to use its own description.</small></label>
  <label className="field">Canonical path<input name="canonical_path" className="app-input" defaultValue={initial?.canonical_path} maxLength={240} placeholder="Use the page's canonical URL" /></label>
  <label className="field">Planning keywords<textarea name="planning_keywords" className="app-input" rows={10} defaultValue={initial?.planning_keywords.join('\n')} maxLength={160000} /><small>Up to 2,000 relevant terms, one per line. Internal planning only; these are never injected into pages or meta-keyword tags.</small></label>
  <SeoKeywordLibrary topics={topics} onAdd={addKeywords} disabled={busy} />
  <label className="field">Editorial notes<textarea name="notes" className="app-input" rows={4} defaultValue={initial?.notes} maxLength={4000} /></label>
 </div><aside className="utility-editor-settings"><h2>Search visibility</h2><label className="checkbox-label"><Checkbox checked={noindex} onCheckedChange={setNoindex} disabled={busy} />Keep this page out of search results</label><p className="admin-secondary-copy">Private account, registration, application, management and admin pages stay excluded. Settings do not publish private content or guarantee rankings.</p><label className="field">Change note<input name="reason" className="app-input" required minLength={3} maxLength={1000} /></label><button className="app-button app-button-primary" disabled={busy}>{busy ? 'Saving...' : 'Save SEO settings'}</button><button type="button" className="app-button app-button-outline" disabled={busy} onClick={useDefaults}>Use page defaults</button>{saved && <a href={`https://hackways.in${saved.path}`} target="_blank" rel="noopener noreferrer">Open public page</a>}{error && <p className="journey-error" role="alert">{error} <button type="button" disabled={busy} onClick={() => void reloadSaved()}>Reload saved version</button></p>}{message && <p role="status">{message}</p>}</aside></fieldset></form>;
}
