export type EventFormat = "in_person" | "online" | "hybrid";
export type EventStatus = "draft" | "published" | "unpublished" | "cancelled";

export interface Event {
  id: string;
  organizer_id: string;
  category_id?: number | null;
  status: EventStatus;
  format: EventFormat;
  title: string;
  slug: string;
  description: string;
  short_description?: string | null;
  banner_url?: string | null;
  location_name?: string | null;
  location_address?: string | null;
  online_url?: string | null;
  starts_at: string;
  ends_at: string;
  timezone: string;
  capacity?: number | null;
  registration_count: number;
  tags: string[];
  published_at?: string | null;
}

export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string | null;
  email_verified: boolean;
  created_at: string;
}

export interface Registration {
  id: string;
  event_id: string;
  user_id?: string | null;
  guest_name?: string | null;
  guest_email?: string | null;
  status: "confirmed" | "cancelled" | "waitlisted";
  registered_at: string;
}

export interface ApiResponse<T> {
  data: T;
  pagination?: {
    cursor?: string;
    has_more: boolean;
    total: number;
  };
}

export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}
