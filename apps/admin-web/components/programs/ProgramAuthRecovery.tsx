'use client';
import { AdminAuthenticationError } from '@/lib/admin-request';
export function ProgramAuthRecovery({ error }: { error: Error | null }) {
 if (!(error instanceof AdminAuthenticationError)) return null;
 return <p className="program-admin-auth-recovery"><a href={error.verificationRequired ? '/verify?next=%2Fprograms' : '/login?next=%2Fprograms'} target="_blank" rel="noopener noreferrer">{error.verificationRequired ? 'Verify admin access in a new tab' : 'Sign in to admin in a new tab'}</a>. Return here and retry your save.</p>;
}
