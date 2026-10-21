export * from "@hackways/api-client";
import type { Event as SharedEvent } from "@hackways/api-client";
export type Event = SharedEvent & {
  registration_opens_at?: string | null;
  registration_closes_at?: string | null;
};
export type EventDetails = Event & {
  space_name?: string | null;
  space_slug?: string | null;
  organizer_name: string;
  organizer_bio?: string | null;
  organizer_website?: string | null;
  organizer_verified: boolean;
};
