'use client';

import { useState } from 'react';
import { Checkbox as Primitive } from '@base-ui/react/checkbox';
import { motion, useReducedMotion } from 'motion/react';
import { HugeiconsIcon } from '@hugeicons/react';
import Tick02Icon from '@hugeicons/core-free-icons/Tick02Icon';
import MinusSignIcon from '@hugeicons/core-free-icons/MinusSignIcon';

export type CheckboxProps = Omit<Primitive.Root.Props, 'className' | 'children'> & { className?: string };

export function Checkbox({ checked, defaultChecked = false, onCheckedChange, indeterminate, className = '', ...props }: CheckboxProps) {
  const [local, setLocal] = useState(defaultChecked);
  const reduced = useReducedMotion();
  const selected = checked ?? local;
  const visible = selected || !!indeterminate;
  return <Primitive.Root {...props} checked={selected} indeterminate={indeterminate} className={`hw-checkbox ${className}`} onCheckedChange={(next, details) => { if (checked === undefined) setLocal(next); onCheckedChange?.(next, details); }}>
    <motion.span aria-hidden="true" className="hw-check-indicator" initial={false} animate={{ opacity: visible ? 1 : 0, scale: reduced ? 1 : visible ? 1 : .5 }} transition={reduced ? { duration: .15 } : visible ? { type: 'spring', stiffness: 500, damping: 24 } : { duration: .15 }}>
      <HugeiconsIcon icon={indeterminate ? MinusSignIcon : Tick02Icon} size={15} strokeWidth={2} />
    </motion.span>
  </Primitive.Root>;
}
