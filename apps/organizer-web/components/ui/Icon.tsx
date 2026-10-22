import { HugeiconsIcon } from '@hugeicons/react';
import Search01Icon from '@hugeicons/core-free-icons/Search01Icon';
import Calendar03Icon from '@hugeicons/core-free-icons/Calendar03Icon';
import Location01Icon from '@hugeicons/core-free-icons/Location01Icon';
import ArrowUpRight01Icon from '@hugeicons/core-free-icons/ArrowUpRight01Icon';
import ArrowRight01Icon from '@hugeicons/core-free-icons/ArrowRight01Icon';
import Cancel01Icon from '@hugeicons/core-free-icons/Cancel01Icon';
import Alert02Icon from '@hugeicons/core-free-icons/Alert02Icon';
import Tick02Icon from '@hugeicons/core-free-icons/Tick02Icon';
import Globe02Icon from '@hugeicons/core-free-icons/Globe02Icon';
import ReloadIcon from '@hugeicons/core-free-icons/ReloadIcon';
import Compass01Icon from '@hugeicons/core-free-icons/Compass01Icon';
import Add01Icon from '@hugeicons/core-free-icons/Add01Icon';
import Image01Icon from '@hugeicons/core-free-icons/Image01Icon';
import UserGroupIcon from '@hugeicons/core-free-icons/UserGroupIcon';
import ArrowDown01Icon from '@hugeicons/core-free-icons/ArrowDown01Icon';
import LockKeyIcon from '@hugeicons/core-free-icons/LockKeyIcon';
import PencilEdit02Icon from '@hugeicons/core-free-icons/PencilEdit02Icon';

import QrCodeScanIcon from '@hugeicons/core-free-icons/QrCodeScanIcon';
import Camera01Icon from '@hugeicons/core-free-icons/Camera01Icon';
const icons = { scan: QrCodeScanIcon, camera: Camera01Icon, search: Search01Icon, calendar: Calendar03Icon, location: Location01Icon, arrow: ArrowRight01Icon, external: ArrowUpRight01Icon, close: Cancel01Icon, alert: Alert02Icon, check: Tick02Icon, globe: Globe02Icon, reload: ReloadIcon, discover: Compass01Icon, add: Add01Icon, image: Image01Icon, people: UserGroupIcon, down: ArrowDown01Icon, lock: LockKeyIcon, edit: PencilEdit02Icon };
export function Icon({ name, size = 20, className = '' }: { name: keyof typeof icons; size?: number; className?: string }) {
  return <HugeiconsIcon icon={icons[name]} size={size} strokeWidth={1.8} className={className} aria-hidden="true" focusable="false" />;
}
