function origin(value: string | undefined, fallback: string, allowPath = false): string {
  const url = new URL(value?.trim() || fallback);
  if (!['http:', 'https:'].includes(url.protocol) || url.username || url.password || (!allowPath && url.pathname !== '/') || url.search || url.hash) throw new Error('Site links must be HTTP(S) origins without credentials, paths, queries, or fragments.');
  return allowPath ? url.origin + (url.pathname === '/' ? '' : url.pathname.replace(/\/$/, '')) : url.origin;
}
export const site = { app: origin(process.env.NEXT_PUBLIC_APP_URL, 'http://localhost:3001'), organizer: origin(process.env.NEXT_PUBLIC_ORGANIZER_URL, 'http://localhost:3002', true) };
