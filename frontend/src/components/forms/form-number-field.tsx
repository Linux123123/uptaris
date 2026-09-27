import { FormFieldLayout } from "@/components/forms/form-field-layout";
import { useFormField } from "@/hooks/use-form-field";
import { Input } from "@/components/ui/input";

export function FormNumberField({
  label,
  minimum = 1,
  maximum,
}: {
  label: string;
  minimum?: number;
  maximum?: number;
}) {
  const { field, inputId, invalid } = useFormField<number>();

  return (
    <FormFieldLayout
      label={label}
      inputId={inputId}
      invalid={invalid}
      errors={field.state.meta.errors}
    >
      <Input
        id={inputId}
        name={field.name}
        type="number"
        min={minimum}
        max={maximum}
        value={Number.isNaN(field.state.value) ? "" : field.state.value}
        onBlur={field.handleBlur}
        onChange={(event) => field.handleChange(event.target.valueAsNumber)}
        aria-invalid={invalid}
      />
    </FormFieldLayout>
  );
}
