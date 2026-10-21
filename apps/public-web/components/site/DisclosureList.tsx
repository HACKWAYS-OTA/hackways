export function DisclosureList({ items, numbered = false }: {
  items: readonly { title: string; body: string }[];
  numbered?: boolean;
}) {
  const entries = items.map((item, index) => (
    <details key={item.title} open={index === 0}>
      <summary>{numbered && <span className="disclosure-number" aria-hidden="true">{index + 1}</span>}<span>{item.title}</span><span className="disclosure-toggle" aria-hidden="true" /></summary>
      <div><p>{item.body}</p></div>
    </details>
  ));

  return numbered
    ? <ol className="disclosure-list disclosure-list--numbered" role="list">{entries.map((entry, index) => <li key={items[index].title}>{entry}</li>)}</ol>
    : <div className="disclosure-list">{entries}</div>;
}
