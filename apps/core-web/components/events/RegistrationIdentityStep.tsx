import type { User } from "@/types/api";
export function RegistrationIdentityStep({
  user,
  name,
  email,
  nameRequired,
  loading,
  fieldErrors,
  setName,
  setEmail,
}: {
  user: User | null;
  name: string;
  email: string;
  nameRequired: boolean;
  loading: boolean;
  fieldErrors: Record<string, string>;
  setName: (value: string) => void;
  setEmail: (value: string) => void;
}) {
  return (
    <div className="registration-identity-step">
      <div className="field">
        <label htmlFor="registration-name">Your name{nameRequired || user ? " *" : " (optional)"}</label>
        <input id="registration-name" className="app-input" autoComplete="name"
          value={name} maxLength={200} required={nameRequired || !!user} disabled={loading}
          aria-invalid={!!fieldErrors.name} aria-describedby={fieldErrors.name ? "registration-name-error" : undefined}
          onChange={(event) => setName(event.target.value)} />
        {fieldErrors.name && <p id="registration-name-error" className="registration-field-error" role="alert">{fieldErrors.name}</p>}
      </div>
      {user ? <div className="field">
        <label htmlFor="registration-email">Email address</label>
        <input id="registration-email" className="app-input" type="email" value={user.email} readOnly aria-describedby="registration-email-help" />
        <p id="registration-email-help" className="field-help">Linked to your signed-in account. Your name change applies to this registration only.</p>
      </div> : <div className="field">
        <label htmlFor="registration-email">Email address *</label>
        <input id="registration-email" className="app-input" type="email" autoComplete="email"
          value={email} maxLength={254} disabled={loading} required
          aria-invalid={!!fieldErrors.email} aria-describedby={fieldErrors.email ? "registration-email-error" : "registration-email-help"}
          onChange={(event) => setEmail(event.target.value)} />
        {fieldErrors.email && <p id="registration-email-error" className="registration-field-error" role="alert">{fieldErrors.email}</p>}
        <p id="registration-email-help" className="field-help">Your registration updates will be sent here.</p>
      </div>}
    </div>
  );
}
