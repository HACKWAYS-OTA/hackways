import 'server-only';
import { redirect } from 'next/navigation';
import { privateApi, requireUser } from '@/lib/account';
export async function mcpData<T>(path: string, destination: string): Promise<{ data?: T; error?: string }> {
  const user = await requireUser(destination);
  if (!user.email_verified) redirect(`/auth/verify-email?next=${encodeURIComponent(destination)}`);
  try { return { data: await privateApi<T>(`/auth/mcp${path}`) }; }
  catch (error) { return { error: error instanceof Error ? error.message : 'The connection request is unavailable. Start again from your app.' }; }
}
