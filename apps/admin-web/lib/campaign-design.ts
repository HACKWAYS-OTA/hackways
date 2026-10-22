export type CampaignBlock = { kind: 'heading' | 'text' | 'image' | 'button' | 'divider'; text: string; url: string; alt: string };
export type CampaignDesign = {
  preheader: string; heading: string; brand: string; banner_url: string; banner_alt: string;
  background: string; surface: string; text: string; accent: string; button_text: string;
  font: 'sans' | 'serif' | 'mono'; align: 'left' | 'center'; layout: 'card' | 'plain'; blocks: CampaignBlock[];
};
export type CampaignConfiguration = { design: CampaignDesign; audience: 'all' | 'account' | 'import'; recipients: string[]; scheduled_at: string | null };
export function campaignConfiguration(value?: Partial<CampaignConfiguration>): CampaignConfiguration {
  return { audience: 'all', recipients: [], scheduled_at: null, ...value, design: {
    preheader: '', heading: '', brand: 'Hackways', banner_url: '', banner_alt: '',
    background: '#f5f2f8', surface: '#ffffff', text: '#29222f', accent: '#604176', button_text: '#ffffff',
    font: 'sans', align: 'left', layout: 'card', blocks: [], ...value?.design,
  } };
}
