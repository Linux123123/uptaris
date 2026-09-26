import { requiredText, descriptionSchema } from "@/lib/form-validation";
import { incidentStatusOptions, incidentSeverityOptions } from "@/lib/resource-options";
import { invalidateInventory } from "@/lib/invalidate-inventory";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { useId, useState } from "react";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useAppForm } from "@/hooks/use-app-form";
import {
  ApiError,
  incidentsApi,
  type Incident,
  type IncidentSeverity,
  type IncidentStatus,
} from "@/lib/api";

const schema = z.object({
  status: z.enum(["open", "acknowledged", "resolved"]),
  title: requiredText("Title"),
  description: descriptionSchema,
  severity: z.enum(["low", "medium", "high", "critical"]),
});

export function IncidentFormDialog({
  serverId,
  monitorId,
  incident,
}: {
  serverId: number;
  monitorId: number;
  incident?: Incident;
}) {
  const formId = useId();
  const [open, setOpen] = useState(false);
  const [serverError, setServerError] = useState("");
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: (value: z.infer<typeof schema>) =>
      incident
        ? incidentsApi.update(serverId, monitorId, incident.id, value)
        : incidentsApi.create(serverId, monitorId, value),
    onSuccess: async () => {
      await invalidateInventory(queryClient);
      await queryClient.invalidateQueries({ queryKey: ["incidents", serverId, monitorId] });
      toast.success(incident ? "Incident updated" : "Incident created");
      form.reset();
      setOpen(false);
    },
  });
  const form = useAppForm({
    defaultValues: {
      title: incident?.title ?? "",
      description: incident?.description ?? "",
      severity: incident?.severity ?? ("medium" as IncidentSeverity),
      status: incident?.status ?? ("open" as IncidentStatus),
    },
    validators: { onChange: schema },
    onSubmit: async ({ value }) => {
      setServerError("");
      try {
        await mutation.mutateAsync(value);
      } catch (error) {
        setServerError(error instanceof ApiError ? error.message : "Unable to save incident");
      }
    },
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        {!incident && <Plus />}
        {incident ? "Edit incident" : "Add incident"}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{incident ? "Edit incident" : "Add incident"}</DialogTitle>
          <DialogDescription>Record operational impact for this monitor.</DialogDescription>
        </DialogHeader>
        <form.AppForm>
          <form
            id={formId}
            className="grid gap-4"
            onSubmit={(event) => {
              event.preventDefault();
              void form.handleSubmit();
            }}
          >
            <form.AppField name="status">
              {(field) => <field.SelectField label="Status" options={incidentStatusOptions} />}
            </form.AppField>
            <form.AppField name="title">
              {(field) => <field.TextField label="Title" />}
            </form.AppField>
            <form.AppField name="severity">
              {(field) => <field.SelectField label="Severity" options={incidentSeverityOptions} />}
            </form.AppField>
            <form.AppField name="description">
              {(field) => <field.TextareaField label="Description" />}
            </form.AppField>
            {serverError && (
              <p role="alert" className="text-sm text-destructive">
                {serverError}
              </p>
            )}
          </form>
          <DialogFooter showCloseButton>
            <form.SubmitButton formId={formId} label="Save incident" />
          </DialogFooter>
        </form.AppForm>
      </DialogContent>
    </Dialog>
  );
}
