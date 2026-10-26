import type { CSSProperties } from 'react';

export type SocialFormat = 'wide' | 'square';
export interface SocialContent {
  title: string;
  description: string;
  label: string;
  detail?: string;
  campaign?: boolean;
}
export const socialSizes = { wide: { width: 1200, height: 630 }, square: { width: 1200, height: 1200 } };
export function socialText(value: string, limit: number) {
  const characters = Array.from(value.replace(/\s+/g, ' ').trim());
  return characters.length > limit ? characters.slice(0, limit - 1).join('') + '…' : characters.join('');
}

/** Presentation only. Entity data and publication checks belong to the Go APIs. */
export function SocialImage({ content, format = 'wide' }: { content: SocialContent; format?: SocialFormat }) {
  const square = format === 'square', campaign = Boolean(content.campaign);
  const title = socialText(content.title, 105);
  const foreground = campaign ? '#f6f8ff' : '#091a40';
  const titleSize = square ? title.length > 65 ? 76 : 98 : title.length > 65 ? 54 : title.length > 42 ? 64 : 78;
  const root: CSSProperties = { width: '100%', height: '100%', display: 'flex', flexDirection: 'column', position: 'relative', overflow: 'hidden', padding: '54px 72px 38px', background: campaign ? '#070f29' : '#f7f9ff', color: foreground, fontFamily: 'SocialBody' };
  return <div style={root}>
    {campaign && <div style={{ position: 'absolute', top: 0, left: 0, ...socialSizes[format], display: 'flex', flexDirection: 'column', overflow: 'hidden', background: '#9cdef9' }}>
      {Array.from({ length: square ? 28 : 17 }, (_, row) => <div key={row} style={{ display: 'flex', whiteSpace: 'pre', height: 24, fontSize: 15, letterSpacing: 4, color: row < 5 ? '#eefbff' : '#55c5f1' }}>{Array.from({ length: 110 }, (_, column) => 'S80#@x+ '[((column * 7 + row * 11) % 19) % 8]).join('')}</div>)}
      <div style={{ position: 'absolute', display: 'flex', left: -300, top: square ? 240 : 140, width: 1800, height: square ? 1800 : 1100, borderRadius: '50%', background: '#070f29' }} />
    </div>}
    <div style={{ display: 'flex', position: 'relative', fontFamily: 'SocialDisplay', fontSize: 35, color: '#091a40' }}>hackways.</div>
    <div style={{ display: 'flex', position: 'relative', flex: 1, flexDirection: 'column', justifyContent: campaign && !square ? 'flex-end' : 'center', paddingTop: 30, paddingBottom: 26 }}>
      <div style={{ display: 'flex', flexDirection: 'column', fontFamily: 'SocialDisplay', fontSize: titleSize, lineHeight: 1.12, letterSpacing: '-0.025em', maxWidth: '100%', overflowWrap: 'anywhere' }}>
        {title === 'Be the reason people meet.' ? <div style={{ display: 'flex', flexDirection: 'column' }}><div style={{ display: 'flex' }}>Be the reason</div><div style={{ display: 'flex', color: campaign ? '#a5d8ff' : '#193d79' }}>people meet.</div></div> : title}
      </div>
      <div style={{ display: 'flex', fontSize: square ? 32 : 27, lineHeight: 1.45, marginTop: 24, maxWidth: 1000 }}>{socialText(content.description, square ? 170 : 130)}</div>
      {content.detail && <div style={{ display: 'flex', fontSize: 24, lineHeight: 1.4, marginTop: 14 }}>{socialText(content.detail, 85)}</div>}
    </div>
    <div style={{ display: 'flex', position: 'relative', justifyContent: 'space-between', alignItems: 'center', borderTop: `1px solid ${campaign ? '#31456b' : '#d2def1'}`, paddingTop: 20, fontSize: 22 }}><span>{socialText(content.label, 70)}</span><span>hackways.in</span></div>
  </div>;
}

export function staticSocialMetadata(origin: string, key: string, title: string, description: string) {
  const images = (['wide', 'square'] as const).map(format => ({ url: new URL(`/social/${key}-${format}.png`, origin).href, ...socialSizes[format], alt: `${title} — Hackways`, type: 'image/png' }));
  return { openGraph: { title, description, siteName: 'Hackways', type: 'website' as const, locale: 'en_IN', images }, twitter: { card: 'summary_large_image' as const, title, description, images: [images[0].url] } };
}


/** The supplied campaign artwork, shared without applicant or result data. */
export function ambassadorSocialMetadata(title: string, description: string) {
  const image = { url: 'https://hackways.in/programs/ambassadors.jpg?v=20261026', width: 3616, height: 2396, alt: 'Hackways Ambassadors', type: 'image/jpeg' };
  return { openGraph: { title, description, siteName: 'Hackways', type: 'website' as const, locale: 'en_IN', images: [image] }, twitter: { card: 'summary_large_image' as const, title, description, images: [image.url] } };
}
