export async function adminVerificationRequired(response: Response): Promise<boolean> {
  if (response.status !== 403) return false;
  const body: unknown = await response.clone().json().catch(() => null);
  return Boolean(body && typeof body === 'object' && 'error' in body && body.error && typeof body.error === 'object' && 'code' in body.error && body.error.code === 'ADMIN_VERIFICATION_REQUIRED');
}
export const adminVerificationPath = (next: string) => `/verify?next=${encodeURIComponent(next)}`;
