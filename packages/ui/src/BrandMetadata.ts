/** Shared brand assets; organizer-web mounts its files below /manage. */
export function brandIcons(basePath = '') {
  return {
    icon: [
      { url: `${basePath}/favicon.ico`, sizes: '16x16 32x32 48x48', type: 'image/x-icon' },
      { url: `${basePath}/brand/hackways-icon-96.png`, sizes: '96x96', type: 'image/png' },
      { url: `${basePath}/brand/hackways-icon.svg`, sizes: 'any', type: 'image/svg+xml' },
    ],
    shortcut: `${basePath}/favicon.ico`,
    apple: [{ url: `${basePath}/brand/hackways-icon-180.png`, sizes: '180x180', type: 'image/png' }],
  };
}
