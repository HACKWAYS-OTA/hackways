'use client';
import { useState } from 'react';
export function CopyEndpoint() {
  const [status, setStatus] = useState('');
  async function copy() {
    try { await navigator.clipboard.writeText('https://mcp.hackways.in/mcp'); setStatus('Server URL copied.'); }
    catch { setStatus('Copy the server URL shown above.'); }
  }
  return <div><button type="button" className="button button-primary" onClick={() => void copy()}>Copy server URL</button><p className="mcp-copy-status" role="status">{status}</p></div>;
}
