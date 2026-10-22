import type { RegistrationQuestion, RegistrationAnswer } from "@/types/api";
export type RegistrationStep = {
  title: string;
  description: string;
  questions: RegistrationQuestion[];
};
export function registrationSteps(
  questions: RegistrationQuestion[],
): RegistrationStep[] {
  const sections: RegistrationStep[] = [];
  let current: RegistrationStep = {
    title: "Registration details",
    description: "",
    questions: [],
  };
  let explicitSection = false;
  for (const question of questions) {
    if (question.kind === "section") {
      if (current.questions.length || current.description || explicitSection)
        sections.push(current);
      explicitSection = true;
      current = {
        title: question.label,
        description: question.description,
        questions: [],
      };
    } else current.questions.push(question);
  }
  if (current.questions.length || current.description || explicitSection)
    sections.push(current);
  if (!questions.some((q) => q.kind === "section") && questions.length > 4)
    return Array.from(
      { length: Math.ceil(questions.length / 4) },
      (_, index) => ({
        title: `Questions ${index * 4 + 1}-${Math.min(questions.length, index * 4 + 4)}`,
        description: "",
        questions: questions.slice(index * 4, index * 4 + 4),
      }),
    );
  return sections;
}
export function answerError(
  q: RegistrationQuestion,
  answer?: RegistrationAnswer,
): string {
  const text = answer?.text.trim() || "",
    selected = answer?.selected || [],
    v = q.validation || {},
    custom = (fallback: string) => v.message || fallback;
  if (q.kind === "checkbox")
    return q.required && !answer?.checked ? "Please confirm this answer." : "";
  if (q.kind === "multiselect")
    return q.required && !selected.length
      ? "Choose at least one option."
      : selected.some((item) => !q.options.includes(item))
        ? "Choose an available option."
        : "";
  if (!text) return q.required ? "This answer is required." : "";
  const length = Array.from(text).length;
  if (v.min_length != null && length < v.min_length)
    return custom(`Use at least ${v.min_length} characters.`);
  if (length > (v.max_length ?? 2000))
    return custom(`Use at most ${v.max_length ?? 2000} characters.`);
  if (q.kind === "number") {
    const value = Number(text);
    if (
      !Number.isFinite(value) ||
      !/^[-+]?(?:\d+\.?\d*|\.\d+)(?:e[-+]?\d+)?$/i.test(text)
    )
      return custom("Enter a valid number.");
    if (v.integer && !Number.isInteger(value))
      return custom("Enter a whole number.");
    if (v.min != null && value < v.min)
      return custom(`Use a number of at least ${v.min}.`);
    if (v.max != null && value > v.max)
      return custom(`Use a number no greater than ${v.max}.`);
  }
  if (q.kind === "email" && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(text))
    return custom("Enter a valid email address.");
  if (q.kind === "url") {
    try {
      const url = new URL(text);
      if (url.protocol !== "https:" || url.username || url.password)
        return custom("Use a full https:// website address.");
    } catch {
      return custom("Use a full https:// website address.");
    }
  }
  if (
    q.kind === "phone" &&
    (!/^[+()\d .-]+$/.test(text) ||
      text.replace(/\D/g, "").length < 5 ||
      text.replace(/\D/g, "").length > 20)
  )
    return custom("Enter a valid phone number.");
  if (q.kind === "select" && !q.options.includes(text))
    return "Choose an available option.";
  if (
    q.kind === "date" &&
    (!/^\d{4}-\d{2}-\d{2}$/.test(text) ||
      Number.isNaN(Date.parse(text)) ||
      new Date(text).toISOString().slice(0, 10) !== text)
  )
    return custom("Choose a valid calendar date.");
  if (q.kind === "time" && !/^([01]\d|2[0-3]):[0-5]\d$/.test(text))
    return custom("Choose a valid time.");
  return "";
}
