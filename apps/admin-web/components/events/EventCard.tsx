import React from "react";
import { Event } from "@/types/api";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

interface EventCardProps {
  event: Event;
  onRegister?: (eventId: string) => void;
}

export const EventCard: React.FC<EventCardProps> = ({ event, onRegister }) => {
  const isAtCapacity =
    event.capacity !== null &&
    event.capacity !== undefined &&
    event.current_rsvp_count >= event.capacity;

  const percentageFull =
    event.capacity && event.capacity > 0
      ? Math.min(100, Math.round((event.current_rsvp_count / event.capacity) * 100))
      : null;

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-sm hover:border-slate-700 transition flex flex-col justify-between">
      <div className="p-5">
        <div className="flex items-center justify-between gap-2 mb-3">
          <Badge variant={event.format === "online" ? "brand" : "default"}>
            {event.format === "in_person"
              ? "In Person"
              : event.format === "online"
              ? "Online"
              : "Hybrid"}
          </Badge>
          <span className="text-xs text-slate-400">
            {new Date(event.starts_at).toLocaleDateString(undefined, {
              month: "short",
              day: "numeric",
              year: "numeric",
            })}
          </span>
        </div>

        <h3 className="text-lg font-semibold text-white mb-2 line-clamp-1">{event.title}</h3>
        <p className="text-sm text-slate-400 line-clamp-2 mb-4">
          {event.short_description || event.description}
        </p>

        {percentageFull !== null && (
          <div className="mb-4">
            <div className="flex justify-between text-xs text-slate-400 mb-1">
              <span>Capacity</span>
              <span>
                {event.current_rsvp_count} / {event.capacity} ({percentageFull}%)
              </span>
            </div>
            <div className="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all duration-300 ${
                  percentageFull >= 100
                    ? "bg-rose-500"
                    : percentageFull >= 80
                    ? "bg-amber-500"
                    : "bg-indigo-500"
                }`}
                style={{ width: `${percentageFull}%` }}
              />
            </div>
          </div>
        )}
      </div>

      <div className="px-5 py-4 bg-slate-950/50 border-t border-slate-800/80 flex items-center justify-between">
        <span className="text-xs text-slate-500 truncate max-w-[180px]">
          {event.location_name || event.online_url || "Details on registration"}
        </span>
        <Button
          size="sm"
          variant={isAtCapacity ? "secondary" : "primary"}
          disabled={isAtCapacity}
          onClick={() => onRegister?.(event.id)}
        >
          {isAtCapacity ? "Full" : "RSVP"}
        </Button>
      </div>
    </div>
  );
};
