import Link from 'next/link';

/** Useful public context stays server-rendered even when the live catalog is empty. */
export function DiscoverGuide() {
  return <section className="feed-container discovery-guide" aria-labelledby="discovery-guide-heading">
    <h2 id="discovery-guide-heading">Find a gathering that fits.</h2>
    <p>Hackways brings technology events and their practical details into one place. Browse upcoming public gatherings, search by an event name or topic, and open an event to read its schedule, format and participation requirements.</p>
    <div className="discovery-guide-columns">
      <section><h3>Before you register</h3><p>Check the organizer&apos;s description, time zone and venue or online joining details. Depending on the event, registration may confirm a place, enter a waiting list or send an application for organizer review. Opening a form does not reserve a place.</p><Link href="/help/joining">How event registration works</Link></section>
      <section><h3>Keep your community close</h3><p>A Space brings a community and its events together. Join one around an interest you share, explore an open program, or create an event for people you want to bring together. Upcoming public events appear here when organizers publish them.</p><nav aria-label="More ways to take part"><Link href="/spaces">Explore Spaces</Link><Link href="/programs">Browse programs</Link><Link href="/organizers">Guide to hosting</Link></nav></section>
    </div>
  </section>;
}
