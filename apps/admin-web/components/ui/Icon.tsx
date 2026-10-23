import { HugeiconsIcon } from '@hugeicons/react';
import Analytics01Icon from '@hugeicons/core-free-icons/Analytics01Icon';
import Mail01Icon from '@hugeicons/core-free-icons/Mail01Icon';
import Ticket01Icon from '@hugeicons/core-free-icons/Ticket01Icon';
import File01Icon from '@hugeicons/core-free-icons/File01Icon';
import Task01Icon from '@hugeicons/core-free-icons/Task01Icon';
import Audit01Icon from '@hugeicons/core-free-icons/Audit01Icon';
import Shield01Icon from '@hugeicons/core-free-icons/Shield01Icon';
import Home01Icon from '@hugeicons/core-free-icons/Home01Icon';
import Calendar03Icon from '@hugeicons/core-free-icons/Calendar03Icon';
import UserGroupIcon from '@hugeicons/core-free-icons/UserGroupIcon';
import Search01Icon from '@hugeicons/core-free-icons/Search01Icon';
import Add01Icon from '@hugeicons/core-free-icons/Add01Icon';
import ArrowRight01Icon from '@hugeicons/core-free-icons/ArrowRight01Icon';
import ArrowLeft01Icon from '@hugeicons/core-free-icons/ArrowLeft01Icon';
import ArrowUpRight01Icon from '@hugeicons/core-free-icons/ArrowUpRight01Icon';
import ArrowUp01Icon from '@hugeicons/core-free-icons/ArrowUp01Icon';
import ArrowDown01Icon from '@hugeicons/core-free-icons/ArrowDown01Icon';
import Cancel01Icon from '@hugeicons/core-free-icons/Cancel01Icon';
import ReloadIcon from '@hugeicons/core-free-icons/ReloadIcon';

const icons = { analytics: Analytics01Icon, up: ArrowUp01Icon, down: ArrowDown01Icon, close: Cancel01Icon, people: UserGroupIcon, messaging: Mail01Icon, tickets: Ticket01Icon, storage: File01Icon, registrations: Task01Icon, audit: Audit01Icon, moderation: Shield01Icon, home: Home01Icon, events: Calendar03Icon, spaces: UserGroupIcon, search: Search01Icon, add: Add01Icon, arrow: ArrowRight01Icon, back: ArrowLeft01Icon, external: ArrowUpRight01Icon, reload: ReloadIcon };
export function Icon({ name, size = 20, className = '' }: { name: keyof typeof icons; size?: number; className?: string }) {
  return <HugeiconsIcon icon={icons[name]} size={size} strokeWidth={1.8} className={className} aria-hidden="true" focusable="false" />;
}
