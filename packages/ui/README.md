# Hackways shared UI controls

[Repository overview](../../README.md) · [Source exports](package.json) · [Architecture](../../docs/ARCHITECTURE.md)

## Package scope

The four independent Next.js apps consume `@hackways/ui` from this root package. It owns the checkbox, grouped radio, select and toggle implementations, theme styles and animation defaults. Business rules and persistence remain in Go.

## Install and use

Run `npm ci` from the repository root to install the committed lockfile. Use `npm install` only when intentionally updating dependencies and reviewing the resulting lockfile. npm workspaces link the shared source so edits apply to all apps without copying components. Run an app with `npm run dev --workspace @hackways/core-web` (or another app name).

Every app root already imports `@hackways/ui/styles.css`, wraps its contents in `UIProvider`, and transpiles `@hackways/ui`. Do not add per-page providers or reinstall dependencies for new forms.

```tsx
import { Checkbox } from '@hackways/ui/checkbox';
import { RadioGroup } from '@hackways/ui/radio-group';
import { Select } from '@hackways/ui/select';
import { Toggle } from '@hackways/ui/toggle';

<label><Checkbox name="notify" checked={notify} onCheckedChange={setNotify} />Notify guest</label>

<RadioGroup label="Event visibility" name="visibility" value={visibility}
  onValueChange={setVisibility} required disabled={busy}
  options={[{ value: 'public', label: 'Public' }, { value: 'private', label: 'Private' }] as const} />

<Select label="Guest status" name="status" value={status} onValueChange={setStatus}
  options={allowedStatuses} placeholder="Choose new status" required disabled={busy} />

<Toggle aria-label="Toggle bold" pressed={bold} onPressedChange={setBold} disabled={busy}>
  <BoldIcon />
</Toggle>
```

Supply real options from the caller. Radio groups accept `layout="stack"` and optional descriptions. Checkbox supports checked/defaultChecked, required, name/value, disabled, readOnly, indeterminate and accessible attributes. Radio `defaultValue` supports an initially uncontrolled selection. Radio/select option values are typed string unions. Existing native form names and required validation are preserved by Base UI's inputs; pass `disabled` explicitly when a surrounding native fieldset is disabled.

Toggle preserves Base UI's `pressed`, `defaultPressed`, `onPressedChange`, `value`, refs, native keyboard/button semantics and state-based class/style callbacks. `Toggle.Props`, `Toggle.State` and change-event types are available. It scales to 0.97 only while pressed, using a spring with stiffness 550/damping 30 (override via `transition`). Colors fade for 300ms; initial state is at rest, with no hover scale or press scale under reduced motion. Supplying `render` keeps Base UI behavior and delegates animation to that element. Use Toggle for independent pressed buttons; exclusive choices, tabs and switches retain their respective semantics. Toggle does not submit a form value.

## Accessibility, navigation, and visual defaults

Account links in core and management use `Avatar` from `@hackways/ui/avatar`. It displays an HTTPS account photo directly, falls back to initials on failure, and uses a Hugeicons account outline for anonymous visitors. The containing link supplies the accessible name. Root layouts resolve the session on the server; React request caching deduplicates that identity read within each render.

Authentication surfaces share `AuthLayout`, `AuthGoogleButton`, `AuthDivider`, `AuthPasswordInput` and `AuthSubmit` through `@hackways/ui/auth`. Import the components from that subpath and import `@hackways/ui/auth.css` once in the consuming app root. Core, management and admin consume this presentation; password reveal uses the shared Toggle and Hugeicons. Keep field labels, autocomplete, pending/error states and actual API submissions in the consuming form. Authentication, session validation and authorization remain in Go; do not add general auth business logic to this TypeScript package. See the [authentication surface guide](../../docs/design/AUTHENTICATION.md) for layout, accessibility and verification scope.

For in-page navigation, import `SectionJumpButton` or `scrollToSection` from `@hackways/ui/section-jump`. They move focus and scroll without changing the URL or browser history. Pass a `targetId`; section IDs do not require fragment links. Use clean routes for cross-page destinations. The button respects reduced motion and does not submit forms.

- Base UI 1.8 and Motion 14 work with the existing React 18 / Next.js 14 apps. This is a local implementation using those libraries, not `@animate-ui/react` (whose published 2.0 release requires React 19).
- Checkmarks pop; radio dots spring; dropdowns fade/scale with a moving highlight. Existing selections render at rest. Reduced motion fades without scales or a moving highlight.
- Hugeicons use existing subpath imports. No shadcn CLI scaffolding or Lucide dependency.
- Styles inherit the existing `--ink`, `--paper`, `--muted`, `--line`, `--soft` tokens and editor equivalents. Select copies these resolved values into its portal, and portals inside the current native dialog when present.
- Use subpath imports to keep unused controls out of app entry points. Add any future approved primitive, export and styles here once, then consume it from the apps.

## Verification and deployment

`npm run type-check --workspace @hackways/ui` and `npm run lint --workspace @hackways/ui` check every shared export, including components not yet used by an app. `npm test --workspace @hackways/core-web` (repeat for other apps) runs existing app type and lint checks. CI installs the root lockfile and checks the shared package. The four frontend Dockerfiles install the root locked workspace with `npm ci`, build the shared API client, and package each app independently with standalone output. They do not use the older per-app `--install-links` installation flow. Production images are built only in CI, never on the Azure VM.

### Check the package

```bash
npm ci
npm run type-check --workspace @hackways/ui
npm run lint --workspace @hackways/ui
```

Run affected app checks as well when shared behavior changes. Type and lint checks are not a substitute for keyboard, focus, reduced-motion, or browser-journey verification. This package is private and is not documented as a published npm package.
