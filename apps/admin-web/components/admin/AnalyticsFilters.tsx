'use client';
import { useState } from 'react';
import { Select } from '@hackways/ui/select';
import { Button } from '@/components/ui/Button';

export function AnalyticsFilters({ from, to, today }: { from: string; to: string; today: string }) {
  const [range, setRange] = useState({ from, to });
  const days = (Date.parse(`${to}T00:00:00Z`) - Date.parse(`${from}T00:00:00Z`)) / 86400000 + 1;
  const [preset, setPreset] = useState(to === today && [7, 30, 90].includes(days) ? String(days) : 'custom');
  const min = new Date(Date.parse(`${today}T00:00:00Z`) - 89 * 86400000).toISOString().slice(0, 10);
  return <form action="/analytics" method="get" className="analytics-filters">
    <Select label="Date range" value={preset} options={[{ value: '7', label: 'Last 7 days' }, { value: '30', label: 'Last 30 days' }, { value: '90', label: 'Last 90 days' }, { value: 'custom', label: 'Custom dates' }]} onValueChange={value => {
      setPreset(value); if (value !== 'custom') setRange({ from: new Date(Date.parse(`${today}T00:00:00Z`) - (Number(value) - 1) * 86400000).toISOString().slice(0, 10), to: today });
    }} />
    <label className="field">From (UTC)<input name="from" className="app-input" type="date" value={range.from} min={min} max={range.to || today} required onChange={event => { setPreset('custom'); setRange(current => ({ ...current, from: event.target.value })); }} /></label>
    <label className="field">To (UTC)<input name="to" className="app-input" type="date" value={range.to} min={range.from || min} max={today} required onChange={event => { setPreset('custom'); setRange(current => ({ ...current, to: event.target.value })); }} /></label>
    <Button type="submit">Apply</Button>
  </form>;
}
