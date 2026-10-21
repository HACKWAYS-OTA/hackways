import type { RegistrationStatus } from '@/types/api';
export interface Guest { id: string; public_code: string; user_id: string | null; guest_name: string | null; guest_email: string | null; status: RegistrationStatus; notes: string | null; registered_at: string; offer_expires_at: string | null; ticket_id: string | null; ticket_status: string | null; checked_in_at: string | null }
export interface GuestStatusOption { value: RegistrationStatus; label: string }
export interface EditableGuest extends Guest { updated_at: string }
