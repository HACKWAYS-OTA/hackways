'use client';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
/** Re-read the server list once when returning from a separate editor tab. */
export function NoticeListSync(){
 const router=useRouter();
 useEffect(()=>{let hidden=document.visibilityState==='hidden';const visibility=()=>{if(document.visibilityState==='hidden'){hidden=true;return;}if(hidden){hidden=false;router.refresh();}};document.addEventListener('visibilitychange',visibility);return()=>document.removeEventListener('visibilitychange',visibility);},[router]);
 return null;
}
