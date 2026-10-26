'use client';

export class AdminAuthenticationError extends Error {
  constructor(public verificationRequired: boolean) { super(verificationRequired ? 'Your daily admin verification is needed. Verify in another tab, then retry. Your input is still here.' : 'Your admin session has expired. Sign in in another tab, then retry. Your input is still here.'); this.name = 'AdminAuthenticationError'; }
}
export class AdminRequestError extends Error {
  constructor(message: string, public status: number, public code: string) { super(message); this.name = 'AdminRequestError'; }
}
export async function adminRequest<T>(path: string, options: RequestInit = {}, preserveInputOnAuth = false): Promise<T> {
  const controller = new AbortController();
  const cancel = () => controller.abort();
  options.signal?.addEventListener('abort', cancel, { once: true });
  if (options.signal?.aborted) controller.abort();
  const timeout = setTimeout(cancel, 12000);
  try {
    const response = await fetch(`/api/v1/admin/${path}`, { ...options, credentials: 'include', cache: 'no-store', signal: controller.signal });
    if (response.status === 401) { if (preserveInputOnAuth) throw new AdminAuthenticationError(false); window.location.assign(`/login?next=${encodeURIComponent(window.location.pathname)}`); throw new Error('Sign in again to continue.'); }
    const result: unknown = await response.json();
    if (!response.ok) {
      const error = result && typeof result === 'object' && 'error' in result ? result.error : null;
      if (error && typeof error === 'object' && 'code' in error && error.code === 'ADMIN_VERIFICATION_REQUIRED') { if (preserveInputOnAuth) throw new AdminAuthenticationError(true); window.location.assign(`/verify?next=${encodeURIComponent(window.location.pathname + window.location.search)}`); }
      if (error && typeof error === 'object' && 'code' in error && error.code === 'ADMIN_ACCESS_REQUIRED') window.location.replace('/access-denied');
      throw new AdminRequestError(error && typeof error === 'object' && 'message' in error && typeof error.message === 'string' ? error.message : 'The operation could not be completed. Try again.', response.status, error && typeof error === 'object' && 'code' in error && typeof error.code === 'string' ? error.code : 'REQUEST_FAILED');
    }
    return result as T;
  } finally { clearTimeout(timeout); options.signal?.removeEventListener('abort', cancel); }
}

export function mutationHeaders(key: string) { return { 'Content-Type': 'application/json', 'Idempotency-Key': key }; }
