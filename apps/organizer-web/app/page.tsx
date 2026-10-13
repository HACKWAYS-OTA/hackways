"use client";

import React, { useState, useEffect, useCallback } from "react";
import { Event } from "@/types/api";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Modal } from "@/components/ui/Modal";
import { Skeleton } from "@/components/ui/Skeleton";
import { fetchApi } from "@/lib/api";

export default function OrganizerDashboardPage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState<boolean>(false);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [createError, setCreateError] = useState<string | null>(null);

  // Form state
  const [title, setTitle] = useState<string>("");
  const [description, setDescription] = useState<string>("");
  const [format, setFormat] = useState<"in_person" | "online" | "hybrid">("in_person");
  const [capacity, setCapacity] = useState<number>(100);
  const [startsAt, setStartsAt] = useState<string>("");
  const [endsAt, setEndsAt] = useState<string>("");
  const [locationName, setLocationName] = useState<string>("");

  const loadOrganizerEvents = useCallback(async () => {
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
    loadOrganizerEvents();
  }, [loadOrganizerEvents]);

  const handleCreateEvent = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setCreateError(null);

    try {
      const payload = {
        title,
        description,
        format,
        capacity: Number(capacity),
        starts_at: startsAt || new Date(Date.now() + 86400000).toISOString(),
        ends_at: endsAt || new Date(Date.now() + 172800000).toISOString(),
        location_name: locationName || "Hackways HQ",
        timezone: "UTC",
      };

      await fetchApi<Event>("/events", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      setIsCreateModalOpen(false);
      // Reset form
      setTitle("");
      setDescription("");
      setCapacity(100);
      setLocationName("");
      // Reload events
      await loadOrganizerEvents();
    } catch (err) {
      setCreateError(
        err instanceof Error ? err.message : "Failed to create event. Please try again."
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  // Metric derivations
  const totalEvents = events.length;
  const totalRsvps = events.reduce((acc, evt) => acc + (evt.registration_count || 0), 0);
  const totalCapacity = events.reduce((acc, evt) => acc + (evt.capacity || 0), 0);
  const utilization = totalCapacity > 0 ? Math.round((totalRsvps / totalCapacity) * 100) : 0;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      {/* Studio Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-8 border-b border-slate-800">
        <div>
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full bg-indigo-950 border border-indigo-800 text-xs font-medium text-indigo-400 mb-2">
            Organizer Studio Engine
          </div>
          <h1 className="text-3xl font-extrabold text-white">Event Studio & Roster Management</h1>
          <p className="text-sm text-slate-400 mt-1">
            Author events, set strict concurrency-locked capacities, and inspect attendee registrations.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="secondary" size="md" onClick={() => loadOrganizerEvents()}>
            Refresh Data
          </Button>
          <Button variant="primary" size="md" onClick={() => setIsCreateModalOpen(true)}>
            + Create New Event
          </Button>
        </div>
      </div>

      {/* Analytics KPI Row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 py-8">
        <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
          <p className="text-xs font-medium text-slate-400 uppercase tracking-wider">Total Events</p>
          <p className="text-3xl font-extrabold text-white mt-2">{isLoading ? "-" : totalEvents}</p>
          <p className="text-xs text-indigo-400 mt-2">Active on platform</p>
        </div>

        <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
          <p className="text-xs font-medium text-slate-400 uppercase tracking-wider">Confirmed RSVPs</p>
          <p className="text-3xl font-extrabold text-white mt-2">{isLoading ? "-" : totalRsvps}</p>
          <p className="text-xs text-emerald-400 mt-2">Pessimistic locked</p>
        </div>

        <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
          <p className="text-xs font-medium text-slate-400 uppercase tracking-wider">Total Capacity</p>
          <p className="text-3xl font-extrabold text-white mt-2">{isLoading ? "-" : totalCapacity}</p>
          <p className="text-xs text-slate-400 mt-2">Across all tiers</p>
        </div>

        <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6">
          <p className="text-xs font-medium text-slate-400 uppercase tracking-wider">Capacity Fill Rate</p>
          <p className="text-3xl font-extrabold text-white mt-2">{isLoading ? "-" : `${utilization}%`}</p>
          <p className="text-xs text-indigo-400 mt-2">Zero overbooking</p>
        </div>
      </div>

      {/* Error state */}
      {!isLoading && error && (
        <div className="text-center py-12 bg-slate-900/40 border border-rose-900/50 rounded-2xl max-w-xl mx-auto p-8 mb-8">
          <h3 className="text-lg font-semibold text-white mb-2">Failed to Load Organizer Data</h3>
          <p className="text-slate-400 text-sm mb-6">{error}</p>
          <Button variant="primary" size="md" onClick={loadOrganizerEvents}>
            Retry Connection
          </Button>
        </div>
      )}

      {/* Events Table */}
      <div className="bg-slate-900/50 border border-slate-800 rounded-2xl overflow-hidden shadow-xl">
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-bold text-white">Managed Events</h2>
            <p className="text-xs text-slate-400">Real-time inventory and registration status</p>
          </div>
        </div>

        {isLoading ? (
          <div className="p-6 space-y-4">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        ) : events.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-slate-400 text-base mb-4">No events found in your studio.</p>
            <Button variant="primary" size="sm" onClick={() => setIsCreateModalOpen(true)}>
              Create First Event
            </Button>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300">
              <thead className="bg-slate-950/60 text-xs font-semibold uppercase text-slate-400 border-b border-slate-800">
                <tr>
                  <th className="px-6 py-3.5">Title</th>
                  <th className="px-6 py-3.5">Format</th>
                  <th className="px-6 py-3.5">Capacity</th>
                  <th className="px-6 py-3.5">Confirmed</th>
                  <th className="px-6 py-3.5">Status</th>
                  <th className="px-6 py-3.5 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {events.map((evt) => {
                  const isFull = evt.capacity ? evt.registration_count >= evt.capacity : false;
                  return (
                    <tr key={evt.id} className="hover:bg-slate-800/30 transition">
                      <td className="px-6 py-4 font-medium text-white">{evt.title}</td>
                      <td className="px-6 py-4 capitalize text-slate-400">{evt.format.replace("_", " ")}</td>
                      <td className="px-6 py-4 font-mono">{evt.capacity ?? "Unlimited"}</td>
                      <td className="px-6 py-4 font-mono">
                        <span className={isFull ? "text-rose-400 font-bold" : "text-emerald-400"}>
                          {evt.registration_count}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="px-2 py-0.5 rounded text-xs font-medium uppercase bg-indigo-950 border border-indigo-800 text-indigo-400">
                          {evt.status}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <span className="text-xs text-slate-500 font-mono">{evt.id.substring(0, 8)}...</span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create Event Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Create New Event"
      >
        <form onSubmit={handleCreateEvent} className="space-y-4">
          {createError && (
            <div className="p-3 rounded-lg bg-rose-950/80 border border-rose-800 text-rose-300 text-sm">
              {createError}
            </div>
          )}

          <Input
            label="Event Title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Hackways Global Hackathon 2026"
            required
          />

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1.5">Description</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Detailed schedule and details..."
              required
              rows={3}
              className="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-2.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-1.5">Format</label>
              <select
                value={format}
                onChange={(e) => setFormat(e.target.value as "in_person" | "online" | "hybrid")}
                className="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-indigo-500 transition"
              >
                <option value="in_person">In Person</option>
                <option value="online">Online</option>
                <option value="hybrid">Hybrid</option>
              </select>
            </div>

            <Input
              label="Maximum Capacity"
              type="number"
              value={capacity}
              onChange={(e) => setCapacity(Number(e.target.value))}
              required
            />
          </div>

          <Input
            label="Location / Venue Name"
            value={locationName}
            onChange={(e) => setLocationName(e.target.value)}
            placeholder="e.g. Tech Hub Auditorium"
          />

          <div className="pt-4 flex items-center justify-end gap-3 border-t border-slate-800">
            <Button
              type="button"
              variant="secondary"
              size="md"
              onClick={() => setIsCreateModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="md"
              isLoading={isSubmitting}
            >
              Publish Event
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
