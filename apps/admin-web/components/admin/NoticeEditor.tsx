'use client';
import { useEffect, useRef, useState } from 'react';
import { HugeiconsIcon } from '@hugeicons/react';
import { Settings01Icon, ArrowLeft01Icon, Menu01Icon, File02Icon, ArrowUpRight01Icon } from '@hugeicons/core-free-icons';
import type { NoticeDocumentNode, NoticeImage } from '@hackways/ui/document';
import { adminRequest, mutationHeaders, AdminRequestError, AdminAuthenticationError } from '@/lib/admin-request';
import type { Notice } from '@/lib/utilities';
import { NoticeDocumentEditor, type NoticeHeading, type NoticeEditorHandle } from './NoticeDocumentEditor';
async function imageDimensions(file:File) {
 const url=URL.createObjectURL(file);
 try {const image=new Image();image.src=url;await image.decode();if(image.naturalWidth<1||image.naturalHeight<1||image.naturalWidth>20000||image.naturalHeight>20000)throw new Error('Use an image with dimensions up to 20,000 pixels.');return {width:image.naturalWidth,height:image.naturalHeight};}
 finally{URL.revokeObjectURL(url);}
}

type DraftFields={title:string;slug:string;summary:string;body:string;document:NoticeDocumentNode|null};
function authored(fields:DraftFields){return fields.title.trim()||fields.body.trim()||Boolean(fields.document?.content?.some(node=>node.type!=='paragraph'||node.content?.length));}
export function NoticeEditor({initial}:{initial?:Notice}) {
 const [fields,setFields]=useState<DraftFields>({title:initial?.title||'',slug:initial?.slug||'',summary:initial?.summary||'',body:initial?.body||'',document:initial?.document||null});
 const [saved,setSaved]=useState(initial),[saving,setSaving]=useState(false),[manualBusy,setManualBusy]=useState(false),[uploadBusy,setUploadBusy]=useState(false),[error,setError]=useState(''),[message,setMessage]=useState(''),[note,setNote]=useState('');
 const fieldsRef=useRef(fields);fieldsRef.current=fields;
 const savedRef=useRef(saved),settings=useRef<HTMLDialogElement>(null),queue=useRef<Promise<Notice>>(),lastSaved=useRef(JSON.stringify(fields)),dirty=JSON.stringify(fields)!==lastSaved.current;
 const receipt=useRef<{body:string;reason:string;key:string;pending:boolean;snapshot:DraftFields;status:Notice['status']}>();
 const uploadRef=useRef<{file:File;resource:string;intent:{upload_url:string;blob_key:string;expires_at:string};uploaded:boolean;width:number;height:number}>();
 const titleRef=useRef<HTMLInputElement>(null),documentEditor=useRef<NoticeEditorHandle>(null);
 const [collapsed,setCollapsed]=useState(true),[outline,setOutline]=useState<NoticeHeading[]>([]),[words,setWords]=useState(0);
 function editorReady(document:NoticeDocumentNode){
  const baseline=JSON.parse(lastSaved.current) as DraftFields;
  lastSaved.current=JSON.stringify({...baseline,document});
  setFields(current=>({...current,document}));
 }

 async function persist(target:Notice['status'],reason:string):Promise<Notice> {
  if(queue.current){try{await queue.current;}catch{/* The following explicit operation retries with its own payload. */}}
  setSaving(true);
  const operation=async()=>{
   async function submit(request:NonNullable<typeof receipt.current>):Promise<Notice>{
    let result:{data:Pick<Notice,'id'|'reference'|'revision'|'slug'|'status'|'has_published'|'summary'>};
    try{result=await adminRequest<typeof result>('global/utilities/notices',{method:'POST',headers:{...mutationHeaders(request.key),'X-Admin-Reason':encodeURIComponent(request.reason)},body:request.body},true);}
    catch(failure){
     // A definitive rejection has no committed write to recover. Corrected input
     // and refreshed revisions must be allowed to issue a new request.
     if(failure instanceof AdminAuthenticationError||failure instanceof AdminRequestError&&failure.status>=400&&failure.status<500)request.pending=false;
     throw failure;
    }
    const snapshot=request.snapshot;
    const record:Notice={...snapshot,...result.data,body_format:snapshot.document?'richtext':'markdown'};
    request.pending=false;savedRef.current=record;setSaved(record);
    // Preserve edits made while this snapshot was travelling to the server.
    const current=fieldsRef.current,next={...current};
    if(current.slug===snapshot.slug)next.slug=record.slug;
    if(current.summary===snapshot.summary)next.summary=record.summary;
    fieldsRef.current=next;setFields(next);
    lastSaved.current=JSON.stringify({...snapshot,slug:record.slug,summary:record.summary});
    if(!initial&&window.location.pathname.endsWith('/new'))window.history.replaceState(null,'',`/utilities/notice/${record.reference}`);
    return record;
   }
   // Resolve an uncertain response with the exact original receipt before any
   // subsequent write. This also keeps image-first draft creation exactly once.
   if(receipt.current?.pending){
    const unresolved=receipt.current;
    const sameFields=JSON.stringify(unresolved.snapshot)===JSON.stringify(fieldsRef.current);
    const confirmed=await submit(unresolved);
    if(sameFields&&unresolved.status===target)return confirmed;
   }
   const snapshot={...fieldsRef.current},previous=savedRef.current;
   const payload={...snapshot,reference:previous?.reference||0,revision:previous?.revision||0,status:target,body_format:snapshot.document?'richtext':'markdown'};
   const request={body:JSON.stringify(payload),reason,key:crypto.randomUUID(),pending:true,snapshot,status:target};
   receipt.current=request;
   return await submit(request);
  };
  const promise=operation();queue.current=promise;
  try{return await promise;}finally{if(queue.current===promise){queue.current=undefined;setSaving(false);}}
 }
 useEffect(()=>{
  if(!dirty||!authored(fields)||manualBusy||uploadBusy||error||saved?.status==='published'||saved?.status==='archived')return;
  const timer=setTimeout(()=>{void persist('draft','Autosave notice draft').catch(failure=>setError(failure instanceof Error?failure.message:'The draft could not be saved. Your edits are still here.'));},2000);
  return()=>clearTimeout(timer);
 // Persist reads the latest field and revision references; each request is serialized.
 // eslint-disable-next-line react-hooks/exhaustive-deps
 },[fields,dirty,manualBusy,uploadBusy,error,saved?.status]);
 useEffect(()=>{const protect=(event:BeforeUnloadEvent)=>{if(dirty||saving){event.preventDefault();event.returnValue='';}};window.addEventListener('beforeunload',protect);return()=>window.removeEventListener('beforeunload',protect);},[dirty,saving]);
 function change<K extends keyof DraftFields>(key:K,value:DraftFields[K]){setFields(current=>({...current,[key]:value}));setMessage('');}
 async function save(target:Notice['status']){
  if(manualBusy||uploadBusy)return;
  if(target==='published'&&!fieldsRef.current.title.trim()){setError('Give your notice a title before publishing.');titleRef.current?.focus();return;}
  setManualBusy(true);setError('');setMessage('');
  try{await persist(target,note.trim()||({draft:'Save notice draft',published:'Publish notice document',archived:'Archive notice'}[target]));setMessage(target==='published'?'Notice published.':target==='archived'?'Notice archived.':'Draft saved.');settings.current?.close();}
  catch(failure){setError(failure instanceof Error?failure.message:'The notice could not be saved. Your edits are still here.');}
  finally{setManualBusy(false);}
 }
 async function uploadImage(file:File):Promise<NoticeImage>{
  const notice=savedRef.current||await persist('draft','Create notice draft for image upload');
  if(!uploadRef.current||uploadRef.current.file!==file||uploadRef.current.resource!==notice.id||!uploadRef.current.uploaded&&Date.parse(uploadRef.current.intent.expires_at)<=Date.now()){
   const size=await imageDimensions(file);
   const intent=await adminRequest<{upload_url:string;blob_key:string;expires_at:string}>('global/media/notice-image/intent',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({resource_id:notice.id,filename:file.name,mime_type:file.type,size:file.size,category:'notice-images'})},true);
   uploadRef.current={file,resource:notice.id,intent,uploaded:false,...size};
  }
  const upload=uploadRef.current;
  if(!upload.uploaded){const controller=new AbortController(),timer=setTimeout(()=>controller.abort(),60000);try{const response=await fetch(upload.intent.upload_url,{method:'PUT',credentials:'omit',signal:controller.signal,headers:{'Content-Type':file.type,'x-ms-blob-type':'BlockBlob'},body:file});if(!response.ok&&response.status!==409&&response.status!==412)throw new Error('The image could not be uploaded. Retry with this file.');upload.uploaded=true;}finally{clearTimeout(timer);}}
  const result=await adminRequest<NoticeImage>('global/media/notice-image/complete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({resource_id:notice.id,blob_key:upload.intent.blob_key,width:upload.width,height:upload.height})},true);
  return {...result,width:upload.width,height:upload.height};
 }
 async function reloadVersion(){
  if(!savedRef.current||saving||manualBusy)return;setManualBusy(true);
  try{const result=await adminRequest<{data:Notice}>(`global/utilities/notices/${savedRef.current.reference}`,{},true);savedRef.current=result.data;setSaved(result.data);setError('');setMessage('Saved version checked. Your unsaved edits are preserved; review them before saving.');}
  catch(failure){setError(failure instanceof Error?failure.message:'The saved version could not be checked.');}finally{setManualBusy(false);}
 }
 const published=Boolean(saved?.has_published||saved?.published_at),locked=manualBusy||uploadBusy;
 return <div className="notice-studio notice-workspace" data-sidebar-collapsed={collapsed}>
  <header className="notice-studio-header">
   <div className="notice-workspace-brand"><button type="button" className="notice-tool" aria-label={collapsed?'Expand editor sidebar':'Collapse editor sidebar'} aria-expanded={!collapsed} aria-controls="notice-editor-sidebar" onClick={()=>setCollapsed(value=>!value)}><HugeiconsIcon icon={Menu01Icon} size={21}/></button><a href="/utilities/notice" className="wordmark">hackways.</a></div>
   <div className="notice-document-identity"><label className="sr-only" htmlFor="notice-title">Notice title</label><input ref={titleRef} id="notice-title" className="notice-title-input" placeholder="Untitled notice" value={fields.title} maxLength={180} disabled={locked} onChange={event=>change('title',event.target.value)}/><span className="notice-save-state" role="status">{uploadBusy?'Uploading image...':saving?'Saving...':error?'Changes not saved':dirty?saved?.status==='published'?'Unpublished changes':'Unsaved changes':saved?saved.status==='published'?'Published':saved.status==='archived'?'Archived':'All changes saved':'Draft'}</span></div>
   <div className="notice-studio-actions"><button type="button" className="notice-tool" onClick={()=>settings.current?.showModal()} aria-label="Publishing settings" title="Publishing settings"><HugeiconsIcon icon={Settings01Icon} size={20}/></button>{saved?.status!=='published'&&<button type="button" className="notice-save-button" disabled={locked||saving} onClick={()=>void save('draft')}>Save draft</button>}<button type="button" className="app-button app-button-primary" disabled={locked||saving} onClick={()=>void save('published')}>{manualBusy?'Saving...':saved?.status==='published'?'Publish changes':'Publish'}</button></div>
  </header>
  <aside id="notice-editor-sidebar" className="notice-editor-sidebar" aria-label="Document navigation">
   {collapsed?<div className="notice-sidebar-rail"><button type="button" className="notice-tool" aria-label="Show document outline" title="Document outline" onClick={()=>setCollapsed(false)}><HugeiconsIcon icon={File02Icon} size={21}/></button><button type="button" className="notice-tool" aria-label="Publishing settings" title="Publishing settings" onClick={()=>settings.current?.showModal()}><HugeiconsIcon icon={Settings01Icon} size={21}/></button></div>:<><section className="notice-outline"><h2>Document outline</h2>{outline.length?<nav aria-label="Document headings">{outline.map(heading=><button key={heading.position} type="button" data-level={heading.level} onClick={()=>documentEditor.current?.focusHeading(heading.position)}>{heading.title}</button>)}</nav>:<p>Headings appear here as you write.</p>}</section><div className="notice-sidebar-links"><button type="button" onClick={()=>settings.current?.showModal()}><HugeiconsIcon icon={Settings01Icon} size={19}/>Publishing settings</button>{saved&&<a href={`/utilities/seo-manager/edit?path=${encodeURIComponent(`/notice/${saved.slug}`)}`} target="_blank" rel="noopener noreferrer">Page SEO<HugeiconsIcon icon={ArrowUpRight01Icon} size={17}/></a>}{saved?.status==='published'&&<a href={`https://hackways.in/notice/${saved.slug}`} target="_blank" rel="noopener noreferrer">View public notice<HugeiconsIcon icon={ArrowUpRight01Icon} size={17}/></a>}</div></>}
   <a className="notice-sidebar-back" href="/utilities/notice" aria-label="Back to notices" title="Back to notices"><HugeiconsIcon icon={ArrowLeft01Icon} size={20}/>{!collapsed&&'Back to notices'}</a>
  </aside>
  <div className="notice-editor-stage">
  {error&&<div className="notice-studio-feedback journey-error" role="alert"><p>{error}</p><button type="button" disabled={locked||saving} onClick={()=>void save(saved?.status==='published'?'published':'draft')}>Retry save</button>{saved&&<button type="button" disabled={locked||saving} onClick={()=>void reloadVersion()}>Check saved version</button>}</div>}
  {message&&<p className="notice-studio-feedback" role="status">{message}</p>}
  <main id="notice-editor-content" className="notice-editor-main"><NoticeDocumentEditor ref={documentEditor} onReady={editorReady} onStructure={(headings,count)=>{setOutline(headings);setWords(count);}} initialDocument={initial?.document} initialMarkdown={initial?.body||''} disabled={manualBusy} onChange={document=>change('document',document)} onUpload={uploadImage} onUploadBusy={setUploadBusy}/></main><footer className="notice-editor-statusbar"><span>{words} {words===1?'word':'words'}</span><span>{saved?.status==='published'?'Changes go live when you publish.':saved?.status==='archived'?'Archived notice':'Drafts save automatically.'}</span></footer>
  </div>
  <dialog ref={settings} className="notice-editor-dialog notice-publish-settings" aria-labelledby="notice-settings-title"><h2 id="notice-settings-title">Publishing settings</h2><p>{published?'This notice has a permanent public address.':'Choose an address before publishing. Drafts save automatically while you write.'}</p><label className="field">Page address<span className="utility-path-prefix">/notice/</span><input className="app-input" value={fields.slug} readOnly={published} disabled={locked} maxLength={120} placeholder="notice-2026-10-1" onChange={event=>change('slug',event.target.value)}/></label><label className="field">Search summary <span className="admin-secondary-copy">Optional</span><textarea className="app-input" rows={3} maxLength={500} value={fields.summary} disabled={locked} placeholder="Leave blank to use the opening text of your notice." onChange={event=>change('summary',event.target.value)}/></label><label className="field">Change note <span className="admin-secondary-copy">Optional</span><input className="app-input" value={note} maxLength={1000} disabled={locked} onChange={event=>setNote(event.target.value)}/></label>{saved?.status==='published'&&<div className="notice-published-links"><a href={`https://hackways.in/notice/${saved.slug}`} target="_blank" rel="noopener noreferrer">Open public notice</a><a href={`/utilities/seo-manager/edit?path=${encodeURIComponent(`/notice/${saved.slug}`)}`} target="_blank" rel="noopener noreferrer">Manage page SEO</a></div>}<div className="notice-dialog-actions">{saved&&saved.status!=='archived'&&<button type="button" className="notice-archive-button" disabled={locked||saving} onClick={()=>void save('archived')}>Archive</button>}<button className="app-button app-button-outline" type="button" onClick={()=>settings.current?.close()}>Done</button></div></dialog>
 </div>;
}
