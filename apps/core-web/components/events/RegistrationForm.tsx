"use client";
import { useEffect, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { scrollToSection } from "@hackways/ui/section-jump";
import type {
  Event,
  Registration,
  User,
  RegistrationFormConfig,
  RegistrationAnswer,
} from "@/types/api";
import { Icon } from "@/components/ui/Icon";
import { fetchApi, ApiClientError } from "@/lib/api";
import {
  eventDate,
  eventPath,
  eventZone,
  registrationState,
} from "@/lib/events";
import { registrationSteps, answerError } from "@/lib/registration-form";
import { recordAnalyticsEvent } from '@hackways/ui/analytics';
import { RegistrationIdentityStep } from "./RegistrationIdentityStep";
import { RegistrationReview } from "./RegistrationReview";
import { RegistrationQuestions } from "./RegistrationQuestions";
import "./registration-steps.css";
export function RegistrationForm({
  event,
  user,
  initialForm,
  invited = false,
  invitationHint = false,
  existingCode,
}: {
  event: Event;
  user: User | null;
  initialForm: RegistrationFormConfig;
  invited?: boolean;
  invitationHint?: boolean;
  existingCode?: string | null;
}) {
  const router = useRouter(),
    submitting = useRef(false),
    attempt = useRef<{ payload: string; key: string } | null>(null),
    feedback = useRef<HTMLDivElement>(null);
  const [form, setForm] = useState(initialForm),
    [answers, setAnswers] = useState<Record<string, RegistrationAnswer>>({}),
    [step, setStep] = useState(0),
    [name, setName] = useState(user?.display_name || ""),
    [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false),
    [error, setError] = useState(""),
    [fieldErrors, setFieldErrors] = useState<Record<string, string>>({}),
    [needsReload, setNeedsReload] = useState(false),
    [confirmed, setConfirmed] = useState<Registration | null>(null);
  const sections = registrationSteps(form.questions),
    reviewStep = sections.length + 1,
    labels = [
      "Your details",
      ...sections.map((section) => section.title),
      "Review",
    ],
    state = registrationState(event),
    joiningWaitlist = state.label === "Join waiting list";
  useEffect(() => {
    if (error || confirmed) feedback.current?.focus({ preventScroll: true });
  }, [error, confirmed]);
  useEffect(() => {
    if (
      confirmed ||
      (!Object.keys(answers).length && name === (user?.display_name || "") && !email)
    )
      return;
    const leave = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", leave);
    return () => window.removeEventListener("beforeunload", leave);
  }, [answers, name, email, confirmed, user?.display_name]);
  function go(next: number) {
    setStep(next);
    recordAnalyticsEvent("registration_step", next);
    setFieldErrors({});
    setError("");
    requestAnimationFrame(() => scrollToSection("registration-step"));
  }
  function validate(index: number): Record<string, string> {
    const errors: Record<string, string> = {};
    if (index === 0) {
      if ((form.name_required || !!user) && !name.trim())
        errors.name = "Your name is required.";
      if (!user && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()))
        errors.email = "Enter a valid email address.";
    } else if (index <= sections.length) {
      for (const q of sections[index - 1].questions) {
        const message = answerError(q, answers[q.public_code]),
          input = document.getElementById(`answer-${q.public_code}`);
        if (message) errors[q.public_code] = message;
        else if (input instanceof HTMLInputElement && input.validity.badInput)
          errors[q.public_code] = "Enter a valid answer.";
      }
    }
    return errors;
  }
  async function reload() {
    const result = await fetchApi<{ data: RegistrationFormConfig }>(
      `/events/by-slug/${encodeURIComponent(event.slug)}/registration-form`,
      { cache: "no-store" },
    );
    setAnswers((current) => {
      const compatible: Record<string, RegistrationAnswer> = {};
      for (const question of result.data.questions) {
        const previous = form.questions.find(
            (item) => item.public_code === question.public_code,
          ),
          answer = current[question.public_code];
        if (!answer || previous?.kind !== question.kind) continue;
        compatible[question.public_code] =
          question.kind === "multiselect"
            ? {
                ...answer,
                selected: answer.selected?.filter((value) =>
                  question.options.includes(value),
                ),
              }
            : question.kind === "select" &&
                !question.options.includes(answer.text)
              ? { code: question.public_code, text: "", checked: false }
              : answer;
      }
      return compatible;
    });
    setForm(result.data);
    setNeedsReload(false);
    setStep(0);
  }
  async function submit(action: FormEvent<HTMLFormElement>) {
    action.preventDefault();
    if (submitting.current || state.disabled || needsReload) return;
    const currentErrors = validate(step);
    if (Object.keys(currentErrors).length) {
      setFieldErrors(currentErrors);
      setError("Check the highlighted answers before continuing.");
      return;
    }
    if (step < reviewStep) {
      go(step + 1);
      return;
    }
    for (let index = 0; index < reviewStep; index++) {
      const errors = validate(index);
      if (Object.keys(errors).length) {
        setStep(index);
        setFieldErrors(errors);
        setError("Review the highlighted answers before submitting.");
        return;
      }
    }
    submitting.current = true;
    setLoading(true);
    setError("");
    try {
      const payload = JSON.stringify({
        event_id: event.id,
        guest_name: name.trim(),
        guest_email: user?.email || email.trim(),
        form_revision: form.revision,
        answers: form.questions
          .filter((q) => q.kind !== "section")
          .map(
            (q) =>
              answers[q.public_code] || {
                code: q.public_code,
                text: "",
                checked: false,
              },
          ),
        ...(joiningWaitlist ? { waitlist: true } : {}),
      });
      if (attempt.current?.payload !== payload)
        attempt.current = { payload, key: crypto.randomUUID() };
      const response = await fetchApi<{ data: Registration }>("/rsvps", {
        method: "POST",
        headers: { "Idempotency-Key": attempt.current.key },
        body: payload,
      });
      if (
        !response.data?.id ||
        response.data.event_id !== event.id ||
        !["confirmed", "pending_approval", "waitlisted"].includes(
          response.data.status,
        )
      )
        throw new Error(
          "The registration could not be confirmed. Check My events before retrying.",
        );
      setConfirmed(response.data);
      recordAnalyticsEvent("registration_complete");
      router.refresh();
    } catch (failure) {
      setError(
        failure instanceof Error
          ? failure.message
          : "Registration could not be saved. Your answers are still here.",
      );
      if (
        failure instanceof ApiClientError &&
        failure.code === "FORM_CHANGED"
      ) {
        setNeedsReload(true);
        try {
          await reload();
          setError(
            "The organizer updated the registration. Your answers are kept. Review each section before submitting again.",
          );
        } catch {
          setError(
            "The registration changed. Load the updated questions to continue. Your answers are kept.",
          );
        }
      } else if (
        failure instanceof ApiClientError &&
        failure.code === "INVALID_ANSWERS"
      ) {
        const index = sections.findIndex((section) =>
          section.questions.some((q) => failure.message.includes(q.label)),
        );
        if (index >= 0) setStep(index + 1);
      }
    } finally {
      submitting.current = false;
      setLoading(false);
    }
  }
  const heading = (
    <header className="event-detail-heading">
      <h2>
        {confirmed
          ? confirmed.status === "confirmed"
            ? "You’re registered"
            : confirmed.status === "waitlisted"
              ? "You’re on the waiting list"
              : "Registration submitted"
          : form.settings?.title ||
            (invited ? "Accept your invitation" : "Register for this event")}
      </h2>
      <p className="registration-event-date">
        {eventDate(event, {
          month: "long",
          day: "numeric",
          year: "numeric",
          hour: "numeric",
          minute: "2-digit",
        })}{" "}
        {eventZone(event)}
      </p>
    </header>
  );
  if (confirmed)
    return (
      <>
        {heading}
        <div
          className="registration-success registration-page-success"
          ref={feedback}
          tabIndex={-1}
          role="status"
        >
          <Icon name="check" size={30} />
          <h2>
            {confirmed.status === "pending_approval"
              ? "Awaiting organizer approval"
              : confirmed.status === "waitlisted"
                ? "Waiting for a place"
                : "Your place is confirmed"}
          </h2>
          <p>
            {confirmed.status === "pending_approval"
              ? "The organizer will review your answers. A place has not been reserved yet."
              : confirmed.status === "waitlisted"
                ? "No place is reserved yet. Check My events for an offer."
                : "Your QR pass is available with your registration."}
          </p>
          {form.settings?.success_message && (
            <p>{form.settings.success_message}</p>
          )}
          <Link
            className="app-button app-button-primary"
            href={`/my-events/${confirmed.public_code}`}
          >
            View registration
            <Icon name="arrow" size={17} />
          </Link>
          <Link className="quiet-link" href={eventPath(event)}>
            Back to event
          </Link>
        </div>
      </>
    );
  if (existingCode)
    return (
      <>
        {heading}
        <section className="event-registration-panel">
          <h2>You have already registered</h2>
          <p>Open your registration to see its current status and pass.</p>
          <Link
            className="app-button app-button-primary"
            href={`/my-events/${existingCode}`}
          >
            View registration
          </Link>
        </section>
      </>
    );
  if (state.disabled)
    return (
      <>
        {heading}
        <section className="event-registration-panel">
          <h2>{state.label}</h2>
          <p>Registration is not available at the moment.</p>
          <Link className="quiet-link" href={eventPath(event)}>
            Back to event
          </Link>
        </section>
      </>
    );
  if (
    !user &&
    (invitationHint ||
      event.allow_guest === false ||
      event.registration_mode === "approval" ||
      joiningWaitlist)
  )
    return (
      <>
        {heading}
        <section className="event-registration-panel">
          <h2>
            {invitationHint
              ? "Sign in to accept your invitation"
              : "Sign in to register"}
          </h2>
          <p>
            {invitationHint
              ? "Use the email address that received the invitation."
              : "Sign in to keep your registration and pass together."}
          </p>
          <Link
            className="app-button app-button-primary"
            href={`/login?next=${encodeURIComponent(eventPath(event) + "/register" + (invitationHint ? "?invitation=1" : ""))}`}
          >
            Sign in or create an account
          </Link>
        </section>
      </>
    );
  if (invitationHint && user && !invited)
    return (
      <>
        {heading}
        <section className="event-registration-panel">
          <h2>Check your invitation email</h2>
          <p>
            {!user.email_verified
              ? "Verify your account email before accepting this invitation."
              : "This account has no invitation for this event. Use the email address the organizer invited, or register normally."}
          </p>
          <Link
            className="app-button app-button-outline"
            href={
              !user.email_verified ? "/account" : eventPath(event) + "/register"
            }
          >
            {!user.email_verified ? "Open your account" : "Register normally"}
          </Link>
        </section>
      </>
    );
  return (
    <>
      {heading}
      <form
        onSubmit={submit}
        className="registration-form event-page-form section-registration-form"
        noValidate
        aria-busy={loading}
      >
        {form.settings?.description && (
          <p className="registration-intro">{form.settings.description}</p>
        )}
        {invited && (
          <p className="registration-invited-note">
            <Icon name="check" size={17} />
            You’re invited. Complete these details to respond.
          </p>
        )}
        <nav className="registration-step-nav" aria-label="Registration steps">
          <ol>
            {labels.map((label, index) => (
              <li key={`${index}-${label}`}>
                <button
                  type="button"
                  aria-current={step === index ? "step" : undefined}
                  disabled={loading || index > step}
                  onClick={() => go(index)}
                >
                  <span aria-hidden="true">{index + 1}</span>
                  {label}
                </button>
              </li>
            ))}
          </ol>
        </nav>
        <div id="registration-step" className="registration-step" tabIndex={-1}>
          <p className="registration-step-count">
            Step {step + 1} of {labels.length}
          </p>
          <h2>{labels[step]}</h2>
          {error && (
            <div
              className="form-alert"
              role="alert"
              ref={feedback}
              tabIndex={-1}
            >
              <Icon name="alert" size={19} />
              <p>{error}</p>
            </div>
          )}
          {needsReload && (
            <button
              className="app-button app-button-outline"
              type="button"
              disabled={loading}
              onClick={async () => {
                setLoading(true);
                try {
                  await reload();
                } catch {
                  setError(
                    "The updated questions could not be loaded. Try again.",
                  );
                } finally {
                  setLoading(false);
                }
              }}
            >
              Load updated questions
            </button>
          )}
          {step === 0 ? (
            <RegistrationIdentityStep
              user={user}
              name={name}
              email={email}
              nameRequired={form.name_required}
              loading={loading}
              fieldErrors={fieldErrors}
              setName={setName}
              setEmail={setEmail}
            />
          ) : step <= sections.length ? (
            <>
              {sections[step - 1].description && (
                <p className="registration-section-description">
                  {sections[step - 1].description}
                </p>
              )}
              <RegistrationQuestions
                questions={sections[step - 1].questions}
                answers={answers}
                errors={fieldErrors}
                disabled={loading}
                onChange={(answer) => {
                  setAnswers((current) => ({
                    ...current,
                    [answer.code]: answer,
                  }));
                  setFieldErrors((current) => ({
                    ...current,
                    [answer.code]: "",
                  }));
                }}
              />
            </>
          ) : (
            <RegistrationReview
              user={user}
              name={name}
              email={email}
              sections={sections}
              answers={answers}
              loading={loading}
              go={go}
              approval={event.registration_mode === "approval"}
            />
          )}
        </div>
        <div className="registration-actions">
          {step > 0 ? (
            <button
              className="app-button app-button-outline"
              type="button"
              disabled={loading}
              onClick={() => go(step - 1)}
            >
              Back
            </button>
          ) : (
            <Link className="quiet-link" href={eventPath(event)}>
              Back to event
            </Link>
          )}
          <button
            className="app-button app-button-primary"
            disabled={loading || needsReload}
          >
            {loading
              ? "Submitting..."
              : step < reviewStep
                ? "Continue"
                : invited
                  ? "Accept invitation"
                  : form.settings?.submit_label ||
                    (joiningWaitlist
                      ? "Join waiting list"
                      : "Register")}
            <Icon name="arrow" size={17} />
          </button>
        </div>
      </form>
    </>
  );
}
