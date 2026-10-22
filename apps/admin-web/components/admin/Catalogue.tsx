"use client";

import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { useCatalogue } from '@/lib/catalogue';
import { PAGE_SIZE, cataloguePath } from '@/lib/catalogue-paths';
import { adminModules, type AdminModule, type AdminCell, type AdminColumn, type AdminRecord } from '@/lib/admin-modules';
import { detailPath } from '@/lib/admin-detail-paths';
import { useEffect, useState } from 'react';
import { Checkbox } from '@hackways/ui/checkbox';
import { BulkActions, supportsBulkActions } from './BulkActions';

function cell(value: AdminCell | undefined, column: AdminColumn) {
  if (value == null || value === '') return '?';
  if (column.booleanLabels) return column.booleanLabels[value ? 1 : 0];
  if (column.date && typeof value === 'string') {
    const date = new Date(value);
    if (!Number.isNaN(date.getTime())) return new Intl.DateTimeFormat(undefined, {dateStyle:'medium',timeStyle:'short'}).format(date);
  }
  return ['status', 'access', 'kind', 'template_type'].includes(column.key) ? String(value).replaceAll('_', ' ') : String(value);
}

export function Catalogue({ resource, query, page, status }: { resource: AdminModule; query: string; page: number; status: string }) {
  const definition = adminModules[resource];
  const { rows, loading, error, refresh } = useCatalogue(resource, query, page, status);
  const [selected, setSelected] = useState<AdminRecord[]>([]);
  const [bulkBusy, setBulkBusy] = useState(false);
  const [columns, setColumns] = useState<string[]>(definition.columns.map(column => column.key));
  useEffect(() => { try { const saved: unknown = JSON.parse(localStorage.getItem(`hackways.admin.columns.${resource}`) || 'null'); if (Array.isArray(saved)) setColumns(definition.columns.filter((column, index) => index === 0 || saved.includes(column.key)).map(column => column.key)); } catch { /* Invalid display preferences use the complete table. */ } }, [resource, definition]);
  function toggleColumn(key: string) { setColumns(previous => { const next = previous.includes(key) ? previous.filter(value => value !== key) : [...previous, key]; try { localStorage.setItem(`hackways.admin.columns.${resource}`, JSON.stringify(next)); } catch { /* The current view still works when browser storage is unavailable. */ } return next; }); }
  const visibleColumns = definition.columns.filter(column => columns.includes(column.key));
  const bulk = supportsBulkActions(resource);
  const allSelected = rows.length > 0 && rows.every(row => selected.some(item => item.reference === row.reference));
  const someSelected = rows.some(row => selected.some(item => item.reference === row.reference));
  function selectRow(row: AdminRecord) { setSelected(previous => previous.some(item => item.reference === row.reference) ? previous.filter(item => item.reference !== row.reference) : previous.length < 20 ? [...previous, row] : previous); }
  return <>
    <div className="admin-page-heading"><h1>{definition.label}</h1>{resource === 'messaging' && <Link className="app-button app-button-primary" href="/messaging/new">Compose message</Link>}<details className="admin-table-options"><summary className="app-button app-button-outline">Customize table</summary><fieldset><legend>Visible columns</legend>{definition.columns.map((column, index) => <label key={column.key}><Checkbox checked={columns.includes(column.key)} disabled={index === 0} onCheckedChange={() => toggleColumn(column.key)} />{column.label}</label>)}</fieldset></details></div>
    {bulk && <BulkActions resource={resource} records={selected} onFinished={refresh} onClear={() => setSelected([])} onBusyChange={setBulkBusy} />}
    {definition.statuses.length > 0 && <nav className="admin-filters" aria-label={resource === 'storage' ? 'Asset type' : 'Filter records'}>
      {['', ...definition.statuses].map(value => <Link key={value} href={cataloguePath(resource,query,1,value)} scroll={false} prefetch={false} aria-current={status === value ? 'page' : undefined}>{value ? value.replaceAll('_',' ') : 'All'}</Link>)}
      {query && <a href={`/${resource}`} className="quiet-link">Clear search</a>}
    </nav>}
    {loading ? <div className="admin-result-state admin-loading" role="status" aria-label="Loading records"><span className="wordmark admin-loading-wordmark" aria-hidden="true">hackways.</span><span className="admin-loading-line" aria-hidden="true" /><span className="sr-only">Loading records</span></div> : error ? <div className="admin-result-state" role="alert"><p>{error}</p><Button onClick={refresh} variant="outline">Try again</Button></div> : <>
      {rows.length === 0 ? <div className="admin-result-state" role="status">No records found.</div> : <div className="admin-table-scroll" tabIndex={0} role="region" aria-label={`${definition.label} table`}><table className="admin-table"><caption className="sr-only">{definition.label}</caption><thead><tr>{bulk && <th scope="col"><Checkbox aria-label="Select this page, up to twenty records" disabled={bulkBusy} checked={allSelected} indeterminate={someSelected && !allSelected} onCheckedChange={() => setSelected(allSelected ? [] : rows.slice(0, 20))} /></th>}{visibleColumns.map(column => <th key={column.key} scope="col">{column.label}</th>)}</tr></thead><tbody>
        {rows.map(row => <tr className="admin-clickable-row" key={`${row.reference}-${row.kind || ''}`}>{bulk && <td className="admin-selection-cell"><Checkbox aria-label={`Select ${String(row[definition.columns[0].key])}`} checked={selected.some(item => item.reference === row.reference)} disabled={bulkBusy || selected.length >= 20 && !selected.some(item => item.reference === row.reference)} onCheckedChange={() => selectRow(row)} /></td>}{visibleColumns.map((column,index) => index===0 ? <th key={column.key} scope="row"><a className="admin-record-link" href={detailPath(resource,row.reference,typeof row.kind === 'string' ? row.kind : undefined)}><span>{cell(row[column.key],column)}</span></a></th> : <td key={column.key}>{cell(row[column.key],column)}</td>)}</tr>)}
      </tbody></table></div>}
      <nav className="admin-pagination" aria-label="Pagination">
        {page > 1 && <a className="app-button app-button-outline app-button-sm" href={cataloguePath(resource,query,page-1,status)}>Previous</a>}
        <span>Page {page}</span>
        {rows.length === PAGE_SIZE && <a className="app-button app-button-outline app-button-sm" href={cataloguePath(resource,query,page+1,status)}>Next</a>}
      </nav>
    </>}
  </>;
}
