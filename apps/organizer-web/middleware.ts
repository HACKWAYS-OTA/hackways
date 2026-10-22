import { NextRequest, NextResponse } from 'next/server';
import { requestRolloutCookie } from '@hackways/api-client';
export async function middleware(request: NextRequest) {
  if (request.cookies.has('__Host-hackways_rollout')) return NextResponse.next();
  const base = process.env.GLOBAL_API_URL || (process.env.NODE_ENV === 'production' ? 'http://global-backend:8082/api/v1' : 'http://localhost:8082/api/v1');
  const cookie = await requestRolloutCookie(base);
  if (!cookie) return NextResponse.next();
  const headers = new Headers(request.headers);
  headers.set('cookie', [headers.get('cookie'), cookie.split(';')[0]].filter(Boolean).join('; '));
  const response = NextResponse.next({ request: { headers } });
  response.headers.append('set-cookie', cookie);
  return response;
}
export const config = { matcher: ['/events/:path*'] };
