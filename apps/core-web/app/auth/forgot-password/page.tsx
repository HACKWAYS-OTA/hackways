import { PasswordRecovery } from '@/components/account/AuthRecovery';
import { safeNext } from '@/lib/journeys';
export const metadata={title:'Forgot password',robots:{index:false,follow:false}};
export default function ForgotPassword({searchParams}:{searchParams:{next?:string}}) { return <PasswordRecovery next={safeNext(searchParams.next)} />; }
