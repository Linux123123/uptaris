import { FormFieldLayout } from "@/components/form-field-layout";
import { useFormField } from "@/hooks/use-form-field";
import { Textarea } from "@/components/ui/textarea";

export function FormTextareaField({ label, placeholder }: { label: string; placeholder?: string }) {
  const { field, inputId, invalid } = useFormField<string>();

  return (
    <FormFieldLayout
      label={label}
      inputId={inputId}
      invalid={invalid}
      errors={field.state.meta.errors}
    >
      <Textarea
        id={inputId}
        name={field.name}
        placeholder={placeholder}
        value={field.state.value}
        onBlur={field.handleBlur}
        onChange={(event) => field.handleChange(event.target.value)}
        aria-invalid={invalid}
      />
    </FormFieldLayout>
  );
}
