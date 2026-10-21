'use client';
const colours = [
  ['#ddd4e9', 'Lavender'], ['#d7e4ef', 'Blue'], ['#d7e8df', 'Sage'],
  ['#f3dfbc', 'Amber'], ['#efd5dd', 'Rose'], ['#ced9da', 'Slate'],
] as const;
export function ColourPicker({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return <div className="theme-colours"><span>Colour</span><div role="group" aria-label="Theme colour">{colours.map(([colour, label]) => <button type="button" key={colour} style={{ backgroundColor: colour }} aria-label={label} aria-pressed={value === colour} onClick={() => onChange(colour)} />)}</div></div>;
}
