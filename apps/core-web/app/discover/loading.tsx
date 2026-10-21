import { EventCardSkeleton } from '@/components/ui/Skeleton';
export default function Loading() { return <div className="feed-container route-loading" role="status"><div className="skeleton skeleton-heading" /><div className="feed-loading"><EventCardSkeleton prominent /><EventCardSkeleton /></div><span className="visually-hidden">Loading events...</span></div>; }
