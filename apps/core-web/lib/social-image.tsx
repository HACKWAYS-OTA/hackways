import 'server-only';
import { ImageResponse } from 'next/og';
import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { createHash } from 'node:crypto';
import { SocialImage, socialSizes, type SocialContent, type SocialFormat } from '@hackways/ui/social-image';

// Bounded, disposable cache of rendered public image bytes, never business records.
const cache = new Map<string, { bytes: Uint8Array; expires: number }>();
let cachedBytes = 0;
let active: { key: string; promise: Promise<Uint8Array> } | null = null;
let waiting = 0;
let fonts: Promise<{ name: string; data: ArrayBuffer; weight: 400; style: 'normal' }[]> | undefined;
function imageFonts() {
  if (!fonts) fonts = Promise.all([['SocialDisplay', 'social-display.ttf'], ['SocialBody', 'social-body.ttf']].map(async ([name, file]) => ({ name, data: new Uint8Array(await readFile(join(process.cwd(), 'public', 'fonts', file))).buffer, weight: 400 as const, style: 'normal' as const }))).catch(error => { fonts = undefined; throw error; });
  return fonts;
}
function remember(key: string, bytes: Uint8Array) {
  if (bytes.byteLength > 2 * 1024 * 1024) return;
  while (cache.size >= 12 || cachedBytes + bytes.byteLength > 8 * 1024 * 1024) {
    const oldest = cache.keys().next().value;
    if (oldest === undefined) break;
    cachedBytes -= cache.get(oldest)!.bytes.byteLength; cache.delete(oldest);
  }
  cache.set(key, { bytes, expires: Date.now() + 10 * 60 * 1000 }); cachedBytes += bytes.byteLength;
}
const unavailable = () => new Response('Image generation is busy. Please retry.', { status: 503, headers: { 'Retry-After': '3', 'Cache-Control': 'no-store' } });
const imageResponse = (bytes: Uint8Array, key: string) => new Response(new Uint8Array(bytes), { headers: { 'Content-Type': 'image/png', 'Cache-Control': 'public, max-age=0, must-revalidate', 'X-Content-Type-Options': 'nosniff', ETag: `"${key}"` } });

export function socialImageMetadata(title: string) {
  return (['wide', 'square'] as const).map(id => ({ id, size: socialSizes[id], alt: `${title} — Hackways`, contentType: 'image/png' }));
}

export async function renderSocialImage(content: SocialContent, id: string) {
  if (id !== 'wide' && id !== 'square') return new Response('Image not found', { status: 404 });
  const format: SocialFormat = id;
  const key = createHash('sha256').update(JSON.stringify([content, format, 'social-v1'])).digest('hex');
  const hit = cache.get(key);
  if (hit && hit.expires > Date.now()) return imageResponse(hit.bytes, key);
  if (hit) { cachedBytes -= hit.bytes.byteLength; cache.delete(key); }
  if (active) {
    if (active.key === key) { try { return imageResponse(await active.promise, key); } catch { return unavailable(); } }
    if (waiting >= 2) return unavailable();
    waiting++;
    let timeout: ReturnType<typeof setTimeout> | undefined;
    try { await Promise.race([active.promise.catch(() => undefined), new Promise(resolve => { timeout = setTimeout(resolve, 3000); })]); }
    finally { waiting--; if (timeout) clearTimeout(timeout); }
    if (active) return unavailable();
    const saved = cache.get(key);
    if (saved && saved.expires > Date.now()) return imageResponse(saved.bytes, key);
  }
  const promise = (async () => {
    const response = new ImageResponse(<SocialImage content={content} format={format} />, { ...socialSizes[format], fonts: await imageFonts() });
    const bytes = new Uint8Array(await response.arrayBuffer());
    remember(key, bytes); return bytes;
  })();
  active = { key, promise };
  try { return imageResponse(await promise, key); }
  catch (error) { console.error('Public social image could not render', { kind: error instanceof Error ? error.name : 'unknown' }); return unavailable(); }
  finally { if (active?.key === key) active = null; }
}
