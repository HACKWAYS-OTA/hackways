import { Checkbox } from "@hackways/ui/checkbox";
import { RadioGroup } from "@hackways/ui/radio-group";
import type { RegistrationQuestion, RegistrationAnswer } from "@/types/api";
export function RegistrationQuestions({
  questions,
  answers,
  onChange,
  disabled,
  errors = {},
}: {
  questions: RegistrationQuestion[];
  answers: Record<string, RegistrationAnswer>;
  onChange: (answer: RegistrationAnswer) => void;
  disabled: boolean;
  errors?: Record<string, string>;
}) {
  return (
    <div className="registration-custom-questions">
      {questions
        .filter((q) => q.kind !== "section")
        .map((q) => {
          const answer = answers[q.public_code] || {
              code: q.public_code,
              text: "",
              checked: false,
            },
            id = `answer-${q.public_code}`,
            helpId = `${id}-help`,
            errorId = `${id}-error`,
            error = errors[q.public_code],
            describedBy =
              [q.description ? helpId : "", error ? errorId : ""]
                .filter(Boolean)
                .join(" ") || undefined;
          const label = (
              <>
                {q.label}
                {q.required ? (
                  <span className="required-mark" aria-label="required">
                    {" "}
                    *
                  </span>
                ) : (
                  <span> (optional)</span>
                )}
              </>
            ),
            help = q.description && (
              <p id={helpId} className="field-help">
                {q.description}
              </p>
            ),
            feedback = error && (
              <p id={errorId} className="registration-field-error" role="alert">
                {error}
              </p>
            );
          if (q.kind === "select")
            return (
              <fieldset
                className="registration-answer-choice"
                key={id}
                aria-invalid={!!error}
              >
                <legend>{label}</legend>
                {help}
                <RadioGroup
                  label={q.label}
                  name={id}
                  options={q.options.map((option) => ({
                    value: option,
                    label: option,
                  }))}
                  value={answer.text}
                  disabled={disabled}
                  describedBy={describedBy}
                  layout="stack"
                  onValueChange={(text) =>
                    onChange({ code: q.public_code, text, checked: false })
                  }
                />
                {!q.required && answer.text && (
                  <button
                    type="button"
                    className="quiet-link"
                    disabled={disabled}
                    onClick={() =>
                      onChange({
                        code: q.public_code,
                        text: "",
                        checked: false,
                      })
                    }
                  >
                    Clear selection
                  </button>
                )}
                {feedback}
              </fieldset>
            );
          if (q.kind === "multiselect")
            return (
              <fieldset
                key={id}
                className="registration-answer-choice"
                aria-describedby={describedBy}
                aria-invalid={!!error}
              >
                <legend>{label}</legend>
                {help}
                <div className="registration-multiple-choices">
                  {q.options.map((option) => (
                    <label key={option}>
                      <Checkbox
                        checked={!!answer.selected?.includes(option)}
                        disabled={disabled}
                        onCheckedChange={(checked) =>
                          onChange({
                            code: q.public_code,
                            text: "",
                            checked: false,
                            selected: checked
                              ? [...(answer.selected || []), option]
                              : (answer.selected || []).filter(
                                  (value) => value !== option,
                                ),
                          })
                        }
                      />
                      <span>{option}</span>
                    </label>
                  ))}
                </div>
                {feedback}
              </fieldset>
            );
          if (q.kind === "checkbox")
            return (
              <div className="field" key={id}>
                <label className="registration-answer-check">
                  <Checkbox
                    id={id}
                    checked={answer.checked}
                    disabled={disabled}
                    aria-invalid={!!error}
                    aria-describedby={describedBy}
                    onCheckedChange={(checked) =>
                      onChange({ code: q.public_code, text: "", checked })
                    }
                  />
                  <span>{label}</span>
                </label>
                {help}
                {feedback}
              </div>
            );
          const type =
              q.kind === "url"
                ? "url"
                : q.kind === "phone"
                  ? "tel"
                  : ["email", "number", "date", "time"].includes(q.kind)
                    ? q.kind
                    : "text",
            v = q.validation?.enabled === false ? {} : q.validation || {};
          return (
            <div className="field" key={id}>
              <label htmlFor={id}>{label}</label>
              {help}
              {q.kind === "textarea" ? (
                <textarea
                  id={id}
                  className="app-input"
                  value={answer.text}
                  maxLength={v.max_length ?? 2000}
                  rows={4}
                  disabled={disabled}
                  placeholder={q.placeholder}
                  aria-invalid={!!error}
                  aria-describedby={describedBy}
                  onChange={(event) =>
                    onChange({
                      code: q.public_code,
                      text: event.target.value,
                      checked: false,
                    })
                  }
                />
              ) : (
                <input
                  id={id}
                  className="app-input"
                  type={type}
                  inputMode={
                    q.kind === "phone"
                      ? "tel"
                      : q.kind === "number"
                        ? "decimal"
                        : q.kind === "email"
                          ? "email"
                          : undefined
                  }
                  autoComplete={
                    q.kind === "email"
                      ? "email"
                      : q.kind === "phone"
                        ? "tel"
                        : undefined
                  }
                  value={answer.text}
                  min={q.kind === "number" ? v.min : undefined}
                  max={q.kind === "number" ? v.max : undefined}
                  step={
                    q.kind === "number" ? (v.integer ? 1 : "any") : undefined
                  }
                  maxLength={v.max_length ?? 2000}
                  disabled={disabled}
                  placeholder={q.placeholder}
                  aria-invalid={!!error}
                  aria-describedby={describedBy}
                  onChange={(event) =>
                    onChange({
                      code: q.public_code,
                      text: event.target.value,
                      checked: false,
                    })
                  }
                />
              )}
              {feedback}
            </div>
          );
        })}
    </div>
  );
}
