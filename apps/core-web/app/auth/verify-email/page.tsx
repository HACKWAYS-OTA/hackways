import { redirect } from 'next/navigation';
import { EmailVerification } from '@/components/account/AuthRecovery';
import { currentUser } from '@/lib/account';
import { safeNext } from '@/lib/journeys';
export const metadata={title:'Verify email',robots:{index:false,follow:false}};
export default async function VerifyEmail({searchParams}:{searchParams:{next?:string;queued?:string;delivery?:string}}) {
 const next=safeNext(searchParams.next),user=await currentUser();
 if(!user) redirect(`/login?next=${encodeURIComponent(`/auth/verify-email?next=${encodeURIComponent(next)}`)}`);
 if(user.email_verified) redirect(next);
 return <EmailVerification email={user.email} next={next} queued={searchParams.queued==='1'} unavailable={searchParams.delivery==='unavailable'} />;
}
