import { FormPasswordField } from "@/components/forms/form-password-field";
import { createFormHook } from "@tanstack/react-form";
import { FormNumberField } from "@/components/forms/form-number-field";
import { FormSelectField } from "@/components/forms/form-select-field";
import { FormSubmitButton } from "@/components/forms/form-submit-button";
import { FormTextField } from "@/components/forms/form-text-field";
import { FormTextareaField } from "@/components/forms/form-textarea-field";
import { fieldContext, formContext } from "@/hooks/form-context";

export const { useAppForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: {
    TextField: FormTextField,
    PasswordField: FormPasswordField,
    TextareaField: FormTextareaField,
    NumberField: FormNumberField,
    SelectField: FormSelectField,
  },
  formComponents: {
    SubmitButton: FormSubmitButton,
  },
});
