"use client";

import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { useCatalogue } from '@/lib/catalogue';
import { PAGE_SIZE, cataloguePath } from '@/lib/catalogue-paths';
import { adminModules, type AdminModule, type AdminCell, type AdminColumn } from '@/lib/admin-modules';
import { detailPath } from '@/lib/admin-detail-paths';

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
  return <>
    <div className="admin-page-heading"><h1>{definition.label}</h1></div>
    {definition.statuses.length > 0 && <nav className="admin-filters" aria-label={resource === 'storage' ? 'Asset type' : 'Filter records'}>
      {['', ...definition.statuses].map(value => <Link key={value} href={cataloguePath(resource,query,1,value)} scroll={false} prefetch={false} aria-current={status === value ? 'page' : undefined}>{value ? value.replaceAll('_',' ') : 'All'}</Link>)}
      {query && <a href={`/${resource}`} className="quiet-link">Clear search</a>}
    </nav>}
    {loading ? <div className="admin-result-state admin-loading" role="status" aria-label="Loading records"><span className="wordmark admin-loading-wordmark" aria-hidden="true">hackways.</span><span className="admin-loading-line" aria-hidden="true" /><span className="sr-only">Loading records</span></div> : error ? <div className="admin-result-state" role="alert"><p>{error}</p><Button onClick={refresh} variant="outline">Try again</Button></div> : <>
      {rows.length === 0 ? <div className="admin-result-state" role="status">No records found.</div> : <div className="admin-table-scroll" tabIndex={0} role="region" aria-label={`${definition.label} table`}><table className="admin-table"><caption className="sr-only">{definition.label}</caption><thead><tr>{definition.columns.map(column => <th key={column.key} scope="col">{column.label}</th>)}</tr></thead><tbody>
        {rows.map(row => <tr key={`${row.reference}-${row.kind || ''}`}>{definition.columns.map((column,index) => index===0 ? <th key={column.key} scope="row"><a className="admin-record-link" href={detailPath(resource,row.reference,typeof row.kind === 'string' ? row.kind : undefined)}><span>{cell(row[column.key],column)}</span></a></th> : <td key={column.key}>{cell(row[column.key],column)}</td>)}</tr>)}
      </tbody></table></div>}
      <nav className="admin-pagination" aria-label="Pagination">
        {page > 1 && <a className="app-button app-button-outline app-button-sm" href={cataloguePath(resource,query,page-1,status)}>Previous</a>}
        <span>Page {page}</span>
        {rows.length === PAGE_SIZE && <a className="app-button app-button-outline app-button-sm" href={cataloguePath(resource,query,page+1,status)}>Next</a>}
      </nav>
    </>}
  </>;
}
