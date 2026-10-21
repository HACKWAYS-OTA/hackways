import { HugeiconsIcon } from "@hugeicons/react";
import CodeIcon from "@hugeicons/core-free-icons/CodeIcon";
import Mic01Icon from "@hugeicons/core-free-icons/Mic01Icon";
import UserGroupIcon from "@hugeicons/core-free-icons/UserGroupIcon";

const symbols = { hackathon: CodeIcon, conference: Mic01Icon, meetup: UserGroupIcon };

export function EventStorySymbol({ kind }: { kind: keyof typeof symbols }) {
  return <span className="event-story-symbol" aria-hidden="true"><HugeiconsIcon data-draw-icon icon={symbols[kind]} size={30} strokeWidth={1.5} /></span>;
}
