"use client";

import React, { useState, useEffect, useCallback } from "react";
import { Event } from "@/types/api";
import { EventCard } from "@/components/events/EventCard";
import { EventCardSkeleton } from "@/components/ui/Skeleton";
import { RSVPModal } from "@/components/events/RSVPModal";
import { Button } from "@/components/ui/Button";
import { fetchApi } from "@/lib/api";

export default function HomePage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [filterFormat, setFilterFormat] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);

  const loadEvents = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await fetchApi<{ data: Event[] }>("/events");
      setEvents(response.data || []);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to load events from server. Please check your connection."
      );
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadEvents();
  }, [loadEvents]);

  const filteredEvents = events.filter((evt) => {
    const matchesFormat = filterFormat === "all" || evt.format === filterFormat;
    const matchesQuery =
      searchQuery.trim() === "" ||
      evt.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      evt.description.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesFormat && matchesQuery;
  });

  const handleOpenRSVP = (eventId: string) => {
    const target = events.find((e) => e.id === eventId);
    if (target) {
      setSelectedEvent(target);
      setIsModalOpen(true);
    }
  };

  const handleRSVPSuccess = (eventId: string) => {
    // Atomically increment the event's registration counter in UI
    setEvents((prev) =>
      prev.map((evt) =>
        evt.id === eventId
          ? { ...evt, registration_count: evt.registration_count + 1 }
          : evt
      )
    );
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      {/* Hero Section */}
      <section className="text-center py-16 border-b border-slate-800">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-950/80 border border-indigo-800/60 text-xs font-semibold text-indigo-400 mb-6">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          Production Engineering Baseline • 6-Service Architecture
        </div>
        <h1 className="text-4xl sm:text-6xl font-extrabold tracking-tight text-white mb-6">
          High-Concurrency Tech Events. <br />
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-cyan-400">
            Zero Overbooking. Guaranteed.
          </span>
        </h1>
        <p className="max-w-3xl mx-auto text-lg text-slate-400 mb-10 leading-relaxed">
          Hackways is an enterprise-grade platform uniting attendees and organizers with transactional PostgreSQL row locking, 
          direct-to-Azure SAS uploads, and an asynchronous transactional email outbox.
        </p>

        <div className="flex flex-wrap items-center justify-center gap-4">
          <a
            href={process.env.NEXT_PUBLIC_APP_URL || "https://app.hackways.in"}
            className="px-6 py-3 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white font-medium text-base shadow-lg shadow-indigo-600/25 transition"
          >
            Launch Attendee Portal (app.hackways.in)
          </a>
          <a
            href={process.env.NEXT_PUBLIC_ORGANIZER_URL || "https://organizer.hackways.in"}
            className="px-6 py-3 rounded-xl bg-slate-900 hover:bg-slate-800 text-slate-200 border border-slate-800 font-medium text-base transition"
          >
            Host with Organizer Studio
          </a>
        </div>
      </section>

      {/* Architectural Guarantees & Features */}
      <section id="features" className="py-16 border-b border-slate-800">
        <div className="text-center mb-12">
          <h2 className="text-2xl sm:text-3xl font-bold text-white mb-3">Architectural Guarantees</h2>
          <p className="text-slate-400 text-base max-w-2xl mx-auto">
            Strict engineering standards enforced at every layer — from PostgreSQL transaction isolation to Azure managed services.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
            <div className="w-10 h-10 rounded-xl bg-indigo-950 border border-indigo-800/80 flex items-center justify-center text-indigo-400 mb-4 font-mono font-bold">
              01
            </div>
            <h3 className="text-lg font-semibold text-white mb-2">Zero Overbooking Invariant</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Pessimistic row-level locking via <code className="text-indigo-300 text-xs">SELECT ... FOR UPDATE</code> with a 2000ms statement timeout prevents capacity exhaustion even under concurrent traffic spikes.
            </p>
          </div>

          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
            <div className="w-10 h-10 rounded-xl bg-indigo-950 border border-indigo-800/80 flex items-center justify-center text-indigo-400 mb-4 font-mono font-bold">
              02
            </div>
            <h3 className="text-lg font-semibold text-white mb-2">Direct-to-Cloud Blob SAS</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Zero uploads touch the VM disk. Clients upload event banners and avatars directly to Azure Blob Storage using scoped 15-minute write-only SAS tokens.
            </p>
          </div>

          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
            <div className="w-10 h-10 rounded-xl bg-indigo-950 border border-indigo-800/80 flex items-center justify-center text-indigo-400 mb-4 font-mono font-bold">
              03
            </div>
            <h3 className="text-lg font-semibold text-white mb-2">Transactional Outbox Worker</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Transactional confirmation emails are written atomically to PostgreSQL and reliably dispatched via Azure Communication Services by a bounded 2-goroutine worker.
            </p>
          </div>
        </div>
      </section>

      {/* Filter and Event Listing */}
      <section id="events" className="py-12">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
          <div>
            <h2 className="text-2xl font-bold text-white">Upcoming Events</h2>
            <p className="text-sm text-slate-400">Explore published hackathons, conferences, and meetups</p>
          </div>

          {/* Filter Pills */}
          <div className="flex items-center gap-2 overflow-x-auto pb-2 sm:pb-0">
            {["all", "in_person", "online", "hybrid"].map((format) => (
              <button
                key={format}
                onClick={() => setFilterFormat(format)}
                className={`px-3 py-1.5 rounded-lg text-xs font-medium capitalize transition ${
                  filterFormat === format
                    ? "bg-indigo-600 text-white shadow-sm"
                    : "bg-slate-900 text-slate-400 hover:text-white border border-slate-800"
                }`}
              >
                {format === "all" ? "All Formats" : format.replace("_", " ")}
              </button>
            ))}
          </div>
        </div>

        {/* Loading Skeletons */}
        {isLoading && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <EventCardSkeleton />
            <EventCardSkeleton />
            <EventCardSkeleton />
          </div>
        )}

        {/* Error State with Recovery */}
        {!isLoading && error && (
          <div className="text-center py-16 bg-slate-900/40 border border-rose-900/50 rounded-2xl max-w-xl mx-auto p-8">
            <div className="w-12 h-12 mx-auto mb-4 bg-rose-950/80 border border-rose-800 rounded-full flex items-center justify-center text-rose-400">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-white mb-2">Unable to Load Events</h3>
            <p className="text-slate-400 text-sm mb-6">{error}</p>
            <Button variant="primary" size="md" onClick={loadEvents}>
              Retry Connection
            </Button>
          </div>
        )}

        {/* Empty State */}
        {!isLoading && !error && filteredEvents.length === 0 && (
          <div className="text-center py-16 bg-slate-900/40 border border-slate-800/80 rounded-2xl">
            <p className="text-slate-400 text-base mb-2">No events found matching your criteria.</p>
            <button
              onClick={() => {
                setFilterFormat("all");
                setSearchQuery("");
              }}
              className="text-sm text-indigo-400 hover:underline"
            >
              Reset filters
            </button>
          </div>
        )}

        {/* Events Grid */}
        {!isLoading && !error && filteredEvents.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredEvents.map((evt) => (
              <EventCard key={evt.id} event={evt} onRegister={handleOpenRSVP} />
            ))}
          </div>
        )}
      </section>

      {/* Interactive RSVP Modal */}
      <RSVPModal
        event={selectedEvent}
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={handleRSVPSuccess}
      />
    </div>
  );
}
