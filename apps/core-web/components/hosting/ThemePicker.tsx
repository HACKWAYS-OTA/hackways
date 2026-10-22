"use client";
import { Select } from '@hackways/ui/select';
import { RadioGroup } from '@hackways/ui/radio-group';
import { eventFonts } from '@hackways/ui/event-fonts';
import { useRef } from 'react';
import { eventThemes, appearanceStyle, type Appearance } from '@/lib/appearance';
import { Icon } from '@/components/ui/Icon';
import { ColourPicker } from './ColourPicker';
import './appearance.css';

export function ThemePicker({ value, onChange, disabled }: {
  value: Appearance; onChange: (value: Appearance) => void; disabled: boolean;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const label = eventThemes.find(theme => theme.value === value.theme)!.label;
  function shuffle() {
    const random = crypto.getRandomValues(new Uint8Array(1))[0];
    const current = eventThemes.findIndex(theme => theme.value === value.theme);
    onChange({ ...value, theme: eventThemes[(current + 1 + random % (eventThemes.length - 1)) % eventThemes.length].value });
  }
  return <div className="theme-controls">
    <button type="button" className="theme-trigger" disabled={disabled} onClick={() => dialog.current?.showModal()} aria-haspopup="dialog">
      <span className={`theme-tile theme-${value.theme}`} style={appearanceStyle(value)} aria-hidden="true"><i /><b /></span>
      <span><small>Theme</small>{label}</span><Icon name="down" size={16} />
    </button>
    <button type="button" className="theme-shuffle" aria-label="Try another theme" disabled={disabled} onClick={shuffle}><Icon name="reload" size={18} /></button>
    <dialog ref={dialog} className="theme-sheet" aria-labelledby="theme-title" onClick={event => { if (event.target === dialog.current) dialog.current.close(); }}>
      <div className="theme-sheet-inner">
        <header><div><h2 id="theme-title">Make it feel like your event.</h2><p>Choose a theme, then give it your own touch.</p></div>
          <button type="button" className="icon-button" aria-label="Close theme picker" onClick={() => dialog.current?.close()}><Icon name="close" size={20} /></button>
        </header>
        <RadioGroup label="Event theme" className="theme-options" value={value.theme} disabled={disabled}
          onValueChange={theme => onChange({ ...value, theme })}
          options={eventThemes.map(theme => ({ value: theme.value, label: <><span className={`theme-tile theme-${theme.value}`} style={appearanceStyle(value)} aria-hidden="true"><i /><b /></span><span>{theme.label}</span></> }))} />
        <div className="theme-adjustments">
          <ColourPicker value={value.theme_color} onChange={theme_color => onChange({ ...value, theme_color })} />
          <fieldset className="theme-segment"><legend>Style</legend>
            <RadioGroup label="Event style" value={value.theme_style} disabled={disabled} options={[{ value: 'soft', label: 'Soft' }, { value: 'crisp', label: 'Crisp' }] as const} onValueChange={theme_style => onChange({ ...value, theme_style })} />
          </fieldset>
          <div className="theme-font"><label htmlFor="event-font">Font</label>
            <Select id="event-font" label="Event font" value={value.theme_font} disabled={disabled} options={eventFonts} onValueChange={theme_font => onChange({ ...value, theme_font })} />
          </div>
          <fieldset className="theme-segment"><legend>Display</legend>
            <RadioGroup label="Event display" value={value.theme_display} disabled={disabled} options={[{ value: 'light', label: 'Light' }, { value: 'dark', label: 'Dark' }] as const} onValueChange={theme_display => onChange({ ...value, theme_display })} />
          </fieldset>
        </div>
        <footer><p>Preview changes here, then save the event.</p><button type="button" className="app-button app-button-primary" onClick={() => dialog.current?.close()}>Done</button></footer>
      </div>
    </dialog>
  </div>;
}
