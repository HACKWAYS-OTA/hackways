import type { Event } from '@/types/api';
import { fetchApi } from './api';
interface UploadIntent { upload_url: string; public_url: string; blob_key: string; expires_at: string }
export interface CoverAttempt { file: File; eventID: string; intent: UploadIntent; uploaded: boolean }
export async function uploadEventCover(file: File, event: Event, previous: CoverAttempt | null, remember: (attempt: CoverAttempt) => void): Promise<Event> {
  let attempt = previous;
  if (!attempt || attempt.file !== file || attempt.eventID !== event.id || (!attempt.uploaded && Date.parse(attempt.intent.expires_at) <= Date.now())) {
    const intent = await fetchApi<UploadIntent>('/storage/upload-intent', { method: 'POST', body: JSON.stringify({ filename: file.name, mime_type: file.type, size: file.size, category: 'event-banners', resource_id: event.id }) });
    attempt = { file, eventID: event.id, intent, uploaded: false }; remember(attempt);
  }
  if (!attempt.uploaded) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 60000);
    try {
      const result = await fetch(attempt.intent.upload_url, { method: 'PUT', body: file, credentials: 'omit', signal: controller.signal, headers: { 'Content-Type': file.type, 'x-ms-blob-type': 'BlockBlob' } });
      // A previous upload may have completed before its response was lost. The Go
      // finalizer still verifies the immutable object before any event attachment.
      if (!result.ok && result.status !== 409 && result.status !== 412) throw new Error('The image upload failed. Your event record is saved; please retry.');
      attempt.uploaded = true;
    } catch (error) {
      if (controller.signal.aborted) throw new Error('The image upload timed out. Your event record is saved; please retry.');
      if (error instanceof TypeError) throw new Error('Couldn’t reach image storage. Check your connection and retry.');
      throw error;
    } finally { clearTimeout(timeout); }
  }
  const response = await fetchApi<{ data: Event }>('/storage/event-banner/complete', { method: 'POST', body: JSON.stringify({ resource_id: event.id, blob_key: attempt.intent.blob_key, revision: event.revision }) });
  return response.data;
}
