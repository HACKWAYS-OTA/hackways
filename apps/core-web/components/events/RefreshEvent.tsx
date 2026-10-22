"use client";
import { useTransition } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/Button';
export function RefreshEvent({ label = 'Try again' }: { label?: string }) {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  return <Button variant="outline" isLoading={pending} onClick={() => startTransition(() => router.refresh())}>{pending ? 'Trying again...' : label}</Button>;
}
