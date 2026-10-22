'use client';

import { MotionConfig } from 'motion/react';
import type { ReactNode } from 'react';

/** Shared animation policy for every Hackways app. */
export function UIProvider({ children, reducedMotion = 'user' }: { children: ReactNode; reducedMotion?: 'user' | 'always' }) {
  return <MotionConfig reducedMotion={reducedMotion} transition={{ type: 'spring', stiffness: 500, damping: 30 }}>{children}</MotionConfig>;
}
