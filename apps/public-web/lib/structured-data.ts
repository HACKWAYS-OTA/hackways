import { site } from './site';

/** Public identity only. Dynamic records and private account data stay in Go. */
export const siteIdentity = {
  '@context': 'https://schema.org',
  '@graph': [
    {
      '@type': 'Organization',
      '@id': `${site.url}/#organization`,
      name: 'Hackways',
      url: site.url,
      description: 'Hackways is an event and community platform for discovering technology gatherings, hosting events and managing registrations.',
      logo: {
        '@type': 'ImageObject',
        url: `${site.url}/brand/hackways-icon-512.png`,
        width: 512,
        height: 512,
      },
    },
    {
      '@type': 'WebSite',
      '@id': `${site.url}/#website`,
      name: 'Hackways',
      url: site.url,
      inLanguage: 'en',
      publisher: { '@id': `${site.url}/#organization` },
    },
  ],
};
