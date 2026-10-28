'use client';
import { SectionJumpButton } from './SectionJump';
import { useEffect, useRef, useState } from 'react';
export function NoticeImage({src,alt='',width,height,title,href,sectionId,thumbnail=false}:{src:string;alt?:string;width?:number;height?:number;title?:string;href?:string;sectionId?:string;thumbnail?:boolean}){
 const image=useRef<HTMLImageElement>(null),[result,setResult]=useState<{src:string;state:'ready'|'error'}>(),[attempt,setAttempt]=useState(0);
 const state=result?.src===src?result.state:'loading';
 useEffect(()=>{if(image.current?.complete)setResult({src,state:image.current.naturalWidth>0?'ready':'error'});},[src,attempt]);
 // The image link wraps only the image, so its recovery button is never nested
 // in an anchor or another button.
 // eslint-disable-next-line @next/next/no-img-element
 const visual=<img key={`${src}:${attempt}`} ref={image} src={src} alt={alt} title={title} width={width} height={height} loading="lazy" decoding="async" onLoad={()=>setResult({src,state:'ready'})} onError={()=>setResult({src,state:'error'})}/>;
 return <span className={`hw-notice-image${thumbnail?' hw-notice-image-thumbnail':''}`} data-image-state={state} aria-busy={state==='loading'} style={{width:thumbnail?'100%':width||undefined,aspectRatio:thumbnail?'16 / 9':width&&height?`${width} / ${height}`:undefined}}>
  {sectionId?<SectionJumpButton targetId={sectionId}>{visual}</SectionJumpButton>:href?<a href={href} rel="noopener noreferrer">{visual}</a>:visual}
  {state==='error'&&<span className="hw-notice-image-error"><span>This image could not load.</span><button type="button" onClick={event=>{event.preventDefault();event.stopPropagation();setResult(undefined);setAttempt(value=>value+1);}}>Retry image</button></span>}
 </span>;
}
