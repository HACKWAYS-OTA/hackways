function origin(value: string | undefined, fallback: string): string {
  const url = new URL(value?.trim() || fallback);
  if (!['http:', 'https:'].includes(url.protocol) || url.username || url.password || url.pathname !== '/' || url.search || url.hash) {
    throw new Error('Site links must be HTTP(S) origins without credentials, paths, queries, or fragments.');
  }
  return url.origin;
}
export const site = {
  app: origin(process.env.NEXT_PUBLIC_APP_URL, 'http://localhost:3001'),
  public: origin(process.env.NEXT_PUBLIC_PUBLIC_URL, 'http://localhost:3000'),
  organizer: origin(process.env.NEXT_PUBLIC_ORGANIZER_URL, 'http://localhost:3002'),
  auth: origin(process.env.NEXT_PUBLIC_AUTH_URL, 'http://localhost:8082'),
};
