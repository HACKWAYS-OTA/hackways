import { Checkbox } from "@hackways/ui/checkbox";
import type { RegistrationFormSettings } from "@/types/api";
export function RegistrationPageSettings({
  settings,
  setting,
  nameRequired,
  setNameRequired,
  disabled,
}: {
  settings: RegistrationFormSettings;
  setting: (patch: Partial<RegistrationFormSettings>) => void;
  nameRequired: boolean;
  setNameRequired: (value: boolean) => void;
  disabled: boolean;
}) {
  return (
    <fieldset className="registration-settings" disabled={disabled}>
      <legend>Registration page</legend>
      <div className="form-pair">
        <div className="field">
          <label htmlFor="form-title">
            Page title <span>(optional)</span>
          </label>
          <input
            id="form-title"
            className="app-input"
            value={settings.title || ""}
            maxLength={180}
            placeholder="Register for this event"
            onChange={(event) => setting({ title: event.target.value })}
          />
        </div>
        <div className="field">
          <label htmlFor="form-submit">
            Submit button label <span>(optional)</span>
          </label>
          <input
            id="form-submit"
            className="app-input"
            value={settings.submit_label || ""}
            maxLength={40}
            placeholder="Use the event default"
            onChange={(event) => setting({ submit_label: event.target.value })}
          />
        </div>
      </div>
      <div className="field">
        <label htmlFor="form-introduction">
          Introduction <span>(optional)</span>
        </label>
        <textarea
          id="form-introduction"
          className="app-input"
          rows={2}
          maxLength={1000}
          value={settings.description || ""}
          onChange={(event) => setting({ description: event.target.value })}
        />
      </div>
      <div className="field">
        <label htmlFor="form-confirmation">
          Message after submission <span>(optional)</span>
        </label>
        <input
          id="form-confirmation"
          className="app-input"
          maxLength={500}
          value={settings.success_message || ""}
          onChange={(event) => setting({ success_message: event.target.value })}
        />
      </div>
      <div className="registration-identity-row">
        <span>Email address is always required.</span>
        <label className="question-required">
          <Checkbox
            checked={nameRequired}
            disabled={disabled}
            onCheckedChange={setNameRequired}
          />
          <span>Require a name</span>
        </label>
      </div>
    </fieldset>
  );
}
