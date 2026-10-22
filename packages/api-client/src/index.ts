/**
 * Hackways Shared TypeScript API Contracts & Client
 */

export type EventFormat = "in_person" | "online" | "hybrid";
export type EventStatus = "draft" | "published" | "unpublished" | "cancelled";
export type RegistrationStatus = "confirmed" | "cancelled" | "waitlisted" | "pending_approval" | "offered" | "rejected" | "expired" | "transferred";

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
  current_rsvp_count: number;
  held_rsvp_count?: number;
  revision?: number;
  registration_mode?: "instant" | "approval";
  allow_guest?: boolean;
  waitlist_enabled?: boolean;
  offer_minutes?: number;
  tags: string[];
  published_at?: string | null;
  space_id?: string | null;
  theme?: string;
  theme_color?: string;
  theme_font?: 'futurist' | 'google';
  theme_style?: 'soft' | 'crisp';
  theme_display?: 'light' | 'dark';
  visibility?: 'public' | 'private';
  faq_text?: string;
  contact_url?: string;
  rewards?: string;
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
  public_code: string;
  event_id: string;
  user_id?: string | null;
  guest_name?: string | null;
  guest_email?: string | null;
  status: RegistrationStatus;
  notes?: string | null;
  registered_at: string;
  offer_expires_at?: string | null;
  approval_expires_at?: string | null;
  transfer_count?: number;
  transferred_from?: string | null;
}

export type QuestionKind = 'text' | 'textarea' | 'url' | 'number' | 'checkbox' | 'select' | 'phone' | 'email' | 'date' | 'time' | 'multiselect' | 'section';
export interface RegistrationQuestion {
  public_code: string; label: string; description: string; kind: QuestionKind;
  required: boolean; options: string[]; placeholder?: string; validation?: QuestionValidation;
}
export interface QuestionValidation { min_length?: number; max_length?: number; min?: number; max?: number; integer?: boolean; message?: string }
export interface RegistrationFormSettings { title?: string; description?: string; submit_label?: string; success_message?: string }
export interface RegistrationFormConfig {
  revision: number; name_required: boolean; questions: RegistrationQuestion[]; settings?: RegistrationFormSettings;
}
export interface RegistrationAnswer { code: string; text: string; checked: boolean; selected?: string[] }
export interface RegistrationAnswerSnapshot extends RegistrationAnswer { label: string; kind: QuestionKind; section?: string }

export interface Space {
  id: string; organizer_id: string; name: string; slug: string; description: string;
  guidelines: string; website_url: string; status: 'active' | 'archived'; revision: number;
  created_at: string; updated_at: string; owner_name?: string; member_count?: number; owned?: boolean;
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

export class HackwaysClient {
  private baseURL: string;

  constructor(baseURL: string = "http://localhost:8081/api/v1") {
    this.baseURL = baseURL.replace(/\/$/, "");
  }

  async getHealth(): Promise<{ status: string; version: string }> {
    const res = await fetch(`${this.baseURL.replace(/\/api\/v1$/, "")}/health`);
    if (!res.ok) throw new Error(`Health check failed with HTTP ${res.status}`);
    return res.json() as Promise<{ status: string; version: string }>;
  }

  async listEvents(): Promise<ApiResponse<Event[]>> {
    const res = await fetch(`${this.baseURL}/events`);
    if (!res.ok) throw new Error(`Failed to list events: HTTP ${res.status}`);
    return res.json() as Promise<ApiResponse<Event[]>>;
  }
}
