import { requiredText } from "@/lib/form-validation";
import { resourceStatusOptions, monitorTypeOptions } from "@/lib/resource-options";
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
import { ApiError, monitorsApi, type Monitor, type MonitorType } from "@/lib/api";

const schema = z.object({
  status: z.enum(["up", "down", "paused"]),
  name: requiredText("Name"),
  type: z.enum(["http", "tcp", "icmp"]),
  target: requiredText("Target", 2048),
  intervalSeconds: z.number().int().min(1).max(86_400),
  expectedHealth: requiredText("Expected health"),
});

export function MonitorFormDialog({ serverId, monitor }: { serverId: string; monitor?: Monitor }) {
  const formId = useId();
  const [open, setOpen] = useState(false);
  const [serverError, setServerError] = useState("");
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: (value: z.infer<typeof schema>) =>
      monitor
        ? monitorsApi.update(serverId, monitor.id, value)
        : monitorsApi.create(serverId, value),
    onSuccess: async () => {
      await invalidateInventory(queryClient);
      await queryClient.invalidateQueries({ queryKey: ["monitors", serverId] });
      if (monitor)
        await queryClient.invalidateQueries({ queryKey: ["monitor", serverId, monitor.id] });
      toast.success(monitor ? "Monitor updated" : "Monitor created");
      form.reset();
      setOpen(false);
    },
  });
  const form = useAppForm({
    defaultValues: {
      status: monitor?.status ?? ("up" as const),
      name: monitor?.name ?? "",
      type: monitor?.type ?? ("http" as MonitorType),
      target: monitor?.target ?? "",
      intervalSeconds: monitor?.intervalSeconds ?? 60,
      expectedHealth: monitor?.expectedHealth ?? "200",
    },
    validators: { onChange: schema },
    onSubmit: async ({ value }) => {
      setServerError("");
      try {
        await mutation.mutateAsync(value);
      } catch (error) {
        setServerError(error instanceof ApiError ? error.message : "Unable to save monitor");
      }
    },
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button variant={monitor ? "outline" : "default"} />}>
        {!monitor && <Plus />}
        {monitor ? "Edit monitor" : "Add monitor"}
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{monitor ? "Edit monitor" : "Add monitor"}</DialogTitle>
          <DialogDescription>
            Define endpoint and expected healthy response. Uptaris stores configuration only.
          </DialogDescription>
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
              {(field) => <field.SelectField label="Status" options={resourceStatusOptions} />}
            </form.AppField>
            <form.AppField name="name">{(field) => <field.TextField label="Name" />}</form.AppField>
            <form.AppField name="type">
              {(field) => <field.SelectField label="Type" options={monitorTypeOptions} />}
            </form.AppField>
            <form.AppField name="target">
              {(field) => (
                <field.TextField label="Target" placeholder="https://example.com/health" />
              )}
            </form.AppField>
            <div className="grid gap-4 sm:grid-cols-2">
              <form.AppField name="intervalSeconds">
                {(field) => <field.NumberField label="Interval (seconds)" maximum={86_400} />}
              </form.AppField>
              <form.AppField name="expectedHealth">
                {(field) => <field.TextField label="Expected health" />}
              </form.AppField>
            </div>
            {serverError && (
              <p role="alert" className="text-sm text-destructive">
                {serverError}
              </p>
            )}
          </form>
          <DialogFooter showCloseButton>
            <form.SubmitButton formId={formId} label="Save monitor" />
          </DialogFooter>
        </form.AppForm>
      </DialogContent>
    </Dialog>
  );
}
