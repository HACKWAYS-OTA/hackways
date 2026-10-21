import { toNextJsHandler } from 'better-auth/next-js';
import { getBetterAuth } from '@/lib/better-auth';

export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';

type HandlerMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';

function disabledResponse(): Response {
  return Response.json(
    { error: { code: 'BETTER_AUTH_DISABLED', message: 'Development Better Auth is disabled.' } },
    { status: 404 },
  );
}

async function dispatch(method: HandlerMethod, request: Request): Promise<Response> {
  const auth = getBetterAuth();
  if (!auth) return disabledResponse();
  const handler = toNextJsHandler(auth)[method];
  return handler(request);
}

export function GET(request: Request) { return dispatch('GET', request); }
export function POST(request: Request) { return dispatch('POST', request); }
export function PATCH(request: Request) { return dispatch('PATCH', request); }
export function PUT(request: Request) { return dispatch('PUT', request); }
export function DELETE(request: Request) { return dispatch('DELETE', request); }
