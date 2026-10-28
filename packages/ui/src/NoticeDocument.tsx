import { NoticeImage as LoadedNoticeImage } from './NoticeImage';
import { Fragment, type CSSProperties, type ReactNode } from 'react';
export type NoticeDocumentNode = { type:string; text?:string; attrs?:{ [key:string]:string|number|boolean|null|number[] }; marks?:{type:string;attrs?:NoticeDocumentNode['attrs']}[]; content?:NoticeDocumentNode[] };
export type NoticeImage = {url:string;width:number;height:number};
const fonts=new Set(['Google Sans Flex','FG Futurist','Arial','Helvetica','Georgia','Times New Roman','Courier New','Verdana','Tahoma','Trebuchet MS','monospace','sans-serif','serif']);
function styles(attrs:NoticeDocumentNode['attrs']={}):CSSProperties {
 const style:CSSProperties={};
 if(!attrs.textAlign&&attrs.align)attrs={...attrs,textAlign:attrs.align};
 if(typeof attrs.textAlign==='string'&&['left','right','center','justify'].includes(attrs.textAlign))style.textAlign=attrs.textAlign as CSSProperties['textAlign'];
 for(const key of ['color','backgroundColor'] as const)if(typeof attrs[key]==='string'&&/^#(?:[a-f0-9]{3}|[a-f0-9]{6}|[a-f0-9]{8})$/i.test(attrs[key]))style[key]=attrs[key];
 if(typeof attrs.fontFamily==='string'&&fonts.has(attrs.fontFamily))style.fontFamily=attrs.fontFamily;
 if(typeof attrs.fontSize==='string'&&/^(?:[89]|[1-6][0-9]|7[0-2])(?:px|pt)$/.test(attrs.fontSize))style.fontSize=attrs.fontSize;
 if(typeof attrs.lineHeight==='string'&&Number(attrs.lineHeight)>=1&&Number(attrs.lineHeight)<=3)style.lineHeight=attrs.lineHeight;
 return style;
}
function safeURL(value:unknown,image=false):string|undefined {
 if(typeof value!=='string'||/[\u0000-\u0020]/.test(value)||value.startsWith('//'))return;
 if(image?!/^(?:https:\/\/|\/(?!\/))/.test(value):!/^(?:https?:\/\/|mailto:|tel:|\/(?!\/))/.test(value))return;
 if(image&&/[?&]sig=/i.test(value))return;
 return value.split('#')[0];
}
/** Render the validated document with React elements; arbitrary HTML and CSS never enter the page. */
export function NoticeDocument({document,images=[]}:{document:NoticeDocumentNode;images?:NoticeImage[]}) {
 function render(node:NoticeDocumentNode,key:number,depth=0):ReactNode {
  if(depth>32)return null;
  const a=node.attrs||{},children=node.content?.map((child,index)=>render(child,index,depth+1));
  if(node.type==='text'){
   let text:ReactNode=node.text||'';
   for(const mark of node.marks||[]){switch(mark.type){case 'bold':text=<strong>{text}</strong>;break;case 'italic':text=<em>{text}</em>;break;case 'underline':text=<u>{text}</u>;break;case 'strike':text=<s>{text}</s>;break;case 'code':text=<code>{text}</code>;break;case 'textStyle':text=<span style={styles(mark.attrs)}>{text}</span>;break;case 'link':{const href=safeURL(mark.attrs?.href);if(href)text=<a href={href} rel="noopener noreferrer">{text}</a>;break;}}}
   return <Fragment key={key}>{text}</Fragment>;
  }
  switch(node.type){
   case 'doc':return <Fragment key={key}>{children}</Fragment>;
   case 'paragraph':return <p key={key} style={styles(a)}>{children||<br/>}</p>;
   case 'heading':{const Tag=Number(a.level)<=2?'h2':Number(a.level)===3?'h3':Number(a.level)===4?'h4':Number(a.level)===5?'h5':'h6';return <Tag key={key} style={styles(a)}>{children}</Tag>;}
   case 'hardBreak':return <br key={key}/>;
   case 'horizontalRule':return <hr key={key}/>;
   case 'blockquote':return <blockquote key={key}>{children}</blockquote>;
   case 'bulletList':return <ul key={key}>{children}</ul>;
   case 'orderedList':return <ol key={key} type={['1','a','A','i','I'].includes(String(a.type))?a.type as '1'|'a'|'A'|'i'|'I':undefined} start={typeof a.start==='number'?a.start:1}>{children}</ol>;
   case 'listItem':return <li key={key}>{children}</li>;
   case 'codeBlock':return <pre key={key}><code>{children}</code></pre>;
   case 'table':return <div key={key} className="hw-notice-table" tabIndex={0} role="region" aria-label="Notice table"><table><tbody>{children}</tbody></table></div>;
   case 'tableRow':return <tr key={key}>{children}</tr>;
   case 'tableCell':case 'tableHeader':{const Tag=node.type==='tableHeader'?'th':'td';return <Tag key={key} style={styles(a)} colSpan={typeof a.colspan==='number'?a.colspan:1} rowSpan={typeof a.rowspan==='number'?a.rowspan:1}>{children}</Tag>;}
   case 'image':{const src=safeURL(a.src,true),size=images.find(image=>image.url===src);return src?<LoadedNoticeImage key={key} src={src} alt={typeof a.alt==='string'?a.alt:''} width={typeof a.width==='number'?a.width:size?.width} height={typeof a.height==='number'?a.height:size?.height}/>:null;}
   default:return null;
  }
 }
 return <div className="hw-notice-markdown hw-notice-document">{render(document,0)}</div>;
}
