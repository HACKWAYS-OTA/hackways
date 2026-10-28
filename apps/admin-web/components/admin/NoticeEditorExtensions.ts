import { Extension } from '@tiptap/core';
import ImageExtension from '@tiptap/extension-image';

/** Paragraph shading is deliberately separate from the inline textStyle mark. */
export const NoticeBlockBackground=Extension.create({
 name:'noticeBlockBackground',
 addGlobalAttributes(){return [{types:['paragraph','heading'],attributes:{backgroundColor:{default:null,parseHTML:element=>element.style.backgroundColor||null,renderHTML:attrs=>typeof attrs.backgroundColor==='string'&&/^(?:#[a-f0-9]{3,8}|rgb\([\d, ]+\))$/i.test(attrs.backgroundColor)?{style:`background-color: ${attrs.backgroundColor}`}:{}}}}];}
});

/** Keep the upstream resize handles, but reveal reserved image space while it loads. */
export const NoticeEditorImage=ImageExtension.extend({
 addNodeView(){
  const parent=this.parent?.();if(!parent)return null;
  return props=>{
   const view=parent(props),dom=view.dom as HTMLElement,img=dom instanceof HTMLImageElement?dom:dom.querySelector('img');
   if(!img)return view;
   dom.classList.add('notice-image-view');
   const error=document.createElement('span');error.className='notice-image-error';error.contentEditable='false';
   const text=document.createElement('span');text.textContent='This image could not load.';
   const retry=document.createElement('button');retry.type='button';retry.textContent='Retry image';
   error.append(text,retry);dom.append(error);
   let source=String(props.node.attrs.src||'');
   const state=(value:'loading'|'ready'|'error')=>{dom.dataset.imageState=value;dom.style.visibility='';dom.style.pointerEvents='';dom.setAttribute('aria-busy',String(value==='loading'));};
   const reserve=()=>{const width=Number(props.node.attrs.width),height=Number(props.node.attrs.height);dom.style.setProperty('--notice-image-ratio',width>0&&height>0?`${width} / ${height}`:'16 / 9');};
   const loaded=()=>state(img.naturalWidth>0?'ready':'error'),failed=()=>state('error');
   const reload=(event:MouseEvent)=>{event.preventDefault();event.stopPropagation();state('loading');const next=img.getAttribute('src');if(next){img.removeAttribute('src');img.setAttribute('src',next);}};
   img.addEventListener('load',loaded);img.addEventListener('error',failed);retry.addEventListener('click',reload);
   reserve();if(img.complete)state(img.naturalWidth>0?'ready':'error');else state('loading');
   const update=view.update?.bind(view);
   if(update)view.update=(node,...args)=>{const accepted=update(node,...args);if(accepted){const next=String(node.attrs.src||'');if(next!==source){source=next;state(img.complete?(img.naturalWidth>0?'ready':'error'):'loading');}}return accepted;};
   const destroy=view.destroy?.bind(view);view.destroy=()=>{img.removeEventListener('load',loaded);img.removeEventListener('error',failed);retry.removeEventListener('click',reload);destroy?.();};
   return view;
  };
 }
});
