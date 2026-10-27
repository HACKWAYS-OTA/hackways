'use client';

import { useEffect, useId, useRef, useState, type FormEvent } from 'react';
import { Select } from '@hackways/ui/select';
import type { WorkspaceMember, WorkspaceTeam } from '@/types/api';
import { ApiClientError, fetchApi } from '@/lib/api';

const roles = [{ value: 'manager', label: 'Manager' }, { value: 'editor', label: 'Editor' }, { value: 'checkin', label: 'Check-in staff' }];
type Role = WorkspaceMember['role'];
const validRole = (value: string): value is Role => value === 'manager' || value === 'editor' || value === 'checkin';

export function TeamAccess({ kind, resourceID, initial }: { kind: 'space' | 'event'; resourceID: string; initial: WorkspaceTeam }) {
  const id = useId(), request = useRef<AbortController>();
  const [team, setTeam] = useState(initial), [email, setEmail] = useState(''), [role, setRole] = useState<Role>('editor');
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState(''), [uncertain, setUncertain] = useState(false);
  const endpoint = `/workspace/${kind}s/${resourceID}/access`;
  useEffect(() => () => request.current?.abort(), []);
  async function reload(offset = team.offset) {
    if (busy) return; setBusy(true); setError('');
    const controller = new AbortController(); request.current = controller;
    try { setTeam(await fetchApi<WorkspaceTeam>(`${endpoint}?limit=50&offset=${offset}`, { cache: 'no-store', signal: controller.signal })); setUncertain(false); }
    catch (failure) { if (!controller.signal.aborted) setError(failure instanceof Error ? failure.message : 'Team access could not be loaded.'); }
    finally { if (!controller.signal.aborted) setBusy(false); }
  }
  async function change(member: { email: string; revision: number }, nextRole: Role, remove?: WorkspaceMember) {
    if (busy || uncertain) return; setBusy(true); setError(''); setMessage('');
    const controller = new AbortController(); request.current = controller;
    try {
      await fetchApi(remove ? `${endpoint}/${remove.id}` : endpoint, { method: remove ? 'DELETE' : 'POST', signal: controller.signal, body: JSON.stringify(remove ? { revision: remove.revision } : { email: member.email, role: nextRole, revision: member.revision }) });
      setMessage(remove ? 'Direct access removed.' : 'Access saved.');
      try { setTeam(await fetchApi<WorkspaceTeam>(`${endpoint}?limit=50&offset=${team.offset}`, { cache: 'no-store', signal: controller.signal })); setUncertain(false); }
      catch (failure) { setUncertain(true); throw new Error('Access was saved. Reload the team to see its current state.'); }
    } catch (failure) {
      if (!controller.signal.aborted) { setError(failure instanceof Error ? failure.message : 'Access could not be saved.'); if (!(failure instanceof ApiClientError && ['INVALID_ACCESS', 'ACCOUNT_NOT_AVAILABLE', 'OWNER_ACCESS'].includes(failure.code))) setUncertain(true); }
    } finally { if (!controller.signal.aborted) setBusy(false); }
  }
  function add(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); const value = email.trim().toLowerCase(), existing = team.data.find(member => member.email.toLowerCase() === value);
    void change({ email: value, revision: existing?.revision || 0 }, role);
  }
  return <section className="workspace-team" aria-labelledby={`${id}-title`}>
    <header className="workspace-section-heading"><h2 id={`${id}-title`}>{kind === 'event' ? 'Direct event access' : 'Team access'}</h2><button type="button" className="app-button app-button-outline app-button-sm" disabled={busy} onClick={() => void reload()}>Reload team</button></header>
    {kind === 'event' && <p className="workspace-access-note">Removing an event override restores their Space role, if they have one.</p>}
    <form className="workspace-invite" onSubmit={add}><label className="field" htmlFor={`${id}-email`}>Account email<input className="app-input" id={`${id}-email`} type="email" autoComplete="off" required maxLength={254} value={email} onChange={event => setEmail(event.target.value)} disabled={busy || uncertain} /></label><Select label="Role" value={role} options={roles} disabled={busy || uncertain} onValueChange={value => { if (validRole(value)) setRole(value); }} /><button type="submit" className="app-button app-button-primary" disabled={busy || uncertain}>{busy ? 'Saving...' : 'Grant access'}</button></form>
    {error && <p className="journey-error" role="alert">{error}</p>}{message && <p className="workspace-confirmation" role="status">{message}</p>}
    <div className="workspace-table-scroll" role="region" aria-label="Assigned team access" tabIndex={0} aria-busy={busy}><table className="workspace-table"><caption className="visually-hidden">Directly assigned {kind} access</caption><thead><tr><th scope="col">Person</th><th scope="col">Role</th><th scope="col">Access</th><th scope="col"><span className="visually-hidden">Action</span></th></tr></thead><tbody>{team.data.length ? team.data.map(member => <tr key={member.id}><th scope="row"><strong>{member.display_name}</strong><span>{member.email}</span></th><td><Select label={`Role for ${member.display_name}`} value={member.role} options={roles} disabled={busy || uncertain || Boolean(member.revoked_at)} onValueChange={value => { if (validRole(value) && value !== member.role) void change(member, value); }} /></td><td>{member.revoked_at ? kind === 'event' ? 'Override removed' : 'Removed' : 'Active'}</td><td><button type="button" className="app-button app-button-outline app-button-sm" disabled={busy || uncertain} onClick={() => void change(member, member.role, member.revoked_at ? undefined : member)} aria-label={`${member.revoked_at ? 'Restore' : 'Remove'} access for ${member.display_name}`}>{member.revoked_at ? 'Restore' : kind === 'event' ? 'Remove override' : 'Remove'}</button></td></tr>) : <tr><td colSpan={4} className="workspace-table-empty">No access assignments yet.</td></tr>}</tbody></table></div>
    {team.inherited.length > 0 && <section className="workspace-inherited"><h2>Space access</h2><div className="workspace-table-scroll" role="region" aria-label="Access inherited from the Space" tabIndex={0}><table className="workspace-table"><caption className="visually-hidden">Access inherited from the Space</caption><thead><tr><th scope="col">Person</th><th scope="col">Role</th></tr></thead><tbody>{team.inherited.map(member => <tr key={member.id}><th scope="row"><strong>{member.display_name}</strong><span>{member.email}</span></th><td>{roles.find(value => value.value === member.role)?.label}</td></tr>)}</tbody></table></div></section>}
    {(team.offset > 0 || team.data.length === 50 || team.inherited.length === 50) && <nav className="journey-pagination" aria-label="Team pages">{team.offset > 0 && <button type="button" className="app-button app-button-outline" disabled={busy} onClick={() => void reload(Math.max(0, team.offset - 50))}>Previous</button>}{(team.data.length === 50 || team.inherited.length === 50) && <button type="button" className="app-button app-button-outline" disabled={busy} onClick={() => void reload(team.offset + 50)}>Next</button>}</nav>}
  </section>;
}
