'use client';
import { useState } from 'react';
import { HugeiconsIcon } from '@hugeicons/react';
import UserCircleIcon from '@hugeicons/core-free-icons/UserCircleIcon';

export function Avatar({ name, image, large = false }: { name?: string; image?: string | null; large?: boolean }) {
  const [failedSource, setFailedSource] = useState('');
  const initials = name?.trim().split(/\s+/).slice(0, 2).map(part => Array.from(part)[0]).join('').toUpperCase();
  const source = image?.startsWith('https://') && image !== failedSource ? image : null;
  return <span className={`hw-avatar${large ? ' hw-avatar-large' : ''}`} aria-hidden="true">{source ?
    // Remote account photos load directly without a Next.js image proxy or arbitrary server fetch.
    // eslint-disable-next-line @next/next/no-img-element
    <img src={source} alt="" width={large ? 96 : 40} height={large ? 96 : 40} referrerPolicy="no-referrer" onError={() => setFailedSource(source)} />
    : initials || <HugeiconsIcon icon={UserCircleIcon} size={large ? 40 : 24} strokeWidth={1.8} />}</span>;
}
