'use client';
import { useState } from 'react';
export function CopyEndpoint({ value = 'https://mcp.hackways.in/mcp', label = 'Copy server URL' }: { value?: string; label?: string }) {
  const [status, setStatus] = useState('');
  async function copy() {
    try { await navigator.clipboard.writeText(value); setStatus('Copied.'); }
    catch { setStatus('Copy the value shown above manually.'); }
  }
  return <div className="mcp-copy"><button type="button" className="button button-primary" onClick={() => void copy()}>{label}</button><p className="mcp-copy-status" role="status">{status}</p></div>;
}
