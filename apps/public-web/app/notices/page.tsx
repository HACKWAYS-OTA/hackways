import Link from 'next/link';
import { managedMetadata } from '@hackways/ui/seo';
import { NoticeImage } from '@hackways/ui/notice-image';
import { HugeiconsIcon } from '@hugeicons/react';
import { ArrowUpRight01Icon } from '@hugeicons/core-free-icons';
import '@hackways/ui/markdown.css';
import { noticeList } from '@/lib/notices';
import '../notice/notice.css';
export function generateMetadata(){return managedMetadata('/notices',{title:'Notices',description:'Official updates and notices from Hackways.',alternates:{canonical:'/notices'}});}
export default async function Notices({searchParams}:{searchParams:{offset?:string}}){
 const raw=Number(searchParams.offset),offset=Number.isSafeInteger(raw)&&raw>0&&raw<=100000?raw:0,result=await noticeList(offset);
 return <div className="notice-directory"><header className="notice-directory-header"><h1>Notices</h1><p>Official updates from Hackways.</p></header><div className="notice-news-grid">{result.data.map(notice=><article key={notice.slug} className="notice-news-card">{notice.cover_url&&<div className="notice-news-media"><NoticeImage src={notice.cover_url} alt="" width={notice.cover_width} height={notice.cover_height} thumbnail/></div>}<div className="notice-news-copy"><time dateTime={notice.published_at}>{new Intl.DateTimeFormat('en-IN',{dateStyle:'long',timeZone:'Asia/Kolkata'}).format(new Date(notice.published_at))}</time><h2><Link href={`/notice/${notice.slug}`} className="notice-news-link">{notice.title}</Link></h2><p>{notice.summary}</p><HugeiconsIcon icon={ArrowUpRight01Icon} size={21} aria-hidden="true"/></div></article>)}</div>{!result.data.length&&<p className="notice-empty">No notices have been published yet.</p>}<nav className="notice-pagination" aria-label="Notice pages">{offset>0&&<Link href={`/notices?offset=${offset-20}`}>Previous</Link>}{result.has_next&&<Link href={`/notices?offset=${offset+20}`}>Next</Link>}</nav></div>;
}
