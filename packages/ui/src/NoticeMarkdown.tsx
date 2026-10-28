import { NoticeImage } from './NoticeImage';
import ReactMarkdown, { defaultUrlTransform } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Checkbox } from './Checkbox';
import { SectionJumpButton } from './SectionJump';
import { Children, cloneElement, isValidElement, type ReactNode } from 'react';

function plain(children: ReactNode): string {
  return Children.toArray(children).map(child => typeof child === 'string' || typeof child === 'number' ? String(child) : isValidElement<{ children?: ReactNode }>(child) ? plain(child.props.children) : '').join('');
}
function sectionName(value: string) { return value.trim().toLowerCase().replace(new RegExp('[^\\p{L}\\p{N}\\s-]', 'gu'), '').replace(/\s+/g, '-').slice(0, 120); }

/** Shared public rendering and editor preview. HTML and executable URLs stay disabled. */
export function NoticeMarkdown({ children, images = [] }: { children: string; images?: {url:string;width:number;height:number}[] }) {
 const headings = new Map<string, number>();
 const headingId = (value: ReactNode) => { const base = sectionName(plain(value)) || 'section', count = headings.get(base) || 0; headings.set(base, count + 1); return `notice-${base}${count ? `-${count}` : ''}`; };
 return <div className="hw-notice-markdown"><ReactMarkdown remarkPlugins={[remarkGfm]} skipHtml urlTransform={(url, key) => {
  const safe = defaultUrlTransform(url);
  if (key === 'src') return /^(?:https:\/\/|\/(?!\/))/.test(safe) && !/[?&](?:sig|se|sp)=/i.test(safe) ? safe : '';
  return safe;
 }} components={{
  h1: ({ children }) => <h2 id={headingId(children)} tabIndex={-1}>{children}</h2>,
  h2: ({ children }) => <h2 id={headingId(children)} tabIndex={-1}>{children}</h2>,
  h3: ({ children }) => <h3 id={headingId(children)} tabIndex={-1}>{children}</h3>,
  h4: ({ children }) => <h4 id={headingId(children)} tabIndex={-1}>{children}</h4>,
  a: ({ href, children, node }) => {
   if (!href) return <span>{children}</span>;
   let targetId:string|undefined;
   if(href.startsWith('#')){try{targetId=`notice-${sectionName(decodeURIComponent(href.slice(1)))}`;}catch{return <span>{children}</span>;}}
   const clean=href.split('#')[0];
   const hasImage=(item:typeof node):boolean=>Boolean(item?.children?.some(child=>child.type==='element'&&(child.tagName==='img'||hasImage(child))));
   const link=(content:ReactNode)=>targetId?<SectionJumpButton targetId={targetId}>{content}</SectionJumpButton>:<a href={clean} rel="noopener noreferrer">{content}</a>;
   if(hasImage(node)){
    type ImageChild={children?:ReactNode;src?:string;alt?:string;title?:string;node?:{tagName?:string}};
    const linked=(content:ReactNode):ReactNode=>Children.map(content,child=>{
     if(isValidElement<ImageChild>(child)){
      if(child.props.node?.tagName==='img'&&child.props.src){const size=images.find(image=>image.url===child.props.src);return <NoticeImage src={child.props.src} alt={child.props.alt} title={child.props.title} width={size?.width} height={size?.height} href={targetId?undefined:clean} sectionId={targetId}/>;}
      if(child.props.children)return cloneElement(child,{children:linked(child.props.children)});
     }
     return link(child);
    });
    return <span className="hw-notice-linked-media">{linked(children)}</span>;
   }
   return link(children);
  },
  img: ({ src, alt, title }) => { const size=images.find(image=>image.url===src); return src ? <NoticeImage src={src} alt={alt || ''} title={title} width={size?.width} height={size?.height} /> : <span>{alt}</span>; },
  input: ({ checked }) => <Checkbox checked={Boolean(checked)} disabled aria-label={checked ? 'Completed task' : 'Incomplete task'} />,
  table: ({ children }) => <div className="hw-notice-table" tabIndex={0} role="region" aria-label="Notice table"><table>{children}</table></div>,
 }}>{children}</ReactMarkdown></div>;
}
