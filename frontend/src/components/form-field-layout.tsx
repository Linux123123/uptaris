import type { ReactNode } from "react";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { fieldErrorMessages } from "@/lib/form-errors";

type FormFieldLayoutProps = {
  label: string;
  inputId: string;
  invalid: boolean;
  errors: unknown[];
  reserveErrorSpace?: boolean;
  children: ReactNode;
};

export function FormFieldLayout({
  label,
  inputId,
  invalid,
  errors,
  reserveErrorSpace = false,
  children,
}: FormFieldLayoutProps) {
  return (
    <Field data-invalid={invalid} className={reserveErrorSpace ? "relative" : undefined}>
      <FieldLabel htmlFor={inputId}>{label}</FieldLabel>
      {children}
      {invalid && (
        <FieldError
          errors={fieldErrorMessages(errors)}
          className={reserveErrorSpace ? "absolute inset-x-0 top-full z-10" : undefined}
        />
      )}
    </Field>
  );
}
