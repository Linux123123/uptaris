import { FormFieldLayout } from "@/components/forms/form-field-layout";
import { useFormField } from "@/hooks/use-form-field";
import { Input } from "@/components/ui/input";
import type { ComponentProps } from "react";

type TextFieldProps = {
  label: string;
  type?: ComponentProps<typeof Input>["type"];
  autoComplete?: string;
  placeholder?: string;
  reserveErrorSpace?: boolean;
};

export function FormTextField({
  label,
  type = "text",
  autoComplete,
  placeholder,
  reserveErrorSpace,
}: TextFieldProps) {
  const { field, inputId, invalid } = useFormField<string>();

  return (
    <FormFieldLayout
      label={label}
      inputId={inputId}
      invalid={invalid}
      errors={field.state.meta.errors}
      reserveErrorSpace={reserveErrorSpace}
    >
      <Input
        id={inputId}
        name={field.name}
        type={type}
        autoComplete={autoComplete}
        placeholder={placeholder}
        value={field.state.value}
        onBlur={field.handleBlur}
        onChange={(event) => field.handleChange(event.target.value)}
        aria-invalid={invalid}
      />
    </FormFieldLayout>
  );
}
