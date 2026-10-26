/** Exact supplied artwork. Only the format/size differs between these assets. */
export function brandIcons(basePath = '', variant: 'general' | 'event' | 'admin-auth' = 'general') {
  const path = `${basePath}/brand/${variant}`;
  return {
    icon: [
      { url: `${path}/favicon.ico`, sizes: '16x16 32x32 48x48 64x64 128x128 256x256', type: 'image/x-icon' },
      { url: `${path}/icon-96.png`, sizes: '96x96', type: 'image/png' },
      { url: `${path}/favicon.svg`, sizes: 'any', type: 'image/svg+xml' },
    ],
    shortcut: `${path}/favicon.ico`,
    apple: [{ url: `${path}/icon-180.png`, sizes: '180x180', type: 'image/png' }],
  };
}
