'use client';

import { useEffect, useId, useMemo, useRef, useState, type KeyboardEvent } from 'react';
import { Icon } from '@/components/ui/Icon';

export function EventTimeZoneField({ value, onChange, zones, disabled }: { value: string; onChange: (value: string) => void; zones: readonly string[]; disabled: boolean }) {
  const id = useId(), menu = useRef<HTMLDetailsElement>(null), search = useRef<HTMLInputElement>(null), list = useRef<HTMLDivElement>(null);
  const [device, setDevice] = useState(''), [query, setQuery] = useState(''), [open, setOpen] = useState(false), [active, setActive] = useState(0);
  useEffect(() => { const zone = Intl.DateTimeFormat().resolvedOptions().timeZone; setDevice(zone); if (!value && zone) onChange(zone); }, [value, onChange]);
  useEffect(() => {
    function outside(event: PointerEvent) { if (event.target instanceof Node && !menu.current?.contains(event.target)) menu.current?.removeAttribute('open'); }
    function escape(event: globalThis.KeyboardEvent) { if (event.key === 'Escape' && menu.current?.open) { event.preventDefault(); menu.current.open = false; menu.current.querySelector('summary')?.focus(); } }
    document.addEventListener('pointerdown', outside); document.addEventListener('keydown', escape);
    return () => { document.removeEventListener('pointerdown', outside); document.removeEventListener('keydown', escape); };
  }, []);
  const choices = useMemo(() => Array.from(new Set([...zones, 'UTC', value, device].filter(Boolean))).sort(), [zones, value, device]);
  const results = useMemo(() => choices.filter(zone => zone.replaceAll('_', ' ').toLowerCase().includes(query.trim().replaceAll('_', ' ').toLowerCase())), [choices, query]);
  function choose(zone: string) { if (disabled) return; onChange(zone); if (menu.current) menu.current.open = false; menu.current?.querySelector('summary')?.focus(); }
  function listKey(event: KeyboardEvent<HTMLDivElement>) {
    if ((event.key === 'Enter' || event.key === ' ') && results[active]) { event.preventDefault(); choose(results[active]); return; }
    const next = event.key === 'ArrowDown' ? Math.min(results.length - 1, active + 1) : event.key === 'ArrowUp' ? Math.max(0, active - 1) : event.key === 'Home' ? 0 : event.key === 'End' ? results.length - 1 : null;
    if (next === null || next < 0) return;
    event.preventDefault(); setActive(next); list.current?.children[next]?.scrollIntoView({ block: 'nearest' });
  }
  return <details ref={menu} className="event-zone-menu" onToggle={event => { const expanded = event.currentTarget.open; setOpen(expanded); if (expanded) { setQuery(''); setActive(0); requestAnimationFrame(() => search.current?.focus()); } }}>
    <input type="hidden" name="timezone" value={value} />
    <summary aria-disabled={disabled || undefined} onClick={event => { if (disabled) event.preventDefault(); }}><Icon name="globe" size={20} /><span>Time zone</span><strong>{value || 'Detecting device zone…'}</strong><Icon name="down" size={14} /></summary>
    {open && <div className="event-zone-dropdown"><label className="event-zone-search"><Icon name="search" size={17} /><input ref={search} value={query} onChange={event => { setQuery(event.target.value); setActive(0); }} placeholder="Search time zones" aria-label="Search all time zones" autoComplete="off" onKeyDown={event => { if (event.key === 'ArrowDown') { event.preventDefault(); list.current?.focus(); } }} /></label>
      {device && <button type="button" className="event-zone-device" disabled={disabled} onClick={() => choose(device)}><span>Use device time zone<small>{device}</small></span><Icon name="arrow" size={16} /></button>}
      <div ref={list} className="event-zone-options" role="listbox" aria-label="Time zones" tabIndex={0} aria-activedescendant={results.length ? `${id}-${active}` : undefined} onKeyDown={listKey}>{results.map((zone, index) => <div id={`${id}-${index}`} key={zone} role="option" aria-selected={value === zone} data-active={index === active || undefined} onClick={() => { if (!disabled) choose(zone); }}><span>{zone.replaceAll('_', ' ')}</span>{value === zone && <Icon name="check" size={16} />}</div>)}</div>
      {!results.length && <p className="event-zone-empty" role="status">No matching time zones. Try a city or region.</p>}
    </div>}
  </details>;
}
