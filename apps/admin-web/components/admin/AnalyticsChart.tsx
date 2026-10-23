import { analyticsNumber } from '@/lib/analytics';

export function AnalyticsChart({ id, title, rows, first, second }: {
  id: string; title: string; rows: { day: string; first: number; second: number }[];
  first: string; second: string;
}) {
  const maximum = Math.max(1, ...rows.map(row => Math.max(row.first, row.second)));
  const width = 640, height = 240, left = 44, bottom = 28, top = 12;
  const innerWidth = width - left - 12, innerHeight = height - bottom - top;
  const group = innerWidth / Math.max(1, rows.length), bar = Math.max(.5, Math.min(16, group * .35));
  const activity = rows.some(row => row.first || row.second);
  const date = (value: string) => new Date(`${value}T00:00:00Z`).toLocaleDateString('en-IN', { month: 'short', day: 'numeric', timeZone: 'UTC' });
  return <section className="analytics-panel analytics-chart">
    <header><h2 id={`${id}-title`}>{title}</h2><ul className="analytics-legend"><li><span className="is-primary" />{first}</li><li><span className="is-secondary" />{second}</li></ul></header>
    {activity ? <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-labelledby={`${id}-title ${id}-description`}>
      <desc id={`${id}-description`}>Daily {first.toLowerCase()} and {second.toLowerCase()} in UTC. Exact values follow in the daily figures table.</desc>
      {[0, Math.ceil(maximum / 2), maximum].filter((value, index, all) => all.indexOf(value) === index).map(value => <g key={value}><line x1={left} x2={width - 12} y1={top + innerHeight * (1 - value / maximum)} y2={top + innerHeight * (1 - value / maximum)} stroke="var(--line)" /><text x={left - 8} y={top + innerHeight * (1 - value / maximum) + 4} textAnchor="end">{analyticsNumber(value)}</text></g>)}
      {rows.map((row, index) => <g key={row.day}><title>{row.day}: {first} {row.first}; {second} {row.second}</title><rect className="analytics-bar-primary" x={left + group * index + group / 2 - bar} y={top + innerHeight * (1 - row.first / maximum)} width={bar} height={innerHeight * row.first / maximum} /><rect className="analytics-bar-secondary" x={left + group * index + group / 2 + 1} y={top + innerHeight * (1 - row.second / maximum)} width={bar} height={innerHeight * row.second / maximum} /></g>)}
      {rows.length > 0 && [0, Math.floor((rows.length - 1) / 2), rows.length - 1].filter((value, index, all) => all.indexOf(value) === index).map(index => <text key={index} x={left + group * index + group / 2} y={height - 6} textAnchor="middle">{date(rows[index].day)}</text>)}
    </svg> : <div className="analytics-chart-empty"><p>No activity recorded in this period.</p></div>}
    <details className="analytics-daily-data"><summary>View daily figures</summary><div className="analytics-table-scroll" tabIndex={0} role="region" aria-label={`${title} daily data`}><table><caption className="sr-only">{title}, daily figures in UTC</caption><thead><tr><th scope="col">Day (UTC)</th><th scope="col">{first}</th><th scope="col">{second}</th></tr></thead><tbody>{rows.map(row => <tr key={row.day}><th scope="row">{row.day}</th><td>{analyticsNumber(row.first)}</td><td>{analyticsNumber(row.second)}</td></tr>)}</tbody></table></div></details>
  </section>;
}
