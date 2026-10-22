import 'server-only';
import { cache } from 'react';
import { notFound } from 'next/navigation';
import { cookies } from 'next/headers';
import { ApiClientError, fetchApi } from './api';
import { privateApi,currentUser } from './account';
import type { EventDetails, RegistrationFormConfig } from '@/types/api';

export type EventManagementAccess = 'allowed' | 'denied' | 'unavailable';

// Personalized, request-scoped only. Go remains the authority for access.
export const readEventManagementAccess = cache(async (slug: string): Promise<EventManagementAccess> => {
  if (!cookies().get('__Host-hackways_session')?.value || !slug || slug.length > 512) return 'denied';
  try {
    const response = await privateApi<{ data: Pick<EventDetails, 'id' | 'slug'> }>(`/events/manage-by-slug/${encodeURIComponent(slug)}`);
    if (!response.data?.id || response.data.slug !== slug) throw new ApiClientError('Management access could not be verified.', 'INVALID_RESPONSE');
    return 'allowed';
  } catch (error) {
    if (error instanceof ApiClientError && ['UNAUTHORIZED', 'FORBIDDEN', 'NOT_FOUND', 'HTTP_401', 'HTTP_403', 'HTTP_404'].includes(error.code)) return 'denied';
    console.error('Event management access check failed', { slug, code: error instanceof ApiClientError ? error.code : 'UNEXPECTED_ERROR' });
    return 'unavailable';
  }
});

// Request-scoped memoization shares the read across layout, page, and metadata.
export const readEvent = cache(async (slug: string): Promise<EventDetails> => {
  if (!slug || slug.length > 512) notFound();
  try {
    const response = await fetchApi<{ data: EventDetails }>(`/events/by-slug/${encodeURIComponent(slug)}`, { cache: 'no-store' });
    if (!response.data?.id || response.data.slug !== slug) throw new ApiClientError('The event response could not be read.', 'INVALID_RESPONSE');
    return response.data;
  } catch (error) {
    if (error instanceof ApiClientError && ['NOT_FOUND', 'HTTP_404', 'INVALID_SLUG'].includes(error.code)) notFound();
    throw error;
  }
});

export const readRegistrationForm = cache(async (slug: string): Promise<RegistrationFormConfig> => {
  if (!slug || slug.length > 512) notFound();
  try {
    return (await fetchApi<{ data: RegistrationFormConfig }>(`/events/by-slug/${encodeURIComponent(slug)}/registration-form`, { cache: 'no-store' })).data;
  } catch (error) {
    if (error instanceof ApiClientError && ['NOT_FOUND', 'HTTP_404', 'INVALID_SLUG'].includes(error.code)) notFound();
    throw error;
  }
});

export const readInvitationState = cache(async (slug:string):Promise<{invited:boolean;registration_code:string|null}> => {
 const user=await currentUser();if(!user)return {invited:false,registration_code:null};
 return privateApi<{invited:boolean;registration_code:string|null}>(`/events/by-slug/${encodeURIComponent(slug)}/invitation`);
});
