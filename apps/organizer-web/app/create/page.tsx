import { redirect } from 'next/navigation';
import { site } from '@/lib/site';
export default function Create() { redirect(`${site.app}/create`); }
