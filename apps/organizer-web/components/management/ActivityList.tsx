import { activityLabel, type Activity } from '@/lib/dashboard';
import { dateLabel } from '@/lib/journeys';
export function ActivityList({ items, zone }: { items: Activity[]; zone: string }) {
  return items.length ? <ol className="host-activity">{items.map(item => <li key={item.id}><span className="host-activity-dot" aria-hidden="true" /><div><p>{activityLabel(item.action)}</p><time dateTime={item.created_at}>{dateLabel(item.created_at, zone)}</time></div></li>)}</ol> : <p className="host-muted">No activity has been recorded for this event yet.</p>;
}
