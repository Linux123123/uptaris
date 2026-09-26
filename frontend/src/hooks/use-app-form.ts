import { createFormHook } from "@tanstack/react-form";
import { FormNumberField } from "@/components/form-number-field";
import { FormSelectField } from "@/components/form-select-field";
import { FormSubmitButton } from "@/components/form-submit-button";
import { FormTextField } from "@/components/form-text-field";
import { FormTextareaField } from "@/components/form-textarea-field";
import { fieldContext, formContext } from "@/hooks/form-context";

export const { useAppForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: {
    TextField: FormTextField,
    TextareaField: FormTextareaField,
    NumberField: FormNumberField,
    SelectField: FormSelectField,
  },
  formComponents: {
    SubmitButton: FormSubmitButton,
  },
});
