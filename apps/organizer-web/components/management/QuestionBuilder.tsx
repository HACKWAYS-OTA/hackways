"use client";
import { RegistrationPageSettings } from "./RegistrationPageSettings";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import type {
  RegistrationFormConfig,
  RegistrationQuestion,
  RegistrationFormSettings,
} from "@/types/api";
import { fetchApi, ApiClientError } from "@/lib/api";
import { Icon } from "@/components/ui/Icon";
import { QuestionEditor, questionKinds } from "./QuestionEditor";
import "./question-builder.css";
type Draft = RegistrationQuestion & { localKey: string };
const drafts = (questions: RegistrationQuestion[]) =>
  questions.map((q, i) => ({ ...q, localKey: q.public_code || `new-${i}` }));
export function QuestionBuilder({
  eventId,
  initial,
  editable,
}: {
  eventId: string;
  initial: RegistrationFormConfig;
  editable: boolean;
}) {
  const router = useRouter(),
    sequence = useRef(initial.questions.length),
    lock = useRef(false);
  const [questions, setQuestions] = useState<Draft[]>(() =>
      drafts(initial.questions),
    ),
    [settings, setSettings] = useState<RegistrationFormSettings>(
      initial.settings || {},
    );
  const [nameRequired, setNameRequired] = useState(initial.name_required),
    [revision, setRevision] = useState(initial.revision),
    [selected, setSelected] = useState<string | null>(
      drafts(initial.questions)[0]?.localKey || null,
    ),
    [remove, setRemove] = useState(false);
  const [baseline, setBaseline] = useState(
    JSON.stringify({
      name: initial.name_required,
      questions: drafts(initial.questions),
      settings: initial.settings || {},
    }),
  );
  const [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [saved, setSaved] = useState(false),
    [conflict, setConflict] = useState(false);
  const dirty =
      JSON.stringify({ name: nameRequired, questions, settings }) !== baseline,
    selectedIndex = questions.findIndex((q) => q.localKey === selected),
    current = questions[selectedIndex];
  const questionCount = questions.filter((q) => q.kind !== "section").length,
    sectionCount = questions.length - questionCount;
  useEffect(() => {
    if (!dirty) return;
    const leave = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", leave);
    return () => window.removeEventListener("beforeunload", leave);
  }, [dirty]);
  function change(patch: Partial<RegistrationQuestion>) {
    setQuestions((items) =>
      items.map((q) => (q.localKey === selected ? { ...q, ...patch } : q)),
    );
    setSaved(false);
  }
  function add(section = false, copy = false) {
    const localKey = `new-${++sequence.current}`,
      q: Draft =
        copy && current
          ? {
              ...current,
              public_code: "",
              localKey,
              options: [...current.options],
              validation: { ...current.validation },
            }
          : {
              localKey,
              public_code: "",
              label: "",
              description: "",
              kind: section ? "section" : "text",
              required: false,
              options: [],
              validation: {},
              placeholder: "",
            };
    setQuestions((items) => [...items, q]);
    setSelected(localKey);
    setRemove(false);
    setSaved(false);
  }
  function move(by: number) {
    setQuestions((items) => {
      const next = [...items];
      [next[selectedIndex], next[selectedIndex + by]] = [
        next[selectedIndex + by],
        next[selectedIndex],
      ];
      return next;
    });
    setSaved(false);
  }
  function setting(patch: Partial<RegistrationFormSettings>) {
    setSettings((value) => ({ ...value, ...patch }));
    setSaved(false);
  }
  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (lock.current || !editable) return;
    const invalid = questions.find(
      (q) =>
        !q.label.trim() ||
        (["select", "multiselect"].includes(q.kind) &&
          (q.options.length < 2 ||
            q.options.some((option) => !option.trim()) ||
            new Set(q.options.map((option) => option.trim())).size !==
              q.options.length)),
    );
    if (invalid) {
      setSelected(invalid.localKey);
      setError(
        "Give every section and question a title, and every choice question at least two different options.",
      );
      return;
    }
    lock.current = true;
    setBusy(true);
    setError("");
    setSaved(false);
    setConflict(false);
    try {
      const result = await fetchApi<{ data: RegistrationFormConfig }>(
        `/events/${eventId}/registration-form`,
        {
          method: "PUT",
          body: JSON.stringify({
            revision,
            name_required: nameRequired,
            settings,
            questions: questions.map(({ localKey: _key, ...q }) => q),
          }),
        },
      );
      const next = drafts(result.data.questions);
      setQuestions(next);
      setRevision(result.data.revision);
      setNameRequired(result.data.name_required);
      setSettings(result.data.settings || {});
      setBaseline(
        JSON.stringify({
          name: result.data.name_required,
          questions: next,
          settings: result.data.settings || {},
        }),
      );
      setSelected(next[0]?.localKey || null);
      setSaved(true);
      router.refresh();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Your questions could not be saved. Your edits are still here.",
      );
      setConflict(err instanceof ApiClientError && err.code === "FORM_CHANGED");
    } finally {
      lock.current = false;
      setBusy(false);
    }
  }
  return (
    <form
      className="host-question-builder section-question-builder"
      onSubmit={save}
      aria-busy={busy}
    >
      <div className="host-section-heading">
        <div>
          <h2>Registration questions</h2>
        </div>
        <span className="host-question-limit">
          {questionCount} / 40 questions · {sectionCount} / 10 sections
        </span>
      </div>
      {!editable && (
        <p className="journey-notice">
          Questions are read-only after the event starts or is cancelled.
        </p>
      )}
      <RegistrationPageSettings
        settings={settings}
        setting={setting}
        nameRequired={nameRequired}
        setNameRequired={(required) => {
          setNameRequired(required);
          setSaved(false);
        }}
        disabled={!editable || busy}
      />
      <div className="question-builder-layout">
        <section
          className="question-outline"
          aria-label="Questions and sections"
        >
          <header>
            <h3>Sections & questions</h3>
          </header>
          <ol>
            {questions.map((q, index) => (
              <li
                key={q.localKey}
                className={q.kind === "section" ? "question-section-break" : ""}
              >
                <button
                  type="button"
                  aria-current={q.localKey === selected ? "true" : undefined}
                  onClick={() => {
                    setSelected(q.localKey);
                    setRemove(false);
                  }}
                >
                  <Icon
                    name={q.kind === "section" ? "add" : "edit"}
                    size={17}
                  />
                  <span>
                    {q.label ||
                      (q.kind === "section"
                        ? "Untitled section"
                        : "Untitled question")}
                    <small>
                      {q.kind === "section"
                        ? "New registration step"
                        : `${questionKinds.find(([kind]) => kind === q.kind)?.[1]}${q.required ? " · Required" : ""}`}
                    </small>
                  </span>
                  <span className="question-position">{index + 1}</span>
                </button>
              </li>
            ))}
          </ol>
          {!questions.length && (
            <p className="question-outline-empty">
              Start with a question, or add a section to introduce the next
              step.
            </p>
          )}
          <div className="question-add-actions">
            <button
              className="app-button app-button-outline"
              type="button"
              disabled={!editable || busy || questionCount >= 40}
              onClick={() => add()}
            >
              <Icon name="add" size={16} />
              Add question
            </button>
            <button
              className="quiet-link"
              type="button"
              disabled={!editable || busy || sectionCount >= 10}
              onClick={() => add(true)}
            >
              Add section break
            </button>
          </div>
        </section>
        <section
          className="question-editor-panel"
          aria-label="Edit selected question"
        >
          {current ? (
            <>
              <header>
                <h3>
                  {current.kind === "section"
                    ? "Section break"
                    : "Edit question"}
                </h3>
                <div>
                  <button
                    type="button"
                    className="icon-button"
                    disabled={!editable || busy || selectedIndex === 0}
                    aria-label="Move selected item up"
                    onClick={() => move(-1)}
                  >
                    <Icon name="down" className="host-arrow-up" size={18} />
                  </button>
                  <button
                    type="button"
                    className="icon-button"
                    disabled={
                      !editable ||
                      busy ||
                      selectedIndex === questions.length - 1
                    }
                    aria-label="Move selected item down"
                    onClick={() => move(1)}
                  >
                    <Icon name="down" size={18} />
                  </button>
                </div>
              </header>
              <QuestionEditor
                key={current.localKey}
                question={current}
                update={change}
                disabled={!editable || busy}
              />
              <footer>
                <button
                  type="button"
                  className="quiet-link"
                  disabled={
                    !editable ||
                    busy ||
                    (current.kind === "section"
                      ? sectionCount >= 10
                      : questionCount >= 40)
                  }
                  onClick={() => add(false, true)}
                >
                  Duplicate
                </button>
                <button
                  type="button"
                  className="quiet-link"
                  disabled={!editable || busy}
                  onClick={() => setRemove(true)}
                >
                  Remove{" "}
                  {current.kind === "section" ? "section break" : "question"}
                </button>
              </footer>
              {remove && (
                <div className="host-question-remove">
                  <p>
                    Remove this item from future registrations? Previous answers
                    are kept.
                  </p>
                  <button
                    type="button"
                    className="app-button app-button-outline"
                    onClick={() => {
                      setQuestions((items) =>
                        items.filter((q) => q.localKey !== selected),
                      );
                      setSelected(null);
                      setRemove(false);
                      setSaved(false);
                    }}
                  >
                    Remove
                  </button>
                  <button
                    type="button"
                    className="quiet-link"
                    onClick={() => setRemove(false)}
                  >
                    Keep it
                  </button>
                </div>
              )}
            </>
          ) : (
            <div className="question-editor-empty">
              <Icon name="edit" size={28} />
              <h3>Select a question or section</h3>
              <p>Edit labels, answer types and validation here.</p>
            </div>
          )}
        </section>
      </div>
      {error && (
        <div className="journey-error" role="alert">
          <p>{error}</p>
          {conflict && (
            <button
              className="quiet-link"
              type="button"
              onClick={() => router.refresh()}
            >
              Load the latest form (replaces your edits)
            </button>
          )}
        </div>
      )}
      {saved && (
        <p className="journey-notice" role="status">
          Registration saved. Guests will use the updated questions.
        </p>
      )}
      {editable && (
        <footer className="host-save">
          <span>
            {dirty
              ? "You have unsaved changes."
              : "Your registration is up to date."}
          </span>
          <button
            className="app-button app-button-primary"
            disabled={busy || !dirty || conflict}
          >
            {busy ? "Saving..." : "Save registration"}
          </button>
        </footer>
      )}
    </form>
  );
}
