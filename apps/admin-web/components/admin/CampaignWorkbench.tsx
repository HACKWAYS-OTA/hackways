"use client";
import { useEffect, useState } from 'react';
import type { Campaign } from '@/lib/campaigns';
import { CampaignComposer } from './CampaignComposer';
import { CampaignActions } from './CampaignEditor';
export function CampaignWorkbench({ campaign, audienceCount }: { campaign: Campaign; audienceCount: number }) {
  const [dirty, setDirty] = useState(false);
  useEffect(() => setDirty(false), [campaign.revision]);
  return <><CampaignComposer key={campaign.revision} campaign={campaign} onDirtyChange={() => setDirty(true)} /><CampaignActions key={campaign.revision} campaign={campaign} audienceCount={audienceCount} disabled={dirty} /></>;
}
