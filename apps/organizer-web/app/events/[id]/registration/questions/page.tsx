import { requireEventPermission } from '@/lib/management';
import { privateApi } from '@/lib/account';
import type { RegistrationFormConfig } from '@/types/api';
import { RegistrationTabs } from '@/components/management/RegistrationTabs';
import { QuestionBuilder } from '@/components/management/QuestionBuilder';
export const metadata = { title: 'Registration questions' };
export default async function Questions({ params }: { params: { id: string } }) {
  const event = (await requireEventPermission(params.id, 'content')).data;
  const form = (await privateApi<{ data: RegistrationFormConfig }>(`/events/${event.id}/registration-form`)).data;
  return <><RegistrationTabs slug={event.slug} questions /><QuestionBuilder key={form.revision} eventId={event.id} initial={form} editable={!event.management_closed} /></>;
}
