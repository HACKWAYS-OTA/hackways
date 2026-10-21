'use client';

import { forwardRef } from 'react';
import { Toggle as Primitive } from '@base-ui/react/toggle';
import { motion, useReducedMotion, type Transition } from 'motion/react';

export type ToggleProps<Value extends string = string> = Primitive.Props<Value> & { transition?: Transition };

export const Toggle = forwardRef<HTMLButtonElement, ToggleProps>(function Toggle({ className, render, disabled, transition = { type: 'spring', stiffness: 550, damping: 30, restSpeed: 10 }, ...props }, ref) {
  const reduced = useReducedMotion();
  return <Primitive {...props} ref={ref} disabled={disabled}
    className={state => `hw-toggle ${typeof className === 'function' ? className(state) ?? '' : className ?? ''}`}
    render={render ?? <motion.button whileTap={reduced || disabled ? undefined : { scale: .97 }} transition={transition} />}
  />;
});

export namespace Toggle {
  export type Props<Value extends string = string> = ToggleProps<Value>;
  export type State = Primitive.State;
  export type ChangeEventReason = Primitive.ChangeEventReason;
  export type ChangeEventDetails = Primitive.ChangeEventDetails;
}
