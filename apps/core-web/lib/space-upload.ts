import type { Space } from '@/types/api';
import { fetchApi } from './api';
type Intent = { upload_url: string; blob_key: string; expires_at: string };
export type SpaceImageAttempt = { file: File; spaceID: string; kind: 'logo' | 'banner'; intent: Intent; uploaded: boolean };
export async function saveSpaceImage(space: Space, kind: 'logo' | 'banner', file: File | null, previous: SpaceImageAttempt | undefined, remember: (value: SpaceImageAttempt) => void): Promise<Space> {
  let attempt = previous;
  if (file) {
    if (!attempt || attempt.file !== file || attempt.spaceID !== space.id || !attempt.uploaded && Date.parse(attempt.intent.expires_at) <= Date.now()) {
      const intent = await fetchApi<Intent>('/storage/upload-intent', { method: 'POST', body: JSON.stringify({ resource_id: space.id, category: `space-${kind}s`, filename: file.name, mime_type: file.type, size: file.size }) });
      attempt = { file, spaceID: space.id, kind, intent, uploaded: false }; remember(attempt);
    }
    if (!attempt.uploaded) {
      const controller = new AbortController(), timer = setTimeout(() => controller.abort(), 60000);
      try { const result = await fetch(attempt.intent.upload_url, { method: 'PUT', credentials: 'omit', body: file, signal: controller.signal, headers: { 'Content-Type': file.type, 'x-ms-blob-type': 'BlockBlob' } }); if (!result.ok && result.status !== 409 && result.status !== 412) throw new Error('Image upload failed. Your Space is saved; retry to finish.'); attempt.uploaded = true; }
      finally { clearTimeout(timer); }
    }
  }
  const result = await fetchApi<{ data: Pick<Space, 'id' | 'slug' | 'revision' | 'logo_url' | 'banner_url'> }>('/storage/space-image/complete', { method: 'POST', body: JSON.stringify({ resource_id: space.id, kind, revision: space.revision, remove: !file, blob_key: attempt?.intent.blob_key || '' }) });
  return { ...space, ...result.data };
}
