"use client";

import React, { useState } from "react";
import { Event } from "@/types/api";
import { EventCard } from "@/components/events/EventCard";

// Initial sample events demonstrating rendering and layout
const SAMPLE_EVENTS: Event[] = [
  {
    id: "evt-001",
    organizer_id: "org-001",
    status: "published",
    format: "in_person",
    title: "Global Hackathon 2026: AI & Systems",
    slug: "global-hackathon-2026",
    description: "Build cutting-edge distributed systems and high-throughput applications with peer developers.",
    short_description: "Build cutting-edge distributed systems with peer developers.",
    location_name: "Metropolitan Tech Hub, Innovation Hall",
    starts_at: "2026-11-20T09:00:00Z",
    ends_at: "2026-11-22T18:00:00Z",
    timezone: "UTC",
    capacity: 250,
    registration_count: 184,
    tags: ["hackathon", "golang", "ai"],
  },
  {
    id: "evt-002",
    organizer_id: "org-002",
    status: "published",
    format: "online",
    title: "Building High-Throughput APIs with Go & Chi",
    slug: "high-throughput-apis-go",
    description: "Deep dive into goroutine concurrency, connection pooling, and zero-allocation routing.",
    short_description: "Deep dive into goroutine concurrency and connection pooling.",
    online_url: "https://stream.hackways.local/go-api-summit",
    starts_at: "2026-12-05T16:00:00Z",
    ends_at: "2026-12-05T19:00:00Z",
    timezone: "UTC",
    capacity: 500,
    registration_count: 495,
    tags: ["go", "webinar", "architecture"],
  },
  {
    id: "evt-003",
    organizer_id: "org-003",
    status: "published",
    format: "hybrid",
    title: "PostgreSQL at Scale: Transactions & Locking",
    slug: "postgres-at-scale",
    description: "Master row-level locking, GIN indexing, and MVCC internals under extreme concurrency.",
    short_description: "Master row-level locking, GIN indexing, and MVCC internals.",
    location_name: "Downtown Dev Lounge & Online Stream",
    starts_at: "2026-12-15T14:00:00Z",
    ends_at: "2026-12-15T17:00:00Z",
    timezone: "UTC",
    capacity: 100,
    registration_count: 100,
    tags: ["database", "postgresql", "performance"],
  },
];

export default function HomePage() {
  const [filterFormat, setFilterFormat] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState<string>("");

  const filteredEvents = SAMPLE_EVENTS.filter((evt) => {
    const matchesFormat = filterFormat === "all" || evt.format === filterFormat;
    const matchesQuery =
      searchQuery.trim() === "" ||
      evt.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      evt.description.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesFormat && matchesQuery;
  });

  const handleRegister = (eventId: string) => {
    alert(`RSVP action triggered for event ID: ${eventId}`);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      {/* Hero Section */}
      <section className="text-center py-12 border-b border-slate-800">
        <h1 className="text-4xl sm:text-5xl font-extrabold tracking-tight text-white mb-4">
          Discover & RSVP to World-Class <span className="text-indigo-400">Tech Events</span>
        </h1>
        <p className="max-w-2xl mx-auto text-lg text-slate-400 mb-8">
          Hackways connects developers, creators, and organizers with concurrency-safe registration and real-time capacity management.
        </p>

        {/* Search input */}
        <div className="max-w-md mx-auto flex items-center bg-slate-900 border border-slate-800 rounded-xl p-1.5 focus-within:border-indigo-500 transition shadow-lg">
          <input
            type="text"
            placeholder="Search events, keywords, or topics..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-transparent px-4 py-2 text-sm text-white placeholder-slate-500 focus:outline-none"
          />
        </div>
      </section>

      {/* Filter and Event Listing */}
      <section id="events" className="py-10">
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

        {filteredEvents.length === 0 ? (
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
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredEvents.map((evt) => (
              <EventCard key={evt.id} event={evt} onRegister={handleRegister} />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
