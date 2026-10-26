import { staticSocialMetadata, ambassadorSocialMetadata } from '@hackways/ui/social-image';
import { socialPages } from '@hackways/ui/social-pages';
import { site } from './site';
export function pageSocialMetadata(key: keyof typeof socialPages) { const page = socialPages[key]; return key === 'ambassadors' ? ambassadorSocialMetadata(page.title, page.description) : staticSocialMetadata(site.url, key, page.title, page.description); }
