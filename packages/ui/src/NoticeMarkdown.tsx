import ReactMarkdown, { defaultUrlTransform } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Checkbox } from './Checkbox';
import { SectionJumpButton } from './SectionJump';
import { Children, isValidElement, type ReactNode } from 'react';

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
  a: ({ href, children }) => {
   if (!href) return <span>{children}</span>;
   if (href.startsWith('#')) { let section = ''; try { section = decodeURIComponent(href.slice(1)); } catch { return <span>{children}</span>; } return <SectionJumpButton targetId={`notice-${sectionName(section)}`}>{children}</SectionJumpButton>; }
   const clean = href.split('#')[0]; return <a href={clean} rel="noopener noreferrer">{children}</a>;
  },
  // Uploaded images have intrinsic file dimensions unknown until browser decode.
  // eslint-disable-next-line @next/next/no-img-element
  img: ({ src, alt, title }) => { const size=images.find(image=>image.url===src); return src ? <img src={src} alt={alt || ''} title={title} width={size?.width} height={size?.height} loading="lazy" decoding="async" /> : <span>{alt}</span>; },
  input: ({ checked }) => <Checkbox checked={Boolean(checked)} disabled aria-label={checked ? 'Completed task' : 'Incomplete task'} />,
  table: ({ children }) => <div className="hw-notice-table" tabIndex={0} role="region" aria-label="Notice table"><table>{children}</table></div>,
 }}>{children}</ReactMarkdown></div>;
}
