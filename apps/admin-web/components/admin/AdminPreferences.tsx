'use client';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { RadioGroup } from '@hackways/ui/radio-group';
import { defaultPreferences, type AdminPreferences as Preferences } from '@/lib/admin-preferences';
export function AdminPreferences({ initial }: { initial: Preferences }) {
  const router = useRouter();
  const [settings, setSettings] = useState(initial), [message, setMessage] = useState('');
  function apply(next: Preferences) {
    setSettings(next);
    for (const [key, value] of Object.entries(next)) document.documentElement.setAttribute(`data-admin-${key}`, value);
    document.cookie = `hackways_admin_display=${encodeURIComponent(JSON.stringify(next))}; Path=/; SameSite=Strict; Max-Age=31536000${location.protocol === 'https:' ? '; Secure' : ''}`;
    setMessage('Display preferences saved for this browser.');
    router.refresh();
  }
  return <div className="admin-preferences"><section><h2>Text size</h2><RadioGroup label="Admin text size" value={settings.text} onValueChange={text => apply({ ...settings, text })} options={[{ value: 'standard', label: 'Standard' }, { value: 'large', label: 'Larger' }] as const} /></section><section><h2>Table spacing</h2><RadioGroup label="Table spacing" value={settings.density} onValueChange={density => apply({ ...settings, density })} options={[{ value: 'comfortable', label: 'Comfortable' }, { value: 'compact', label: 'Compact' }] as const} /></section><section><h2>Contrast</h2><RadioGroup label="Admin contrast" value={settings.contrast} onValueChange={contrast => apply({ ...settings, contrast })} options={[{ value: 'standard', label: 'Standard' }, { value: 'high', label: 'High contrast' }] as const} /></section><section><h2>Motion</h2><RadioGroup label="Animation preference" value={settings.motion} onValueChange={motion => apply({ ...settings, motion })} options={[{ value: 'system', label: 'Follow device setting' }, { value: 'reduced', label: 'Reduce motion' }] as const} /></section><section><h2>Scrollbars</h2><RadioGroup label="Scrollbar visibility" value={settings.scrollbars} onValueChange={scrollbars => apply({ ...settings, scrollbars })} options={[{ value: 'hidden', label: 'Hidden' }, { value: 'visible', label: 'Visible' }] as const} /></section><button type="button" className="app-button app-button-outline" onClick={() => apply(defaultPreferences)}>Reset display preferences</button><p role="status" className="admin-action-success">{message}</p></div>;
}
