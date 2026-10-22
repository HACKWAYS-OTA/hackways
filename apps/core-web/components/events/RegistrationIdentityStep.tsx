import Link from "next/link";
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
      {user ? (
        <>
          <dl className="registration-review-identity">
            <div>
              <dt>Name</dt>
              <dd>{user.display_name}</dd>
            </div>
            <div>
              <dt>Email</dt>
              <dd>{user.email}</dd>
            </div>
          </dl>
          <p className="field-help">
            Use <Link href="/account">your account</Link> to update your name.
          </p>
          {fieldErrors.name && (
            <p className="registration-field-error" role="alert">
              {fieldErrors.name}
            </p>
          )}
        </>
      ) : (
        <>
          <div className="field">
            <label htmlFor="registration-name">
              Your name{nameRequired ? " *" : " (optional)"}
            </label>
            <input
              id="registration-name"
              className="app-input"
              autoComplete="name"
              value={name}
              maxLength={200}
              disabled={loading}
              aria-invalid={!!fieldErrors.name}
              aria-describedby={
                fieldErrors.name ? "registration-name-error" : undefined
              }
              onChange={(event) => setName(event.target.value)}
            />
            {fieldErrors.name && (
              <p
                id="registration-name-error"
                className="registration-field-error"
                role="alert"
              >
                {fieldErrors.name}
              </p>
            )}
          </div>
          <div className="field">
            <label htmlFor="registration-email">Email address *</label>
            <input
              id="registration-email"
              className="app-input"
              type="email"
              autoComplete="email"
              value={email}
              maxLength={254}
              disabled={loading}
              aria-invalid={!!fieldErrors.email}
              aria-describedby={
                fieldErrors.email ? "registration-email-error" : undefined
              }
              onChange={(event) => setEmail(event.target.value)}
            />
            {fieldErrors.email && (
              <p
                id="registration-email-error"
                className="registration-field-error"
                role="alert"
              >
                {fieldErrors.email}
              </p>
            )}
            <p className="field-help">
              Your registration updates will be sent here.
            </p>
          </div>
        </>
      )}
    </div>
  );
}
