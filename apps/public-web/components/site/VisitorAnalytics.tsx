"use client";
import { usePathname } from 'next/navigation';
import { AnalyticsTracker } from '@hackways/ui/analytics';
export function VisitorAnalytics(){return <AnalyticsTracker path={usePathname()} endpoint={`${(process.env.NEXT_PUBLIC_GLOBAL_API_URL || '/api/v1').replace(/\/$/,'')}/analytics/events`} />;}
