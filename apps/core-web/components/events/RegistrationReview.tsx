import type { User, RegistrationAnswer } from "@/types/api";
import type { RegistrationStep } from "@/lib/registration-form";
export function RegistrationReview({
  user,
  name,
  email,
  sections,
  answers,
  loading,
  go,
  approval,
}: {
  user: User | null;
  name: string;
  email: string;
  sections: RegistrationStep[];
  answers: Record<string, RegistrationAnswer>;
  loading: boolean;
  go: (step: number) => void;
  approval: boolean;
}) {
  return (
    <div className="registration-review">
      <p>Check your answers before sending them to the organizer.</p>
      <dl className="registration-review-identity">
        <div>
          <dt>Name</dt>
          <dd>{name.trim() || "Not provided"}</dd>
        </div>
        <div>
          <dt>Email</dt>
          <dd>{user?.email || email}</dd>
        </div>
      </dl>
      {sections.map((section, index) => (
        <section key={index}>
          <header>
            <h3>{section.title}</h3>
            <button
              className="quiet-link"
              type="button"
              disabled={loading}
              onClick={() => go(index + 1)}
            >
              Edit answers
            </button>
          </header>
          <dl>
            {section.questions.map((q) => (
              <div key={q.public_code}>
                <dt>{q.label}</dt>
                <dd>
                  {q.kind === "checkbox"
                    ? answers[q.public_code]?.checked
                      ? "Confirmed"
                      : "Not selected"
                    : q.kind === "multiselect"
                      ? answers[q.public_code]?.selected?.join(", ") ||
                        "Not provided"
                      : answers[q.public_code]?.text || "Not provided"}
                </dd>
              </div>
            ))}
          </dl>
        </section>
      ))}
      <p className="registration-note">
        {approval
          ? "The organizer must approve your registration before your place is confirmed."
          : "Availability is checked when you submit."}
      </p>
    </div>
  );
}
