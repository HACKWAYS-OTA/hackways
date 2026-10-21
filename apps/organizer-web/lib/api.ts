import type { ApiError } from '@/types/api';
function apiBase(path: string): string {
  const global = path.startsWith('/auth/') || path.startsWith('/storage/');
  const configured = global ? process.env.NEXT_PUBLIC_GLOBAL_API_URL : process.env.NEXT_PUBLIC_API_URL;
  if (typeof window !== 'undefined') return configured || (process.env.NODE_ENV === 'production' ? '/api/v1' : `http://localhost:${global ? 8082 : 8081}/api/v1`);
  const internal = global ? process.env.GLOBAL_API_URL : process.env.CORE_API_URL;
  return internal || (configured?.startsWith('http') ? configured : process.env.NODE_ENV === 'production' ? `http://${global ? 'global-backend:8082' : 'core-backend:8081'}/api/v1` : `http://localhost:${global ? 8082 : 8081}/api/v1`);
}
export class ApiClientError extends Error {
  constructor(message: string, public code: string, public details?: Record<string, unknown>) { super(message); this.name = 'ApiClientError'; }
}
export async function fetchApi<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set('Accept', 'application/json');
  if (options.body && !headers.has('Content-Type') && !(options.body instanceof FormData)) headers.set('Content-Type', 'application/json');
  const controller = new AbortController();
  const abort = () => controller.abort();
  if (options.signal?.aborted) controller.abort();
  options.signal?.addEventListener('abort', abort, { once: true });
  const timeout = setTimeout(() => controller.abort(), 12000);
  try {
    const response = await fetch(`${apiBase(path).replace(/\/$/, '')}${path.startsWith('/') ? path : `/${path}`}`, { ...options, headers, signal: controller.signal, credentials: 'include' });
    if (!response.ok) {
      let envelope: ApiError | null = null;
      try { envelope = await response.json() as ApiError; } catch { /* Non-JSON error responses use the HTTP status. */ }
      throw new ApiClientError(envelope?.error?.message || 'The request could not be completed. Please try again.', envelope?.error?.code || `HTTP_${response.status}`, envelope?.error?.details);
    }
    if (response.status === 204) return {} as T;
    return await response.json() as T;
  } catch (error) {
    if (options.signal?.aborted) throw error;
    if (controller.signal.aborted) throw new ApiClientError('The event service is taking longer than expected. Please try again.', 'REQUEST_TIMEOUT');
    if (error instanceof TypeError) throw new ApiClientError("We couldn't reach the event service. Please try again in a moment.", 'NETWORK_ERROR');
    if (error instanceof SyntaxError) throw new ApiClientError('The event service returned an unreadable response. Please try again.', 'INVALID_RESPONSE');
    throw error;
  } finally { clearTimeout(timeout); options.signal?.removeEventListener('abort', abort); }
}
