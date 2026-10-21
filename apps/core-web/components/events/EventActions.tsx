"use client";
import { useState } from 'react';
import type { Event } from '@/types/api';
import { eventPath } from '@/lib/events';
import { Button } from '@/components/ui/Button';
import { Icon } from '@/components/ui/Icon';

function escapeCalendar(value: string): string { return value.replace(/\0/g, '').replace(/\\/g, '\\\\').replace(/\r\n|\r|\n/g, '\\n').replace(/;/g, '\\;').replace(/,/g, '\\,'); }
function foldCalendar(line: string): string {
  let result = '', size = 0;
  const encoder = new TextEncoder();
  for (const char of line) {
    const bytes = encoder.encode(char).length;
    if (size + bytes > 75) { result += '\r\n '; size = 1; }
    result += char; size += bytes;
  }
  return result;
}
export function EventActions({ event }: { event: Event }) {
  const [message, setMessage] = useState('');
  const [copyValue, setCopyValue] = useState('');
  async function copy() {
    const url = `${window.location.origin}${eventPath(event)}`;
    try { await navigator.clipboard.writeText(url); setCopyValue(''); setMessage('Event link copied.'); }
    catch { setCopyValue(url); setMessage('Select and copy the event link below.'); }
  }
  function download() {
    try {
      const utc = (value: string) => new Date(value).toISOString().replace(/[-:]/g, '').replace(/\.\d{3}Z$/, 'Z');
      const lines = ['BEGIN:VCALENDAR', 'VERSION:2.0', 'PRODID:-//Hackways//Events//EN', 'CALSCALE:GREGORIAN', 'BEGIN:VEVENT', `UID:${event.slug}@hackways`, `DTSTAMP:${utc(new Date().toISOString())}`, `DTSTART:${utc(event.starts_at)}`, `DTEND:${utc(event.ends_at)}`, `SUMMARY:${escapeCalendar(event.title)}`, `DESCRIPTION:${escapeCalendar(event.description)}`, `URL:${escapeCalendar(`${window.location.origin}${eventPath(event)}`)}`];
      const location = [event.location_name, event.location_address].filter(Boolean).join(', ') || (event.format === 'online' ? event.online_url : null);
      if (location) lines.push(`LOCATION:${escapeCalendar(location)}`);
      lines.push('END:VEVENT', 'END:VCALENDAR');
      const url = URL.createObjectURL(new Blob([lines.map(foldCalendar).join('\r\n') + '\r\n'], { type: 'text/calendar;charset=utf-8' }));
      const link = document.createElement('a');
      link.href = url; link.download = `${event.slug.replace(/[^a-zA-Z0-9_-]/g, '-')}.ics`;
      document.body.appendChild(link); link.click(); link.remove();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
      setMessage('Calendar file prepared.');
    } catch { setMessage('The calendar file could not be prepared. Please try again.'); }
  }
  return <div className="event-actions"><div><Button variant="outline" size="sm" onClick={download}><Icon name="calendar" size={16} />Add to calendar</Button><Button variant="outline" size="sm" onClick={() => { void copy(); }}><Icon name="external" size={16} />Copy event link</Button></div>{message && <p role="status">{message}</p>}{copyValue && <label className="field">Event link<input className="app-input" readOnly value={copyValue} onFocus={(action) => action.target.select()} /></label>}</div>;
}
