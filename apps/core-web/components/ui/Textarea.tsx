'use client';

import { useLayoutEffect, useRef, type TextareaHTMLAttributes } from 'react';
import './textarea.css';

type TextareaProps = Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'value' | 'defaultValue'> & { value: string };

export function Textarea({ value, className = '', rows = 2, ...props }: TextareaProps) {
  const input = useRef<HTMLTextAreaElement>(null);
  useLayoutEffect(() => {
    const element = input.current;
    if (!element) return;
    function resize() {
      if (!element || element.offsetParent === null) return;
      element.style.height = '0px';
      const border = parseFloat(getComputedStyle(element).borderBottomWidth) || 0;
      element.style.height = `${element.scrollHeight + border}px`;
    }
    resize();
    // Width changes affect wrapping; height changes from typing need no second observer pass.
    let width = element.getBoundingClientRect().width;
    const observer = new ResizeObserver(entries => {
      const nextWidth = entries[0]?.contentRect.width;
      if (nextWidth !== undefined && nextWidth !== width) { width = nextWidth; resize(); }
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, [value]);
  return <textarea {...props} ref={input} value={value} rows={rows} className={`text-area ${className}`} />;
}
