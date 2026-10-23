function environmentOrigin(name: string, configured: string | undefined, localOrigin: string, allowPath = false): string {
  const url = new URL(configured?.trim() || localOrigin);
  if (!['http:', 'https:'].includes(url.protocol) || url.username || url.password || (!allowPath && url.pathname !== '/') || url.search || url.hash) {
    throw new Error(`${name} must be an absolute HTTP or HTTPS origin without credentials, a path, query, or fragment.`);
  }
  return allowPath ? url.origin + (url.pathname === '/' ? '' : url.pathname.replace(/\/$/, '')) : url.origin;
}

export const site = {
  name: "Hackways",
  url: environmentOrigin("NEXT_PUBLIC_SITE_URL", process.env.NEXT_PUBLIC_SITE_URL, "http://localhost:3000"),
  app: environmentOrigin("NEXT_PUBLIC_APP_URL", process.env.NEXT_PUBLIC_APP_URL, "http://localhost:3001"),
  organizer: environmentOrigin("NEXT_PUBLIC_ORGANIZER_URL", process.env.NEXT_PUBLIC_ORGANIZER_URL, "http://localhost:3002", true),
};

const configuredEmail = process.env.NEXT_PUBLIC_SUPPORT_EMAIL?.trim().toLowerCase();
export const supportEmail = configuredEmail && /^[^\s@<>]+@[^\s@<>]+\.[^\s@<>]+$/.test(configuredEmail)
  ? configuredEmail
  : 'support@hackota.com';
