"use client";
import { useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { fetchApi } from '@/lib/api';
import './undo-check-in.css';

export function UndoCheckInButton({ eventId, ticketId, checkedInAt, name, disabled=false, onUndone, onBusy }: {
 eventId:string; ticketId:string; checkedInAt:string; name:string; disabled?:boolean; onUndone?:()=>void; onBusy?:(busy:boolean)=>void;
}) {
 const router=useRouter(), lock=useRef(false);
 const [busy,setBusy]=useState(false),[error,setError]=useState(''),[done,setDone]=useState(false);
 async function undo(){
  if(lock.current || done)return;lock.current=true;setBusy(true);onBusy?.(true);setError('');
  try {await fetchApi(`/events/${eventId}/check-in/undo`,{method:'POST',body:JSON.stringify({ticket_id:ticketId,checked_in_at:checkedInAt})});setDone(true);onUndone?.();router.refresh();}
  catch(failure){setError(failure instanceof Error?failure.message:'Check-in could not be undone. Try again.');}
  finally{lock.current=false;setBusy(false);onBusy?.(false);}
 }
 return <span className="checkin-undo-action">{done?<span role="status">Check-in undone</span>:<button type="button" className="app-button app-button-outline app-button-sm" disabled={disabled || busy} aria-label={`Undo check-in for ${name}`} onClick={undo}>{busy?'Undoing...':'Undo check-in'}</button>}{error && <span className="journey-error" role="alert">{error}</span>}</span>;
}
