import { z } from "zod";

export function requiredText(label: string, maximum = 200) {
  return z
    .string()
    .trim()
    .min(1, `${label} is required`)
    .refine(
      (value) => Array.from(value).length <= maximum,
      `${label} must contain at most ${maximum} characters`,
    );
}

export const descriptionSchema = z
  .string()
  .refine(
    (value) => new TextEncoder().encode(value).length <= 10_000,
    "Description must fit within 10,000 UTF-8 bytes",
  );
