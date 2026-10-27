import type { NoticeDocumentNode } from '@hackways/ui/document';
import 'server-only';
import { cache } from 'react';
import { notFound } from 'next/navigation';
import { siteContentApi } from '@hackways/ui/seo';
export type PublicNotice = { slug: string; title: string; summary: string; body: string; document:NoticeDocumentNode|null; body_format:string; published_at: string; updated_at: string; images: {url:string;width:number;height:number}[] };
export const readNotice = cache(async (slug: string): Promise<PublicNotice> => {
 if (!/^[a-z0-9][a-z0-9-]{0,119}$/.test(slug)) notFound();
 const response = await fetch(`${siteContentApi()}/notices/${encodeURIComponent(slug)}`, { cache:'no-store', credentials:'omit', signal:AbortSignal.timeout(6000) });
 if (response.status === 404) notFound();
 if (!response.ok) throw new Error('The notice could not be loaded. Please try again.');
 const result=await response.json() as {data:Omit<PublicNotice,'images'>;images:PublicNotice['images']};return {...result.data,images:result.images};
});
export async function noticeList(offset: number) {
 const response = await fetch(`${siteContentApi()}/notices?limit=20&offset=${offset}`, { cache:'no-store', credentials:'omit', signal:AbortSignal.timeout(6000) });
 if (!response.ok) throw new Error('Notices could not be loaded. Please try again.');
 return await response.json() as { data: Omit<PublicNotice,'body'|'images'>[]; has_next:boolean };
}
