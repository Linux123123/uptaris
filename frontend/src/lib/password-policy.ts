import { z } from "zod";

export const passwordMinLength = 8;

export const passwordMaxLength = 128;

// Zod checks mirror the backend library rules, including Unicode character categories.
const requirements = [
  {
    label: "8-128 characters",
    schema: z
      .string()
      .transform((value) => Array.from(value))
      .pipe(z.array(z.string()).min(passwordMinLength).max(passwordMaxLength)),
  },
  { label: "One lowercase letter", schema: z.string().regex(/\p{Ll}/u) },
  { label: "One uppercase letter", schema: z.string().regex(/\p{Lu}/u) },
  { label: "One number", schema: z.string().regex(/\p{N}/u) },
  { label: "One symbol", schema: z.string().regex(/[\p{P}\p{S}]/u) },
];

export function passwordChecks(value: string) {
  return requirements.map(({ label, schema }) => ({ label, met: schema.safeParse(value).success }));
}

export function validPassword(value: string) {
  return passwordChecks(value).every((check) => check.met);
}
