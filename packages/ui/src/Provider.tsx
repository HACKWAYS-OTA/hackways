'use client';

import { MotionConfig } from 'motion/react';
import type { ReactNode } from 'react';

/** Shared animation policy for every Hackways app. */
export function UIProvider({ children }: { children: ReactNode }) {
  return <MotionConfig reducedMotion="user" transition={{ type: 'spring', stiffness: 500, damping: 30 }}>{children}</MotionConfig>;
}
