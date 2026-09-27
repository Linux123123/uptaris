import { Spinner } from "@/components/ui/spinner";
import { Button } from "@/components/ui/button";
import { useFormContext } from "@/hooks/form-context";

export function FormSubmitButton({
  label,
  pendingLabel = "Saving…",
  formId,
}: {
  label: string;
  pendingLabel?: string;
  formId?: string;
}) {
  const form = useFormContext();

  return (
    <form.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting] as const}>
      {([canSubmit, isSubmitting]) => (
        <Button type="submit" form={formId} disabled={!canSubmit || isSubmitting}>
          {isSubmitting && <Spinner aria-hidden="true" />}
          {isSubmitting ? pendingLabel : label}
        </Button>
      )}
    </form.Subscribe>
  );
}
