import Link from 'next/link';
import { managedMetadata } from '@hackways/ui/seo';
import { noticeList } from '@/lib/notices';
import '../notice/notice.css';
export function generateMetadata() { return managedMetadata('/notices',{ title:'Notices',description:'Official updates and notices from Hackways.',alternates:{canonical:'/notices'} }); }
export default async function Notices({ searchParams }: { searchParams:{ offset?:string } }) {
 const raw = Number(searchParams.offset), offset = Number.isSafeInteger(raw) && raw > 0 && raw <= 100000 ? raw : 0;
 const result = await noticeList(offset);
 return <div className="notice-page"><header className="notice-page-header"><h1>Notices</h1><p>Official updates from Hackways.</p></header><div className="notice-list">{result.data.map(notice => <article key={notice.slug}><Link href={`/notice/${notice.slug}`}><h2>{notice.title}</h2><p>{notice.summary}</p><time dateTime={notice.published_at}>{new Intl.DateTimeFormat('en-IN',{dateStyle:'long',timeZone:'Asia/Kolkata'}).format(new Date(notice.published_at))}</time></Link></article>)}</div>{!result.data.length && <p className="notice-empty">No notices have been published yet.</p>}<nav className="notice-pagination" aria-label="Notice pages">{offset > 0 && <Link href={`/notices?offset=${offset-20}`}>Previous</Link>}{result.has_next && <Link href={`/notices?offset=${offset+20}`}>Next</Link>}</nav></div>;
}
