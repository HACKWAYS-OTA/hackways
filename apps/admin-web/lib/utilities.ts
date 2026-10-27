import type { NoticeDocumentNode } from '@hackways/ui/document';
export type Notice = { id: string; reference: number; revision: number; slug: string; title: string; summary: string; body: string; document?:NoticeDocumentNode|null; body_format?:'markdown'|'richtext'; published_at?:string|null; has_published?:boolean; status: 'draft' | 'published' | 'archived' };
export type SeoPage = { reference: number; revision: number; path: string; title: string; description: string; canonical_path: string; noindex: boolean; planning_keywords: string[]; notes: string };
export type NoticeListItem = Omit<Notice, 'body'>;
export type SeoListItem = Pick<SeoPage, 'path' | 'title' | 'description' | 'noindex' | 'revision'> & { keyword_count: number };
export type SeoTopic = { topic_key:string; theme:string; title:string; keywords:string[] };
