import { staticSocialMetadata, socialSizes } from '@hackways/ui/social-image';
import { socialPages } from '@hackways/ui/social-pages';
import { site } from './site';
export function pageSocialMetadata(key: keyof typeof socialPages) { const page = socialPages[key]; return staticSocialMetadata(site.public, key, page.title, page.description); }
export function entitySocialMetadata(path: string, title: string, description: string) {
  const images = (['wide', 'square'] as const).map(format => ({ url: new URL(`${path}/opengraph-image/${format}`, site.app).href, ...socialSizes[format], alt: `${title} - Hackways`, type: 'image/png' }));
  return { openGraph: { title, description, siteName: 'Hackways', type: 'website' as const, url: path, images }, twitter: { card: 'summary_large_image' as const, title, description, images: [images[0].url] } };
}
