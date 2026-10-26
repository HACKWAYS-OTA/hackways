'use client';
import Link from 'next/link';
import { useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { Avatar } from '@hackways/ui/avatar';
import { RadioGroup } from '@hackways/ui/radio-group';
import { Checkbox } from '@hackways/ui/checkbox';
import type { Program, ProgramDecision, AdminProgramApplication } from '@hackways/api-client';
import { adminRequest, AdminRequestError, AdminAuthenticationError } from '@/lib/admin-request';
import { ProgramAuthRecovery } from './ProgramAuthRecovery';

type Update = {decision: ProgramDecision;notes: string;revision: number;confirmed: boolean};
export function ProgramReview({ program, application: initial }: { program: Program; application: AdminProgramApplication }) {
 const router=useRouter();
 const [application,setApplication]=useState(initial),[decision,setDecision]=useState<ProgramDecision>(initial.decision),[notes,setNotes]=useState(initial.reviewer_notes);
 const [editing,setEditing]=useState(false),[confirmed,setConfirmed]=useState(false),[busy,setBusy]=useState(false),[uncertain,setUncertain]=useState(false),[error,setError]=useState<Error|null>(null),[notice,setNotice]=useState('');
 const attempt=useRef<{key:string;body:Update}|null>(null);
 const published=Boolean(application.published_at),locked=application.status!=='submitted'||(published&&!editing);
 function reset() {setEditing(false);setDecision(application.decision);setNotes(application.reviewer_notes);setConfirmed(false);setError(null);attempt.current=null;}
 async function save(event:FormEvent<HTMLFormElement>) {
  event.preventDefault();if(busy||locked||(published&&(!confirmed||decision===application.published_decision)))return;
  setBusy(true);setError(null);setNotice('');
  const body:Update={decision,notes,revision:application.revision,confirmed:published&&confirmed};
  if(published&&!attempt.current)attempt.current={key:crypto.randomUUID(),body};
  try {
   const result=await adminRequest<{application:AdminProgramApplication}>(`global/programs/${program.slug}/applications/${application.id}/${published?'published-decision':'decision'}`,{method:'POST',headers:{'Content-Type':'application/json',...(published&&attempt.current?{'Idempotency-Key':attempt.current.key}:{})},body:JSON.stringify(published&&attempt.current?attempt.current.body:body)},true);
   setApplication(result.application);setDecision(result.application.decision);setNotes(result.application.reviewer_notes);setEditing(false);setConfirmed(false);setUncertain(false);attempt.current=null;
   setNotice(published?'Updated result published. An update email has been scheduled.':'Private decision saved. Release results when you are ready to share them.');router.refresh();
  } catch(cause) {
   setError(cause instanceof Error?cause:new Error('The result could not be saved. Your notes are still here.'));
   if(published) {
    if(cause instanceof AdminAuthenticationError||(cause instanceof AdminRequestError&&cause.status<500)){attempt.current=null;setUncertain(false);setConfirmed(false);}
    else setUncertain(true);
   }
  } finally {setBusy(false);}
 }
 return <div className="program-admin-review"><article className="program-admin-profile">
  <header><Avatar name={application.full_name} image={application.avatar_url} large /><div><h1>{application.full_name}</h1><p>{application.email}</p><span>{application.city} &middot; Age {application.age}</span></div></header>
  <div className="program-profile-actions"><Link className="app-button app-button-primary" href={`/programs/${program.slug}/communications?ids=${application.id}`}>Send communication</Link><span>Send a personal update, including after results are released.</span></div>
  <section><h2>About the applicant</h2><p className="program-admin-prose">{application.biography||'No introduction added.'}</p>{application.phone&&<p>Phone: {application.phone}</p>}{application.profile_url&&<a className="program-admin-profile-link" href={application.profile_url} target="_blank" rel="noopener noreferrer">Open profile or portfolio (new tab)</a>}</section>
  <section><h2>The event idea</h2><p className="program-admin-prose">{application.proposal}</p></section>{application.question_snapshot.map(q=><section key={q.id}><h2>{q.label}</h2><p className="program-admin-prose">{application.answers[q.id]||'Not answered'}</p></section>)}<p className="program-admin-caption">Application {application.id}</p>
 </article><form className="program-admin-decision" onSubmit={event=>void save(event)}>
  <h2>Your review</h2>{published?<div className="program-admin-review-note"><p>Released result: <strong>{application.published_decision==='rejected'?'Not selected':application.published_decision}</strong></p>{!editing&&application.status==='submitted'&&<button className="app-button app-button-outline" type="button" onClick={()=>{setEditing(true);setNotice('');setConfirmed(false);}}>Edit released result</button>}{editing&&<p>The applicant currently sees the released result. Publishing a change updates their account and sends an update email.</p>}</div>:application.status!=='submitted'?<div className="program-admin-review-note">This application is {application.status} and cannot receive a decision.</div>:<p>Mark the application here. The review team can see it until you release results.</p>}
  <fieldset disabled={busy||locked||uncertain}>
   <RadioGroup label="Decision" name="program-decision" value={decision} onValueChange={value=>{setDecision(value);setConfirmed(false);attempt.current=null;}} layout="stack" options={[...(!published?[{value:'pending' as const,label:'Needs review'}]:[]),{value:'accepted',label:'Accept'},{value:'waitlisted',label:'Waitlist'},{value:'rejected',label:'Not selected'}]} />
   <label className="program-admin-notes" htmlFor="program-review-notes">Private review notes (optional)<textarea id="program-review-notes" maxLength={3000} rows={6} value={notes} onChange={event=>{setNotes(event.target.value);setConfirmed(false);attempt.current=null;}} /></label><p className="program-admin-help">Private notes are never shown to applicants.</p>
  </fieldset>
  {published&&editing&&<label className="program-admin-check"><Checkbox checked={confirmed} disabled={busy||uncertain} onCheckedChange={setConfirmed} />Publish this updated result and email the applicant.</label>}
  {uncertain&&<p role="status">The request could not be confirmed. Retry safely before making another change.</p>}{error&&<p className="program-admin-error" role="alert">{error.message}</p>}<ProgramAuthRecovery error={error} />{notice&&<p className="program-admin-success" role="status">{notice}</p>}
  {(!published||editing)&&<div className="program-review-actions"><button className="app-button app-button-primary" type="submit" disabled={busy||locked||(published&&(!confirmed||decision===application.published_decision))}>{busy?'Saving...':uncertain?'Retry update safely':published?'Publish updated result':'Save private decision'}</button>{published&&<button className="app-button app-button-outline" type="button" disabled={busy||uncertain} onClick={reset}>Cancel</button>}</div>}
 </form></div>;
}
