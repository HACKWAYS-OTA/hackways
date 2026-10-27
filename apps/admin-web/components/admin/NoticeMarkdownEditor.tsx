'use client';
import dynamic from 'next/dynamic';
import { useDeferredValue, useRef, useState, type KeyboardEvent } from 'react';
import { RadioGroup } from '@hackways/ui/radio-group';
import '@hackways/ui/markdown.css';

const Preview = dynamic(() => import('@hackways/ui/markdown').then(module => module.NoticeMarkdown), { loading: () => <div className="admin-result-state admin-loading" role="status" aria-label="Loading preview"><span className="wordmark admin-loading-wordmark" aria-hidden="true">hackways.</span><span className="admin-loading-line" aria-hidden="true" /></div> });
const modes = [{ value:'write',label:'Write' },{ value:'split',label:'Side by side' },{ value:'preview',label:'Preview' }] as const;
export type NoticeImage = { url:string; width?:number; height?:number };
export function NoticeMarkdownEditor({ value, onChange, onUpload, disabled = false }: { value:string; onChange:(value:string)=>void; onUpload:(file:File)=>Promise<NoticeImage>; disabled?:boolean }) {
 const textarea = useRef<HTMLTextAreaElement>(null), picker = useRef<HTMLInputElement>(null), current = useRef(value);
 const history = useRef<{ past:string[]; future:string[]; time:number; typing:boolean }>({past:[],future:[],time:0,typing:false});
 const [mode,setMode] = useState<'write'|'split'|'preview'>('write'), [file,setFile] = useState<File|null>(null), [alt,setAlt] = useState(''), [uploading,setUploading] = useState(false), [error,setError] = useState(''), [,setHistoryVersion] = useState(0);
 const [images,setImages]=useState<{url:string;width:number;height:number}[]>([]);
 const preview = useDeferredValue(value); current.current = value;
 function commit(next:string, typing=false) {
  if (next===current.current) return;
  if (next.length>100000) { setError('Keep notice content within 100,000 characters.'); return; }
  const state=history.current, now=performance.now();
  if (!typing || !state.typing || now-state.time>700) state.past=[...state.past,current.current].slice(-25);
  state.future=[];state.typing=typing;state.time=now;onChange(next);current.current=next;setHistoryVersion(count=>count+1);
 }
 function restore(redo=false) {
  const state=history.current, source=redo?state.future:state.past, next=source.pop();if(next===undefined)return;
  if(redo)state.past.push(current.current);else state.future.push(current.current);
  state.typing=false;current.current=next;onChange(next);setHistoryVersion(count=>count+1);
 }
 function insert(before:string,after='',sample='text') {
  const field=textarea.current,start=field?.selectionStart??current.current.length,end=field?.selectionEnd??start,selected=current.current.slice(start,end)||sample;
  const replacement=before+selected+after;commit(current.current.slice(0,start)+replacement+current.current.slice(end));
  if(mode==='preview')setMode('write');requestAnimationFrame(()=>{textarea.current?.focus();textarea.current?.setSelectionRange(start+before.length,start+before.length+selected.length);});
 }
 function keys(event:KeyboardEvent<HTMLTextAreaElement>){
  if(!(event.ctrlKey||event.metaKey)||event.altKey)return;
  const key=event.key.toLowerCase();if(key==='b'||key==='i'){event.preventDefault();insert(key==='b'?'**':'*',key==='b'?'**':'*');}
  else if(key==='z'||key==='y'){event.preventDefault();restore(key==='y'||event.shiftKey);}
 }
 async function upload(){
  if(!file||uploading||disabled)return;if(!alt.trim()){setError('Describe the image before inserting it.');return}
  setUploading(true);setError('');
  try{const image=await onUpload(file);if(image.width&&image.height)setImages(previous=>[...previous,{url:image.url,width:image.width!,height:image.height!}].slice(-100));insert(`\n\n![${alt.trim().replace(/[\[\]\\]/g,'\\$&')}](${image.url})\n\n`,'','');setFile(null);setAlt('');}
  catch(failure){setError(failure instanceof Error?failure.message:'The image could not be uploaded. Retry with this file.');}
  finally{setUploading(false)}
 }
 return <section className="notice-composer" aria-label="Notice content editor">
  <header className="notice-composer-heading"><h2>Content</h2><RadioGroup label="Editor view" value={mode} options={modes} onValueChange={setMode} disabled={disabled||uploading} /></header>
  <div className="notice-format-toolbar" role="toolbar" aria-label="Text formatting">
   {([{label:'Bold',before:'**',after:'**'},{label:'Italic',before:'*',after:'*'},{label:'Heading',before:'\n\n## ',after:'\n\n',sample:'Heading'},{label:'List',before:'\n- ',after:'\n',sample:'List item'},{label:'Numbered list',before:'\n1. ',after:'\n',sample:'List item'},{label:'Quote',before:'\n> ',after:'\n'},{label:'Link',before:'[',after:'](https://)',sample:'Link text'},{label:'Code',before:'`',after:'`'},{label:'Table',before:'\n\n| Column | Column |\n| --- | --- |\n| ',after:' | Value |\n\n',sample:'Value'}]).map(item=><button key={item.label} type="button" disabled={disabled||uploading} onMouseDown={event=>event.preventDefault()} onClick={()=>insert(item.before,item.after,item.sample)}>{item.label}</button>)}
   <button type="button" disabled={disabled||uploading} onClick={()=>picker.current?.click()}>Image</button>
   <button type="button" disabled={disabled||uploading||!history.current.past.length} onClick={()=>restore()}>Undo</button><button type="button" disabled={disabled||uploading||!history.current.future.length} onClick={()=>restore(true)}>Redo</button>
  </div>
  <input ref={picker} type="file" accept="image/png,image/jpeg,image/webp" className="sr-only" tabIndex={-1} aria-label="Choose notice image" onChange={event=>{const selected=event.target.files?.[0];event.target.value='';if(!selected)return;if(!['image/png','image/jpeg','image/webp'].includes(selected.type)||selected.size<1||selected.size>5*1024*1024){setError('Choose a PNG, JPEG or WebP image up to 5 MB.');return}setFile(selected);setError('');}} />
  {file&&<div className="notice-image-insert"><strong>{file.name}</strong><p className="admin-secondary-copy">New notices are saved as drafts before uploading. Save again after inserting the image.</p><label className="field">Image description<input className="app-input" value={alt} onChange={event=>setAlt(event.target.value)} maxLength={300} disabled={uploading||disabled} placeholder="Describe what the image shows" /></label><div><button type="button" className="app-button app-button-primary app-button-sm" disabled={disabled||uploading} onClick={()=>void upload()}>{uploading?'Uploading...':'Upload and insert'}</button><button type="button" className="app-button app-button-outline app-button-sm" disabled={uploading||disabled} onClick={()=>setFile(null)}>Cancel</button></div></div>}
  {error&&<p className="journey-error" role="alert">{error}</p>}
  <div className={`notice-writing-surface notice-writing-${mode}`}>
   {mode!=='preview'&&<textarea ref={textarea} className="notice-markdown-source" value={value} onChange={event=>commit(event.target.value,true)} onKeyDown={keys} disabled={disabled||uploading} maxLength={100000} rows={20} aria-label="Notice Markdown content" spellCheck />}
   {mode!=='write'&&<div className="notice-markdown-preview" aria-label="Notice preview" aria-busy={preview!==value}>{preview?<Preview images={images}>{preview}</Preview>:<p className="admin-secondary-copy">Your formatted content will appear here.</p>}</div>}
  </div><p className="notice-editor-help">Markdown supports headings, lists, links, tables, code and images. Raw HTML is disabled.</p>
 </section>;
}
