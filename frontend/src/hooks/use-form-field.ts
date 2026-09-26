import { useId } from "react";
import { useFieldContext } from "@/hooks/form-context";

export function useFormField<Value>() {
  const field = useFieldContext<Value>();
  const inputId = useId();
  const invalid = field.state.meta.isDirty && !field.state.meta.isValid;

  return { field, inputId, invalid };
}
