import Link from "next/link";
import type { RegistrationStatus } from "@/types/api";
import { SectionJumpButton } from "@hackways/ui/section-jump";
import { requireEventPermission } from "@/lib/management";
import { pageOffset, privateApi, requireUser } from "@/lib/account";
import { dateLabel, registrationLabels } from "@/lib/journeys";
import { queryPath } from "@/lib/dashboard";
import { site } from "@/lib/site";
import { Icon } from "@/components/ui/Icon";
import { InviteAttendeesForm } from "@/components/management/InviteAttendeesForm";
import "./invitations.css";
type Invitation = {
  recipient_email: string;
  status: string;
  created_at: string;
  sent_at: string | null;
  registration_code: string | null;
  registration_status: RegistrationStatus | null;
  checked_in_at: string | null;
  guest_name: string;
};
type Summary = {
  total: number;
  sent: number;
  queued: number;
  failed: number;
  responded: number;
  awaiting: number;
};
export const metadata = { title: "Invitations" };
export default async function Invitations({
  params,
  searchParams,
}: {
  params: { id: string };
  searchParams: { offset?: string; q?: string; status?: string };
}) {
  const [event, user] = await Promise.all([
    requireEventPermission(params.id, 'guests.manage').then(result => result.data),
    requireUser(`/events/${params.id}/invitations`),
  ]);
  const offset = pageOffset(searchParams.offset),
    q =
      typeof searchParams.q === "string"
        ? searchParams.q.trim().slice(0, 200)
        : "",
    status = [
      "all",
      "awaiting",
      "responded",
      "sent",
      "failed",
      "queued",
    ].includes(searchParams.status || "")
      ? searchParams.status!
      : "all",
    path = `/events/${event.slug}/invitations`;
  const result = await privateApi<{ data: Invitation[]; summary: Summary }>(
    queryPath(`/events/${event.id}/invitations`, {
      limit: 26,
      offset,
      q,
      status,
    }),
  );
  const available =
    event.status === "published" &&
    Date.parse(event.starts_at) > Date.now() &&
    (!event.registration_closes_at ||
      Date.parse(event.registration_closes_at) > Date.now()) &&
    user.email_verified;
  const checkinOpen =
    event.status === "published" &&
    Date.now() >= Date.parse(event.starts_at) - 86400000 &&
    Date.now() < Date.parse(event.ends_at);
  const delivery: Record<string, string> = {
    pending: "Queued",
    processing: "Sending",
    sent: "Sent",
    failed: "Failed",
  };
  const filters = [
    ["all", "Everyone", result.summary.total],
    ["awaiting", "Awaiting response", result.summary.awaiting],
    ["responded", "Registered", result.summary.responded],
    ["queued", "Queued", result.summary.queued],
    ["failed", "Failed", result.summary.failed],
  ] as const;
  return (
    <>
      <div className="host-section-heading">
        <div>
          <h2>Invitations</h2>

        </div>
        <SectionJumpButton
          className="app-button app-button-primary"
          targetId="invite-composer"
        >
          <Icon name="add" size={17} />
          Invite attendees
        </SectionJumpButton>
      </div>
      <div className="invitation-list-tools">
        <nav aria-label="Invitation status">
          {filters.map(([value, label, count]) => (
            <Link
              key={value}
              href={queryPath(path, { status: value, q })}
              aria-current={status === value ? "page" : undefined}
            >
              {label}
              <span>{count}</span>
            </Link>
          ))}
        </nav>
        <form role="search" className="host-search">
          <input type="hidden" name="status" value={status} />
          <Icon name="search" size={17} />
          <input
            name="q"
            type="search"
            defaultValue={q}
            maxLength={200}
            aria-label="Search invitations by name or email"
            placeholder="Name or email"
          />
          <button>Search</button>
        </form>
      </div>
      <section
        className="invitation-history invitation-history-first"
        aria-label="Invited people"
      >
        {result.data.length ? (
          <div className="invitation-table-wrap">
            <table>
              <caption className="visually-hidden">
                Invitation responses and email delivery
              </caption>
              <thead>
                <tr>
                  <th scope="col">Person</th>
                  <th scope="col">Response</th>
                  <th scope="col">Email</th>
                  <th scope="col">Invited</th>
                  <th scope="col">
                    <span className="visually-hidden">Actions</span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {result.data.slice(0, 25).map((invitation) => (
                  <tr key={invitation.recipient_email}>
                    <td>
                      <div className="invited-person">
                        <span className="host-guest-avatar" aria-hidden="true">
                          {(invitation.guest_name || invitation.recipient_email)
                            .slice(0, 1)
                            .toUpperCase()}
                        </span>
                        <div>
                          {invitation.guest_name && (
                            <strong>{invitation.guest_name}</strong>
                          )}
                          <span>{invitation.recipient_email}</span>
                        </div>
                      </div>
                    </td>
                    <td>
                      <span
                        className={
                          invitation.registration_status
                            ? `journey-status status-${invitation.registration_status}`
                            : "invitation-awaiting"
                        }
                      >
                        {invitation.checked_in_at
                          ? "Checked in"
                          : invitation.registration_status
                            ? registrationLabels[
                                invitation.registration_status
                              ] || invitation.registration_status
                            : "Awaiting response"}
                      </span>
                    </td>
                    <td>{delivery[invitation.status] || invitation.status}</td>
                    <td>{dateLabel(invitation.created_at, event.timezone)}</td>
                    <td>
                      {invitation.registration_code ? (
                        <div className="invitation-row-actions">
                          <Link
                            className="quiet-link"
                            href={queryPath(`/events/${event.slug}/guests`, {
                              guest: invitation.registration_code,
                            })}
                          >
                            Guest details
                            <Icon name="arrow" size={14} />
                          </Link>
                          {invitation.registration_status === "confirmed" &&
                            !invitation.checked_in_at &&
                            checkinOpen && (
                              <Link
                                className="app-button app-button-outline"
                                href={queryPath(
                                  `/events/${event.slug}/check-in`,
                                  {
                                    guest: invitation.registration_code,
                                    checkin: "1",
                                  },
                                )}
                              >
                                Check in
                              </Link>
                            )}
                        </div>
                      ) : (
                        <span className="host-muted">Not registered</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="invitation-empty">
            <Icon name="people" size={28} />
            <h3>
              {result.summary.total
                ? "No invitations match this view"
                : "Invite your first attendees"}
            </h3>
            <p>
              {result.summary.total
                ? "Try a different name, email or response filter."
                : "Add email addresses below. Guests receive a link to accept and complete your registration."}
            </p>
          </div>
        )}
        {(offset > 0 || result.data.length > 25) && (
          <nav className="journey-pagination" aria-label="Invitation pages">
            {offset > 0 && (
              <Link
                className="app-button app-button-outline"
                href={queryPath(path, {
                  q,
                  status,
                  offset: Math.max(0, offset - 25),
                })}
              >
                Previous
              </Link>
            )}
            {result.data.length > 25 && (
              <Link
                className="app-button app-button-outline"
                href={queryPath(path, { q, status, offset: offset + 25 })}
              >
                Next
              </Link>
            )}
          </nav>
        )}
        <p className="field-help">
          Sent means accepted by the email provider. Attendance is confirmed
          only after registration.
        </p>
      </section>
      <section id="invite-composer" tabIndex={-1} className="invite-composer">
        <header>
          <h2>Invite people</h2>
          <p>
            They’ll receive an invitation to complete your registration
            questions.
          </p>
        </header>
        <div className="invite-layout">
          <div>
            {!available && (
              <p className="journey-notice">
                {!user.email_verified ? (
                  <>
                    Verify your email in{" "}
                    <a href={`${site.app}/account`}>your account</a> before
                    sending invitations.
                  </>
                ) : (
                  <>
                    Publish an upcoming event with registration open before
                    sending invitations.{" "}
                    <Link href={`/events/${event.slug}/settings`}>
                      Event settings
                    </Link>
                  </>
                )}
              </p>
            )}
            <InviteAttendeesForm eventId={event.id} available={available} />
          </div>
          <aside className="invite-note">
            <h3>What guests receive</h3>
            <p>
              An invitation with the event details, your message and an Accept
              invitation link.
            </p>
            <p>
              Already invited or registered guests are skipped. Approval and
              capacity settings still apply.
            </p>
            <Link
              className="quiet-link"
              href={`/events/${event.slug}/registration/questions`}
            >
              Edit registration questions
              <Icon name="arrow" size={15} />
            </Link>
          </aside>
        </div>
      </section>
    </>
  );
}
