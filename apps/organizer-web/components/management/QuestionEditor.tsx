"use client";
import { Checkbox } from "@hackways/ui/checkbox";
import { RadioGroup } from "@hackways/ui/radio-group";
import type {
  RegistrationQuestion,
  QuestionKind,
  QuestionValidation,
} from "@/types/api";
export const questionKinds: [QuestionKind, string][] = [
  ["text", "Short answer"],
  ["textarea", "Long answer"],
  ["email", "Email"],
  ["phone", "Phone"],
  ["number", "Number"],
  ["url", "Website"],
  ["date", "Date"],
  ["time", "Time"],
  ["select", "Choose one"],
  ["multiselect", "Choose several"],
  ["checkbox", "Confirmation"],
];
export function QuestionEditor({
  question: q,
  disabled,
  update,
}: {
  question: RegistrationQuestion;
  disabled: boolean;
  update: (patch: Partial<RegistrationQuestion>) => void;
}) {
  const validation = q.validation || {},
    written = ![
      "section",
      "checkbox",
      "select",
      "multiselect",
      "number",
      "date",
      "time",
    ].includes(q.kind);
  function rule(patch: Partial<QuestionValidation>) {
    update({ validation: { ...validation, ...patch } });
  }
  const numberValue = (value: string) =>
    value === "" ? undefined : Number(value);
  return (
    <fieldset className="question-detail-fields" disabled={disabled}>
      <legend className="visually-hidden">
        {q.kind === "section" ? "Section details" : "Question details"}
      </legend>
      <div className="field">
        <label htmlFor="question-label">
          {q.kind === "section" ? "Section title" : "Question"}
        </label>
        <input
          id="question-label"
          className="app-input"
          value={q.label}
          maxLength={180}
          onChange={(event) => update({ label: event.target.value })}
          placeholder={
            q.kind === "section"
              ? "About your team"
              : "What would you like to ask?"
          }
        />
      </div>
      <div className="field">
        <label htmlFor="question-description">
          {q.kind === "section" ? "Section introduction" : "Help text"}{" "}
          <span>(optional)</span>
        </label>
        <textarea
          id="question-description"
          className="app-input"
          rows={2}
          maxLength={500}
          value={q.description}
          onChange={(event) => update({ description: event.target.value })}
        />
      </div>
      {q.kind !== "section" && (
        <>
          <fieldset className="hw-choice-field">
            <legend>Answer type</legend>
            <RadioGroup
              label="Answer type"
              name="question-type"
              value={q.kind}
              disabled={disabled}
              options={questionKinds.map(([value, label]) => ({
                value,
                label,
              }))}
              onValueChange={(kind) =>
                update({
                  kind,
                  options: ["select", "multiselect"].includes(kind)
                    ? q.options.length
                      ? q.options
                      : ["", ""]
                    : [],
                  validation: {},
                  placeholder: "",
                })
              }
            />
          </fieldset>
          {["select", "multiselect"].includes(q.kind) ? (
            <div className="field">
              <label htmlFor="question-options">Choices, one per line</label>
              <textarea
                id="question-options"
                className="app-input"
                value={q.options.join("\n")}
                maxLength={2020}
                rows={5}
                onChange={(event) =>
                  update({ options: event.target.value.split("\n") })
                }
              />
              <p className="field-help">
                2-20 different choices, up to 100 characters each.
              </p>
            </div>
          ) : (
            q.kind !== "checkbox" && (
              <div className="field">
                <label htmlFor="question-placeholder">
                  Input hint <span>(optional)</span>
                </label>
                <input
                  id="question-placeholder"
                  className="app-input"
                  maxLength={180}
                  value={q.placeholder || ""}
                  onChange={(event) =>
                    update({ placeholder: event.target.value })
                  }
                />
              </div>
            )
          )}
          <label className="question-required">
            <Checkbox
              checked={q.required}
              disabled={disabled}
              onCheckedChange={(required) => update({ required })}
            />
            <span>Required answer</span>
          </label>
          {(written || q.kind === "number") && (
            <section className="question-validation">
              <h3>Answer validation</h3>
              <div className="form-pair">
                {written ? (
                  <>
                    <div className="field">
                      <label htmlFor="question-min-length">
                        Minimum characters
                      </label>
                      <input
                        id="question-min-length"
                        className="app-input"
                        type="number"
                        min={0}
                        max={2000}
                        value={validation.min_length ?? ""}
                        onChange={(event) =>
                          rule({ min_length: numberValue(event.target.value) })
                        }
                      />
                    </div>
                    <div className="field">
                      <label htmlFor="question-max-length">
                        Maximum characters
                      </label>
                      <input
                        id="question-max-length"
                        className="app-input"
                        type="number"
                        min={1}
                        max={2000}
                        value={validation.max_length ?? ""}
                        placeholder="2000"
                        onChange={(event) =>
                          rule({ max_length: numberValue(event.target.value) })
                        }
                      />
                    </div>
                  </>
                ) : (
                  <>
                    <div className="field">
                      <label htmlFor="question-min">Minimum number</label>
                      <input
                        id="question-min"
                        className="app-input"
                        type="number"
                        step="any"
                        value={validation.min ?? ""}
                        onChange={(event) =>
                          rule({ min: numberValue(event.target.value) })
                        }
                      />
                    </div>
                    <div className="field">
                      <label htmlFor="question-max">Maximum number</label>
                      <input
                        id="question-max"
                        className="app-input"
                        type="number"
                        step="any"
                        value={validation.max ?? ""}
                        onChange={(event) =>
                          rule({ max: numberValue(event.target.value) })
                        }
                      />
                    </div>
                  </>
                )}
              </div>
              {q.kind === "number" && (
                <label className="question-required">
                  <Checkbox
                    checked={!!validation.integer}
                    disabled={disabled}
                    onCheckedChange={(integer) => rule({ integer })}
                  />
                  <span>Whole numbers only</span>
                </label>
              )}
              <div className="field">
                <label htmlFor="question-validation-message">
                  Custom validation message <span>(optional)</span>
                </label>
                <input
                  id="question-validation-message"
                  className="app-input"
                  maxLength={200}
                  value={validation.message || ""}
                  onChange={(event) => rule({ message: event.target.value })}
                />
              </div>
            </section>
          )}
        </>
      )}
    </fieldset>
  );
}
