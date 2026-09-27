import { FormFieldLayout } from "@/components/forms/form-field-layout";
import { PasswordInput } from "@/components/forms/password-input";
import { useFormField } from "@/hooks/use-form-field";

export function FormPasswordField({
  label,
  showRequirements = false,
  autoComplete = "new-password",
}: {
  label: string;
  showRequirements?: boolean;
  autoComplete?: string;
}) {
  const { field, inputId, invalid } = useFormField<string>();

  return (
    <FormFieldLayout
      label={label}
      inputId={inputId}
      invalid={invalid && !showRequirements}
      errors={field.state.meta.errors}
    >
      <PasswordInput
        id={inputId}
        name={field.name}
        value={field.state.value}
        showRequirements={showRequirements}
        autoComplete={autoComplete}
        onBlur={field.handleBlur}
        onChange={(event) => field.handleChange(event.target.value)}
        aria-invalid={invalid}
        required
      />
    </FormFieldLayout>
  );
}
