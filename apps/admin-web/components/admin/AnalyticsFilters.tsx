'use client';
import { useState } from 'react';
import { Button } from '@/components/ui/Button';

export function AnalyticsFilters({ from, to, today, onEdit }: { from: string; to: string; today: string; onEdit?: () => void }) {
  const [range, setRange] = useState({ from, to });
  const min = new Date(Date.parse(`${today}T00:00:00Z`) - 89 * 86400000).toISOString().slice(0, 10);
  return <form action="/analytics" method="get" className="analytics-filters">
    <label className="field">From (UTC)<input name="from" className="app-input" type="date" value={range.from} min={min} max={range.to || today} required onChange={event => { onEdit?.(); setRange(current => ({ ...current, from: event.target.value })); }} /></label>
    <label className="field">To (UTC)<input name="to" className="app-input" type="date" value={range.to} min={range.from || min} max={today} required onChange={event => { onEdit?.(); setRange(current => ({ ...current, to: event.target.value })); }} /></label>
    <Button type="submit">Apply</Button>
  </form>;
}
