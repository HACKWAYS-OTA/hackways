'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Select } from '@hackways/ui/select';
import { AnalyticsFilters } from './AnalyticsFilters';

const options = [
  { value: '5m', label: '5 min' }, { value: '1h', label: '1 hr' }, { value: '24h', label: '24 hr' },
  { value: '7', label: 'Last 7 days' }, { value: '30', label: 'Last 30 days' },
  { value: '90', label: 'Last 90 days' }, { value: 'custom', label: 'Custom dates' },
];

export function AnalyticsRangeControl({ value, from, to, today, onIntervalChange, datesVisible = false }: {
  value: string; from: string; to: string; today: string;
  datesVisible?: boolean;
  onIntervalChange?: (value: '5m' | '1h' | '24h') => void;
}) {
  const router = useRouter();
  const [custom, setCustom] = useState(false);
  return <div className="analytics-range-control">
    <Select label="Time range" value={custom ? 'custom' : value} options={options} onValueChange={next => {
      if (next === 'custom') { setCustom(true); return; }
      setCustom(false);
      if (next === '5m' || next === '1h' || next === '24h') {
        if (onIntervalChange) onIntervalChange(next);
        else router.push(`/analytics?window=${next}`);
      } else if (next === '7' || next === '30' || next === '90') {
        const start = new Date(Date.parse(`${today}T00:00:00Z`) - (Number(next) - 1) * 86400000).toISOString().slice(0, 10);
        router.push(`/analytics?${new URLSearchParams({ from: start, to: today })}`);
      }
    }} />
    {(custom || datesVisible) && <AnalyticsFilters from={from} to={to} today={today} onEdit={() => setCustom(true)} />}
  </div>;
}
