"use client";
import { useEffect, useRef, useState, useTransition } from 'react';
import { useRouter } from 'next/navigation';
export function CampaignProgress({ active }: { active: boolean }) {
 const router=useRouter(), passes=useRef(0), [pending,startTransition]=useTransition(), [visible,setVisible]=useState(true), [paused,setPaused]=useState(false);
 useEffect(()=>{const update=()=>setVisible(document.visibilityState==='visible');update();document.addEventListener('visibilitychange',update);return()=>document.removeEventListener('visibilitychange',update);},[]);
 useEffect(()=>{
  if(!active || !visible || pending || paused)return;
  if(passes.current>=120){setPaused(true);return;}
  const timer=window.setTimeout(()=>{passes.current++;startTransition(()=>router.refresh());},15000);
  return()=>window.clearTimeout(timer);
 },[active,visible,pending,paused,router]);
 if(!active)return null;
 return <p className="campaign-progress" role="status">{paused?'Automatic updates paused. Reopen this page for current delivery figures.':'Delivery figures update automatically while this page is visible.'}</p>;
}
