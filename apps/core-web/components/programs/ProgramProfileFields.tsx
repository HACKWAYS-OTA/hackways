'use client';

import { Checkbox } from '@hackways/ui/checkbox';
import type { ProgramProfile } from '@hackways/api-client';

export function ProgramProfileFields({ values, email, minimumAge, ambassador, busy, onChange }: {
  values: ProgramProfile;
  email: string;
  minimumAge: number;
  ambassador: boolean;
  busy: boolean;
  onChange: <K extends keyof ProgramProfile>(key: K, value: ProgramProfile[K]) => void;
}) {
  return <div className="program-fields">
    <div className="program-contact-grid">
      <label htmlFor="program-name">Full name
        <input id="program-name" name="full_name" autoComplete="name" value={values.full_name} onChange={e => onChange('full_name', e.target.value)} required minLength={2} maxLength={120} />
      </label>
      <label htmlFor="program-email">Account email
        <input id="program-email" value={email} type="email" readOnly aria-describedby="program-email-note" />
        <span id="program-email-note">The verified email connected to your application.</span>
      </label>
    </div>
    <div className="program-field-pair">
      <label htmlFor="program-city">City or community
        <input id="program-city" name="city" autoComplete="address-level2" required minLength={2} maxLength={150} value={values.city} onChange={e => onChange('city', e.target.value)} />
      </label>
      <label htmlFor="program-age">Age
        <input id="program-age" name="age" type="number" inputMode="numeric" min={minimumAge} max={120} step={1} required value={values.age || ''} onChange={e => onChange('age', Number(e.target.value))} />
      </label>
    </div>
    <label htmlFor="program-phone">WhatsApp number{!ambassador && <span>(optional)</span>}
      <input id="program-phone" name="phone" type="tel" autoComplete="tel" maxLength={32} required={ambassador} placeholder="10-digit number" value={values.phone} onChange={e => onChange('phone', e.target.value)} aria-describedby="program-phone-note" />
      <span id="program-phone-note">Country code is optional for Indian numbers.</span>
    </label>
    <label htmlFor="program-profile">{ambassador ? 'LinkedIn profile' : 'Profile or portfolio'}{!ambassador && <span>(optional)</span>}
      <input id="program-profile" name="profile_url" type="url" inputMode="url" autoCapitalize="none" autoCorrect="off" placeholder="https://" maxLength={2048} required={ambassador} value={values.profile_url} onChange={e => onChange('profile_url', e.target.value)} aria-describedby={ambassador ? 'program-linkedin-note' : undefined} />
      {ambassador && <span id="program-linkedin-note">Share the full link to your profile on linkedin.com.</span>}
    </label>
    <label htmlFor="program-bio">A short introduction{!ambassador && <span>(optional)</span>}
      <textarea id="program-bio" name="biography" rows={4} minLength={ambassador ? 10 : undefined} required={ambassador} maxLength={2000} value={values.biography} onChange={e => onChange('biography', e.target.value)} />
      <span>Your interests, what you do, and the community you would bring together.</span>
    </label>
    <label className="program-check"><Checkbox disabled={busy} required checked={values.age_confirmed} onCheckedChange={checked => onChange('age_confirmed', checked)} />I confirm my details are accurate and I meet the program requirements.</label>
  </div>;
}
