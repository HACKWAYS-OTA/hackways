'use client';

import { useState, type ReactNode } from 'react';
import { RadioGroup as Group } from '@base-ui/react/radio-group';
import { Radio } from '@base-ui/react/radio';
import { motion, useReducedMotion } from 'motion/react';

export type RadioOption<Value extends string> = { value: Value; label: ReactNode; description?: string; disabled?: boolean };
export type RadioGroupProps<Value extends string> = {
  options: readonly RadioOption<Value>[]; value?: Value | ''; defaultValue?: Value;
  onValueChange?: (value: Value) => void; name?: string; required?: boolean; disabled?: boolean;
  label: string; describedBy?: string; id?: string; layout?: 'inline' | 'stack'; className?: string;
};

export function RadioGroup<Value extends string>({ options, value, defaultValue, onValueChange, label, describedBy, layout = 'inline', className = '', ...props }: RadioGroupProps<Value>) {
  const [local, setLocal] = useState<Value | ''>(defaultValue ?? '');
  const selected = value ?? local;
  const reduced = useReducedMotion();
  return <Group {...props} value={selected} onValueChange={next => { const option = options.find(item => item.value === next); if (option) { if (value === undefined) setLocal(option.value); onValueChange?.(option.value); } }} aria-label={label} aria-describedby={describedBy} className={`hw-radio-group hw-radio-group--${layout} ${className}`}>
    {options.map(option => <label key={option.value} className="hw-radio-option" data-selected={selected === option.value || undefined} data-disabled={props.disabled || option.disabled || undefined}>
      <Radio.Root value={option.value} disabled={option.disabled || props.disabled} className="hw-radio" render={<motion.span whileTap={reduced || props.disabled || option.disabled ? undefined : { scale: .97 }} transition={{ type: 'spring', stiffness: 550, damping: 30 }} />}>
        <motion.span aria-hidden="true" className="hw-radio-dot" initial={false} animate={{ opacity: selected === option.value ? 1 : 0, scale: reduced ? 1 : selected === option.value ? 1 : 0 }} transition={{ type: 'spring', stiffness: 200, damping: 16 }} />
      </Radio.Root>
      <span className="hw-radio-copy">{option.label}{option.description && <small>{option.description}</small>}</span>
    </label>)}
  </Group>;
}
