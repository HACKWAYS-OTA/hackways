'use client';
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Select } from '@hackways/ui/select';
import { adminRequest, mutationHeaders } from '@/lib/admin-request';
import type { Notice } from '@/lib/utilities';
import { NoticeMarkdownEditor, type NoticeImage } from './NoticeMarkdownEditor';

const states = [{ value: 'draft', label: 'Draft' }, { value: 'published', label: 'Published' }, { value: 'archived', label: 'Archived' }] as const;
async function imageDimensions(file:File) {
 const url=URL.createObjectURL(file);
 try {const image=new Image();image.src=url;await image.decode();if(image.naturalWidth<1||image.naturalHeight<1||image.naturalWidth>20000||image.naturalHeight>20000)throw new Error('Use an image with dimensions up to 20,000 pixels.');return {width:image.naturalWidth,height:image.naturalHeight};}
 finally{URL.revokeObjectURL(url);}
}
export function NoticeEditor({ initial }: { initial?: Notice }) {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [saved, setSaved] = useState(initial), [status, setStatus] = useState<Notice['status']>(initial?.status || 'draft');
  const [markdown, setMarkdown] = useState(initial?.body || '');
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  const receipt = useRef<{ body: string; reason: string; key: string }>();
  const uploadRef = useRef<{ file:File; resource:string; intent:{ upload_url:string; blob_key:string; expires_at:string }; uploaded:boolean; width:number; height:number }>();
  async function persist(form:FormData, targetStatus:Notice['status'], reason:string):Promise<Notice> {
    const fields = { reference:saved?.reference || 0, revision:saved?.revision || 0, slug:String(form.get('slug')).trim(), title:String(form.get('title')).trim(), summary:String(form.get('summary')).trim(), body:markdown.trim(), status:targetStatus };
    const body = JSON.stringify(fields);
    if (receipt.current?.body !== body || receipt.current.reason !== reason) receipt.current = { body, reason, key:crypto.randomUUID() };
    const result = await adminRequest<{ data: Pick<Notice,'id'|'reference'|'revision'|'slug'|'status'> }>('global/utilities/notices',{ method:'POST',headers:{...mutationHeaders(receipt.current.key),'X-Admin-Reason':encodeURIComponent(reason)},body },true);
    const notice = { ...fields,...result.data };setSaved(notice);return notice;
  }
  async function uploadImage(file:File):Promise<NoticeImage> {
    if (busy || !formRef.current) throw new Error('Wait for the current save to finish.');
    setBusy(true);setError('');setMessage('');
    try {
      const notice = saved || await persist(new FormData(formRef.current),'draft','Create notice draft for image upload');
      if (!uploadRef.current || uploadRef.current.file !== file || uploadRef.current.resource !== notice.id || !uploadRef.current.uploaded && Date.parse(uploadRef.current.intent.expires_at) <= Date.now()) {
        const size=await imageDimensions(file);
        const intent = await adminRequest<{ upload_url:string; blob_key:string; expires_at:string }>('global/media/notice-image/intent',{ method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({resource_id:notice.id,filename:file.name,mime_type:file.type,size:file.size,category:'notice-images'}) },true);
        uploadRef.current = {file,resource:notice.id,intent,uploaded:false,...size};
      }
      const upload = uploadRef.current;
      if (!upload.uploaded) { const controller = new AbortController(), timer = setTimeout(()=>controller.abort(),60000);try { const response = await fetch(upload.intent.upload_url,{method:'PUT',credentials:'omit',signal:controller.signal,headers:{'Content-Type':file.type,'x-ms-blob-type':'BlockBlob'},body:file});if(!response.ok&&response.status!==409&&response.status!==412)throw new Error('The image could not be uploaded. Retry with this file.');upload.uploaded=true; } finally {clearTimeout(timer);} }
      const image = await adminRequest<NoticeImage>('global/media/notice-image/complete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({resource_id:notice.id,blob_key:upload.intent.blob_key,width:upload.width,height:upload.height})},true);
      setMessage('Image uploaded. Save the notice to keep it in the content.');return {...image,width:upload.width,height:upload.height};
    } finally {setBusy(false);}
  }
  async function reloadSaved() {
    if (busy || !formRef.current) return;setBusy(true);
    try {const path=saved?`global/utilities/notices/${saved.reference}`:`global/utilities/notices/by-slug/${encodeURIComponent(String(new FormData(formRef.current).get('slug')).trim())}`;
      const result = await adminRequest<{data:Notice}>(path,{},true);setSaved(result.data);const address=formRef.current.elements.namedItem('slug');if(address instanceof HTMLInputElement)address.value=result.data.slug;setError('');setMessage('Latest saved version loaded. Your unsaved content is preserved; review it before saving.');}
    catch(failure){setError(failure instanceof Error?failure.message:'The saved version could not be loaded.');}finally{setBusy(false);}
  }
  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (busy) return;
    const form = new FormData(event.currentTarget), reason = String(form.get('reason')).trim();
    setBusy(true); setError(''); setMessage('');
    try {
      const notice = await persist(form,status,reason); setMessage(notice.status === 'published' ? 'Notice published.' : 'Notice saved.');
      if (!initial) router.replace(`/utilities/notice/${notice.reference}`); else router.refresh();
    } catch (failure) { setError(failure instanceof Error ? failure.message : 'The notice could not be saved. Your content is still here.'); }
    finally { setBusy(false); }
  }
  return <form ref={formRef} className="utility-editor" onSubmit={save}><fieldset className="utility-editor-fields" disabled={busy}>
    <div className="utility-editor-main"><label className="field">Title<input className="app-input" name="title" defaultValue={initial?.title} required maxLength={180} /></label>
      <label className="field">Summary<textarea className="app-input" name="summary" defaultValue={initial?.summary} rows={3} maxLength={500} required={status === 'published'} /></label>
      <NoticeMarkdownEditor value={markdown} onChange={setMarkdown} onUpload={uploadImage} disabled={busy} />
    </div><aside className="utility-editor-settings"><label className="field">Page address<span className="utility-path-prefix">/notice/</span><input className="app-input" name="slug" defaultValue={initial?.slug} readOnly={Boolean(saved)} required pattern="[a-z0-9][a-z0-9-]{0,119}" maxLength={120} placeholder="notice-2026-10-1" /></label>
      <label className="field">Visibility<Select label="Notice visibility" value={status} options={states} onValueChange={setStatus} disabled={busy} /></label>
      <p className="admin-secondary-copy">Draft and archived notices are private. The page address stays fixed after creation.</p>
      <label className="field">Change note<input className="app-input" name="reason" required minLength={3} maxLength={1000} /></label>
      <button className="app-button app-button-primary" type="submit" disabled={busy}>{busy ? 'Saving...' : status === 'published' ? 'Publish notice' : 'Save notice'}</button>
      {saved?.status === 'published' && <a className="app-button app-button-outline" target="_blank" rel="noopener noreferrer" href={`https://hackways.in/notice/${saved.slug}`}>Open public notice</a>}
      {saved && <a href={`/utilities/seo-manager/edit?path=${encodeURIComponent(`/notice/${saved.slug}`)}`}>Manage page SEO</a>}
      {error && <p className="journey-error" role="alert">{error} <button type="button" onClick={() => void reloadSaved()} disabled={busy}>Reload saved version</button></p>}{message && <p role="status">{message}</p>}
    </aside></fieldset>
  </form>;
}
