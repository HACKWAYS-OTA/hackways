export const faqTopics = [
  { id: 'finding', label: 'Finding an event', intro: 'Find a gathering, understand event visibility and check the schedule before making plans.' },
  { id: 'joining', label: 'Registration and status', intro: 'Understand confirmation, applications, waiting-list offers and the result of a registration request.' },
  { id: 'changes', label: 'Your registration', intro: 'Find your ticket and understand the cancellation, note and transfer actions available for your booking.' },
  { id: 'accounts', label: 'Accounts and email', intro: 'Get help with sign-in sessions, verification and transactional email.' },
  { id: 'hosting', label: 'Hosting an event', intro: 'Prepare an event, choose how guests join and manage changes in the organizer workspace.' },
  { id: 'safety', label: 'Safety and support', intro: 'Find the right contact, report a problem and protect private information when asking for help.' },
] as const;
export type FAQTopic = typeof faqTopics[number]['id'];
