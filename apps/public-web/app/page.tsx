import type { Metadata } from "next";
import Link from "next/link";
import { FAQ } from "@/components/site/FAQ";
import { TextAnimate } from "@/components/ui/text-animate";
import { EventStories } from "@/components/site/EventStories";
import { EventStorySymbol } from "@/components/site/EventStorySymbol";
import { InterestTiles } from "@/components/site/InterestTiles";
import { ContentSwitcher } from "@/components/site/ContentSwitcher";
import { DisclosureList } from "@/components/site/DisclosureList";
import { LandingNavigation } from "@/components/site/LandingNavigation";
import { site } from "@/lib/site";

export const metadata: Metadata = { alternates: { canonical: "/" } };

export default function HomePage() {
  return (
    <div className="landing-page">
      <div className="hero-discovery">
      <section className="landing-section reference-hero" aria-labelledby="hero-heading">
        <div className="site-container reference-hero-inner">
          <div className="hero-lead">
            <TextAnimate as="h1" id="hero-heading" className="hero-headline" animation="blurInUp" by="word" duration={0.9} delay={0.15} startOnView={false} once>
              {"Good things happen\nwhen people meet."}
            </TextAnimate>
            <p className="hero-description">Find your next tech event. Meet people who share your curiosity, or bring a community together around an idea.</p>
            <div className="hero-cta-group">
              <a className="button button-primary hero-cta" href={`${site.app}/discover`}>
                Explore events
                <span className="hero-cta-arrow" aria-hidden="true">
                  <svg width="20" height="20" viewBox="0 0 20 20" fill="none"><path d="M4 10h11m-5-5 5 5-5 5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" /></svg>
                </span>
              </a>
              <a className="button button-outline hero-secondary" href="/create">Host an event</a>
            </div>
          </div>
          <div className="hero-details">
            <div className="hero-benefits">
              <div><span className="hero-check" aria-hidden="true" /><h2>Find your people.</h2><p>Events for what you are curious about.</p></div>
              <div><span className="hero-check" aria-hidden="true" /><h2>Join your way.</h2><p>In person, online, or a little of both.</p></div>
              <div><span className="hero-check" aria-hidden="true" /><h2>Bring an idea.</h2><p>Give your community a place to start.</p></div>
            </div>
          </div>
        </div>
      </section>

      <section className="landing-section discovery-section" aria-labelledby="possibilities-heading">
        <div className="site-container possibilities">
          <div className="section-intro">
            <h2 id="possibilities-heading">A place for shared interests.</h2>
            <p>Learn something new. Work on an idea. Meet the people behind the projects. Choose the kind of gathering that interests you.</p>
          </div>
          <ContentSwitcher label="Event types" items={[
            { label: "Hackathons", title: "Build something together.", icon: "code", description: "Spend focused time working on an idea, trying a tool, or solving a problem with others.", note: "Read the organizer’s brief to understand the theme, team requirements, schedule, and what you should bring." },
            { label: "Conferences", title: "Make room for a new perspective.", icon: "mic", description: "Hear different perspectives, explore a subject in depth, and take a few ideas back to your own work.", note: "Look through the event description and schedule to decide whether the topics match what you want to learn." },
            { label: "Meetups", title: "Start with a shared interest.", icon: "people", description: "Make time for a shared interest and conversations with people working on similar things.", note: "A local discussion or a small online gathering can be a useful starting point when you are exploring a new community." },
          ]} />
          <div className="section-actions centered-actions"><a className="button button-outline" href={`${site.app}/discover`}>Browse events</a></div>
        </div>
      </section>
      </div>

      <LandingNavigation />

      <EventStories>
        <section className="landing-section event-story event-story--hackathon" aria-labelledby="hackathons-heading" data-event-story>
          <div className="event-story-inner">
            <div className="event-story-copy">
              <EventStorySymbol kind="hackathon" />
              <h2 id="hackathons-heading" tabIndex={-1}>Hackathons.<br />Make room for an idea.</h2>
              <p>Spend focused time working on an idea, trying a tool, or solving a problem with others. Start with a question and see where a few hours of building can take you.</p>
              <details className="context-note"><summary>Before you join<span aria-hidden="true" /></summary><p>Read the organizer’s brief for the theme, team requirements, schedule, and what to bring. You do not need to arrive with everything figured out; check the event’s experience requirements before joining.</p></details>
              <a className="button button-primary" href={`${site.app}/discover`}>Explore events</a>
            </div>
          </div>
        </section>

        <section className="landing-section event-story event-story--conference" aria-labelledby="conferences-heading" data-event-story>
          <div className="event-story-inner">
            <div className="event-story-copy">
              <EventStorySymbol kind="conference" />
              <h2 id="conferences-heading">Conferences.<br />Leave with a new perspective.</h2>
              <p>Hear how other people approach the work you care about. Explore a subject in depth and take a few ideas back to your own projects.</p>
              <details className="context-note"><summary>Before you join<span aria-hidden="true" /></summary><p>Look through the topics and schedule to find the sessions that interest you. Make room for conversations between them, and check the format and participation details before making plans.</p></details>
              <a className="button button-primary" href={`${site.app}/discover`}>Explore events</a>
            </div>
          </div>
        </section>

        <section className="landing-section event-story event-story--meetup" aria-labelledby="meetups-heading" data-event-story>
          <div className="event-story-inner">
            <div className="event-story-copy">
              <EventStorySymbol kind="meetup" />
              <h2 id="meetups-heading">Meetups.<br />Start with a conversation.</h2>
              <p>Make time for a shared interest and people working on similar things. A small gathering can be a good place to ask a question, share something you learned, or simply listen.</p>
              <details className="context-note"><summary>Before you join<span aria-hidden="true" /></summary><p>Find a subject that makes you curious. Read the description for the format, location, and expected experience, whether the gathering is around the corner or online.</p></details>
              <a className="button button-primary" href={`${site.app}/discover`}>Explore events</a>
            </div>
          </div>
        </section>
      </EventStories>

      <InterestTiles />

      <section className="landing-section formats-section" aria-labelledby="formats-heading">
        <div className="site-container content-section">
          <div className="section-intro">
            <h2 id="formats-heading">Choose how you take part.</h2>
            <p>Your location and your schedule matter. Use the format filter in the attendee app to narrow the published events to the way you want to join.</p>
          </div>
          <ContentSwitcher label="Ways to attend" items={[
            { label: "In person", title: "Be in the same room.", icon: "people", description: "Meet at a physical venue and spend time in the same room. Review the location, start time, travel arrangements, and any entry requirements before you RSVP.", note: "For accessibility or venue questions, use the contact information shared by the organizer." },
            { label: "Online", title: "Join from where you are.", icon: "laptop", description: "Join without travelling to a venue. Check the event’s time zone and the online participation details, especially if you are joining from a different region.", note: "Make sure the schedule and any software requirements work for you." },
            { label: "Hybrid", title: "Choose your way to join.", icon: "globe", description: "Look for an event that combines a physical location with online participation. Read the organizer’s instructions to understand which activities are available in each format.", note: "Choose your way of attending based on the event’s own guidance." },
          ]} />
          <a className="button button-outline" href={`${site.app}/discover`}>Find a format that fits</a>
        </div>
      </section>

      <section className="landing-section rsvp-section" aria-labelledby="rsvp-heading">
        <div className="site-container content-section">
          <div className="section-intro">
            <h2 id="rsvp-heading" tabIndex={-1}>From finding an event to reserving a place.</h2>
            <p>Start with the event information, then submit your registration. You can make a guest RSVP with your name and email address.</p>
          </div>
          <DisclosureList numbered items={[
            { title: "Find something relevant", body: "Browse the published events, search by title or description, and use the format filter to focus your search." },
            { title: "Read the details", body: "Check the description, dates, venue or online format, and registration availability. Ask the organizer about anything that affects your participation." },
            { title: "Submit your RSVP", body: "Enter your name and email, add any optional notes, and submit the form. A place is reserved only after the service confirms your request." },
          ]} />
          <p className="section-note">Availability is checked when your request is processed. Opening the RSVP form does not hold a place, and registration may close before the event starts.</p>
          <div className="section-actions"><a className="button button-primary" href={`${site.app}/discover`}>Find an event</a><Link className="text-link" href="/help">Registration FAQs</Link></div>
        </div>
      </section>

      <section className="landing-section preparation-section" aria-labelledby="preparation-heading">
        <div className="site-container split-section">
          <div className="section-intro">
            <h2 id="preparation-heading">Arrive knowing what to expect.</h2>
            <p>A little preparation leaves more room for the event itself. Check the practical details before you make plans, especially when joining a new community or trying a different format.</p>
            <a className="button button-outline" href={`${site.app}/discover`}>Find your next event</a>
          </div>
          <DisclosureList items={[
            { title: "Make time for the whole experience", body: "Check the start and end times, time zone, and any scheduled breaks. Allow time for travel or setting up an online connection." },
            { title: "Bring what you need", body: "Look for equipment, software, or preparation requirements. A workshop may need a laptop; a discussion may simply need your attention." },
            { title: "Ask about access", body: "Confirm venue accessibility, entry instructions, or online participation arrangements directly with the organizer. The event description is your starting point; ask about anything it does not cover before you register." },
          ]} />
        </div>
      </section>

      <section className="landing-section organizer-surface hosting-section" aria-labelledby="hosting-heading">
        <div className="site-container organizer-feature">
          <div><h2 id="hosting-heading" tabIndex={-1}>Bring your<br />community together.</h2><p>A workshop, a local meetup, or a weekend of building. Give your next event a place to start.</p><details className="context-note"><summary>How hosting works<span aria-hidden="true" /></summary><p>Create an account and organizer profile, then add the information attendees need to decide whether to join. Your published events appear in your own organizer workspace.</p></details><Link className="text-link" href="/organizers">Organizer guide</Link></div>
          <div className="organizer-copy">
            <p>Start with the essentials and make the practical details clear.</p>
            <dl className="feature-notes">
              <div><dt>The purpose</dt><dd>Choose a title and describe what people will do, learn, or work on.</dd></div>
              <div><dt>The schedule</dt><dd>Set the start and end dates with the event’s time zone.</dd></div>
              <div><dt>The place</dt><dd>Add a venue, an online link, or both for a hybrid event.</dd></div>
              <div><dt>The capacity</dt><dd>Set an attendee limit if your event needs one.</dd></div>
            </dl>
            <div className="section-actions"><a className="button button-primary" href="/create">Host an event</a><a className="button button-outline" href={`${site.organizer}/login`}>Organizer sign in</a></div>
          </div>
        </div>
      </section>

      <section className="landing-section brief-section" aria-labelledby="brief-heading">
        <div className="site-container content-section">
          <div className="section-intro">
            <h2 id="brief-heading">A clear invitation starts a better gathering.</h2>
            <p>Help people decide whether your event is right for them. Use the event description to answer the questions you would ask before joining something new.</p>
          </div>
          <DisclosureList items={[
            { title: "Who is it for?", body: "Describe the interests and experience your gathering suits. Say whether beginners are welcome and whether any prior knowledge is expected." },
            { title: "What will people do?", body: "Explain the activities and the purpose behind them. Give attendees a sense of how the time will be spent, from listening and discussion to hands-on work." },
            { title: "What should they prepare?", body: "Mention anything attendees need to bring, install, or read. Clear preparation instructions help people arrive ready to take part." },
            { title: "How will they join?", body: "Make the format, schedule, location, and participation instructions easy to find. Include a way to ask about details that need a personal answer." },
          ]} />
          <div className="section-actions"><a className="button button-primary" href="/create">Create your event</a><Link className="text-link" href="/organizers">Hosting guide</Link></div>
        </div>
      </section>

      <section className="landing-section participation-section" aria-labelledby="participation-heading">
        <div className="site-container split-section">
          <div className="section-intro">
            <h2 id="participation-heading">There is more than one way to belong.</h2>
            <p>You might be learning your first framework, working through a difficult project, or planning a gathering for people you already know.</p>
            <Link className="text-link" href="/about">About Hackways</Link>
          </div>
          <ContentSwitcher label="Find your starting point" items={[
            { label: "Learning", title: "Follow a new interest.", icon: "book", description: "Look for a subject you want to understand better. Check whether the event expects prior experience or offers an introduction to the topic." },
            { label: "Building", title: "Work through an idea.", icon: "code", description: "Find time to explore an approach, talk through a problem, or work alongside people with a shared technical interest." },
            { label: "Organizing", title: "Bring people together.", icon: "people", description: "Give your gathering a clear purpose and a practical schedule. A focused description helps the right people decide to join." },
          ]} />
        </div>
      </section>

      <section className="landing-section answers-section" aria-labelledby="answers-heading">
        <div className="site-container home-faq">
          <div><h2 id="answers-heading" tabIndex={-1}>A few useful answers.</h2><p>The essentials for finding events, submitting a registration, and publishing a gathering of your own.</p><details className="context-note"><summary>Questions about an event?<span aria-hidden="true" /></summary><p>For a question about a particular event, use the contact details supplied by its organizer. Our help page explains the registration and hosting features available on Hackways.</p></details><div className="section-actions"><Link className="text-link" href="/help">Help & FAQs</Link><Link className="text-link" href="/contact">Contact</Link></div></div>
          <FAQ full />
        </div>
      </section>

      <section className="landing-section start-section" aria-labelledby="get-started-heading">
        <div className="site-container content-section">
          <div className="section-intro">
            <h2 id="get-started-heading">Where would you like to start?</h2>
            <p>Take a look at what is happening, or create a gathering around something that matters to your community.</p>
          </div>
          <div className="start-options">
            <div><h3>Find your next event.</h3><p>Browse hackathons, conferences, and meetups, then read the details and RSVP.</p><a className="button button-primary" href={`${site.app}/discover`}>Explore events</a></div>
            <div><h3>Host your own event.</h3><p>Create an organizer profile, add the details, and publish your next gathering.</p><a className="button button-primary" href="/create">Create an event</a></div>
          </div>
        </div>
      </section>
    </div>
  );
}
