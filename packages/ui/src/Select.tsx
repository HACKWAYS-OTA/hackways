'use client';

import { useId, useRef, useState, type CSSProperties } from 'react';
import { Select as Primitive } from '@base-ui/react/select';
import { motion, useReducedMotion } from 'motion/react';
import { HugeiconsIcon } from '@hugeicons/react';
import ArrowDown01Icon from '@hugeicons/core-free-icons/ArrowDown01Icon';
import Tick02Icon from '@hugeicons/core-free-icons/Tick02Icon';

export type SelectOption<Value extends string> = { value: Value; label: string; disabled?: boolean };
export type SelectProps<Value extends string> = {
  options: readonly SelectOption<Value>[]; value: Value | ''; onValueChange: (value: Value) => void;
  label: string; name?: string; id?: string; required?: boolean; disabled?: boolean;
  placeholder?: string; autoFocus?: boolean; className?: string;
};

export function Select<Value extends string>({ options, value, onValueChange, label, placeholder = 'Choose an option', autoFocus, className = '', id, ...props }: SelectProps<Value>) {
  const trigger = useRef<HTMLButtonElement>(null);
  const [container, setContainer] = useState<HTMLElement | null>(null);
  const [palette, setPalette] = useState<CSSProperties>({});
  const highlight = useId();
  const reduced = useReducedMotion();
  return <Primitive.Root {...props} items={options} value={value || null} onValueChange={next => { const option = options.find(item => item.value === next); if (option) onValueChange(option.value); }} onOpenChange={open => {
    if (!open || !trigger.current) return;
    setContainer(trigger.current.closest('dialog'));
    const style = getComputedStyle(trigger.current);
    setPalette({ ...Object.fromEntries(['--hw-ink', '--hw-paper', '--hw-line', '--hw-muted', '--hw-fill'].map(token => [token, style.getPropertyValue(token)])), fontFamily: style.fontFamily } as CSSProperties);
  }}>
    <Primitive.Trigger ref={trigger} id={id} autoFocus={autoFocus} aria-label={label} className={`hw-select-trigger ${className}`}>
      <Primitive.Value placeholder={placeholder} /><Primitive.Icon className="hw-select-icon"><HugeiconsIcon icon={ArrowDown01Icon} size={17} strokeWidth={1.8} /></Primitive.Icon>
    </Primitive.Trigger>
    <Primitive.Portal container={container}>
      <Primitive.Positioner sideOffset={6} alignItemWithTrigger={false} className="hw-select-positioner" style={palette}>
        <Primitive.Popup className="hw-select-popup">
          <Primitive.List className="hw-select-list">{options.map(option => <Primitive.Item key={option.value} value={option.value} disabled={option.disabled} className="hw-select-item" render={(itemProps, state) => <div {...itemProps}>
            {state.highlighted && <motion.span className="hw-select-highlight" aria-hidden="true" layoutId={reduced ? undefined : highlight} transition={{ type: 'spring', stiffness: 500, damping: 40 }} />}
            {itemProps.children}
            <motion.span aria-hidden="true" className="hw-select-check" initial={false} animate={{ opacity: state.selected ? 1 : 0, scale: reduced ? 1 : state.selected ? 1 : .5 }} transition={reduced ? { duration: .15 } : { type: 'spring', stiffness: 500, damping: 24 }}><HugeiconsIcon icon={Tick02Icon} size={17} strokeWidth={2} /></motion.span>
          </div>}>
            <Primitive.ItemText className="hw-select-text">{option.label}</Primitive.ItemText>
          </Primitive.Item>)}</Primitive.List>
        </Primitive.Popup>
      </Primitive.Positioner>
    </Primitive.Portal>
  </Primitive.Root>;
}
