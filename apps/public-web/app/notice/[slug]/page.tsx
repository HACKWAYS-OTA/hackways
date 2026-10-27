import { managedMetadata } from '@hackways/ui/seo';
import { NoticeDocument } from '@hackways/ui/document';
import { NoticeMarkdown } from '@hackways/ui/markdown';
import '@hackways/ui/markdown.css';
import { readNotice } from '@/lib/notices';
import { site } from '@/lib/site';
import '../notice.css';
export async function generateMetadata({ params }: { params: { slug: string } }) {
 const notice = await readNotice(params.slug), path = `/notice/${notice.slug}`;
 return managedMetadata(path, { title:notice.title, description:notice.summary, alternates:{ canonical:path }, openGraph:{ type:'article', title:notice.title, description:notice.summary, url:path, publishedTime:notice.published_at, modifiedTime:notice.updated_at }, twitter:{ card:'summary', title:notice.title, description:notice.summary } });
}
export default async function NoticePage({ params }: { params: { slug: string } }) {
 const notice = await readNotice(params.slug), url = `${site.url}/notice/${notice.slug}`;
 const structured = { '@context':'https://schema.org', '@type':'Article', headline:notice.title, description:notice.summary, datePublished:notice.published_at, dateModified:notice.updated_at, mainEntityOfPage:url, author:{ '@type':'Organization',name:'Hackways',url:site.url }, publisher:{ '@type':'Organization',name:'Hackways',url:site.url,logo:{ '@type':'ImageObject',url:`${site.url}/brand/hackways-icon-512.png` } } };
 return <article className="notice-page"><script type="application/ld+json" dangerouslySetInnerHTML={{ __html:JSON.stringify(structured).replace(/</g,'\\u003c') }} /><header className="notice-page-header"><h1>{notice.title}</h1><p>{notice.summary}</p><time dateTime={notice.published_at}>{new Intl.DateTimeFormat('en-IN',{ dateStyle:'long',timeZone:'Asia/Kolkata' }).format(new Date(notice.published_at))}</time></header>{notice.body_format==='richtext'&&notice.document?<NoticeDocument document={notice.document} images={notice.images}/>:<NoticeMarkdown images={notice.images}>{notice.body}</NoticeMarkdown>}</article>;
}
