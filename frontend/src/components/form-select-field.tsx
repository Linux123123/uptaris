import { FormFieldLayout } from "@/components/form-field-layout";
import { useFormField } from "@/hooks/use-form-field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export function FormSelectField({
  label,
  options,
}: {
  label: string;
  options: ReadonlyArray<{ value: string; label: string }>;
}) {
  const { field, inputId, invalid } = useFormField<string>();

  return (
    <FormFieldLayout
      label={label}
      inputId={inputId}
      invalid={invalid}
      errors={field.state.meta.errors}
    >
      <Select
        value={field.state.value}
        onValueChange={(value) => value && field.handleChange(value)}
      >
        <SelectTrigger id={inputId} className="w-full" aria-invalid={invalid}>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </FormFieldLayout>
  );
}
