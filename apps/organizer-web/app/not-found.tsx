import { PlatformNotFound } from '@hackways/ui/not-found';
import '@hackways/ui/not-found.css';
export default function NotFound(){ return <PlatformNotFound title="This page isn?t in your workspace." description="Check the address and make sure you?re using the account with access to this event or Space." primaryHref={process.env.NEXT_PUBLIC_ORGANIZER_URL||'/manage'} primaryLabel="Back to workspace"/>; }
