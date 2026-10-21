'use client';
import { useTransition } from 'react';
import { useRouter } from 'next/navigation';
import { Icon } from '@/components/ui/Icon';
export function RefreshButton() { const router = useRouter(); const [pending, start] = useTransition(); return <button type="button" className="quiet-link" disabled={pending} onClick={() => start(() => router.refresh())}><Icon name="reload" size={15} />{pending ? 'Refreshing…' : 'Refresh'}</button>; }
