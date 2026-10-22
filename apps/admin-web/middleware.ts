import { NextResponse, type NextRequest } from 'next/server';
import { verifyAdminAccess } from './lib/admin-session';

// Reject protected requests before Next renders any dashboard or RSC payload.
// The shared server layout independently rechecks access as defense in depth.
export async function middleware(request: NextRequest) {
  const requestHeaders = new Headers(request.headers);
  requestHeaders.set('x-hackways-admin-path', `${request.nextUrl.pathname}${request.nextUrl.search}`);
  if (!['/login', '/verify', '/access-denied', '/auth-unavailable'].includes(request.nextUrl.pathname)) {
    const sessionCookies = ['__Host-hackways_admin_session', '__Host-hackways_session']
      .map(name => request.cookies.get(name))
      .filter((cookie): cookie is { name: string; value: string } => Boolean(cookie?.value));
    const access = await verifyAdminAccess(sessionCookies);
    if (access !== 'allowed') {
      const destination = new URL(access === 'verification-required' ? '/verify' : access === 'signed-out' ? '/login' : access === 'denied' ? '/access-denied' : '/auth-unavailable', request.url);
      if (access === 'signed-out' || access === 'verification-required') destination.searchParams.set('next', `${request.nextUrl.pathname}${request.nextUrl.search}`);
      const blocked = NextResponse.redirect(destination);
      blocked.headers.set('Cache-Control', 'private, no-store');
      return blocked;
    }
  }
  const response = NextResponse.next({
    request: { headers: requestHeaders },
    ...(request.nextUrl.pathname === '/access-denied' ? { status: 403 }
      : request.nextUrl.pathname === '/auth-unavailable' ? { status: 503 } : {}),
  });
  response.headers.set('Cache-Control', 'private, no-store');
  if (request.nextUrl.pathname === '/auth-unavailable') response.headers.set('Retry-After', '30');
  return response;
}

export const config = {
  matcher: ['/((?!api(?:/|$)|_next(?:/|$)|fonts(?:/|$)|favicon.ico$).*)'],
};
