import { Alert, AlertDescription } from "@/components/ui/alert";
import { requiredText, descriptionSchema } from "@/lib/form-validation";
import { resourceStatusOptions } from "@/lib/resource-options";
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
import { ApiError, serversApi, type Server } from "@/lib/api";

const schema = z.object({
  status: z.enum(["up", "down", "paused"]),
  name: requiredText("Name"),
  address: requiredText("Address", 2048),
  operatingSystem: requiredText("Operating system"),
  description: descriptionSchema,
});

export function ServerFormDialog({ server }: { server?: Server }) {
  const formId = useId();
  const [open, setOpen] = useState(false);
  const [serverError, setServerError] = useState("");
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: (value: z.infer<typeof schema>) =>
      server ? serversApi.update(server.id, value) : serversApi.create(value),
    onSuccess: async () => {
      await invalidateInventory(queryClient);
      await queryClient.invalidateQueries({ queryKey: ["servers"] });

      if (server) await queryClient.invalidateQueries({ queryKey: ["server", server.id] });

      toast.success(server ? "Server updated" : "Server created");
      form.reset();
      setOpen(false);
    },
  });
  const form = useAppForm({
    defaultValues: {
      status: server?.status ?? ("up" as const),
      name: server?.name ?? "",
      address: server?.address ?? "",
      operatingSystem: server?.operatingSystem ?? "",
      description: server?.description ?? "",
    },
    validators: { onChange: schema },
    onSubmit: async ({ value }) => {
      setServerError("");

      try {
        await mutation.mutateAsync(value);
      } catch (error) {
        setServerError(error instanceof ApiError ? error.message : "Unable to save server");
      }
    },
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button variant={server ? "outline" : "default"} />}>
        {!server && <Plus />}
        {server ? "Edit server" : "Add server"}
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{server ? "Edit server" : "Add server"}</DialogTitle>
          <DialogDescription>Keep inventory details concise and recognizable.</DialogDescription>
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
            <form.AppField name="address">
              {(field) => <field.TextField label="Address" placeholder="server.example.com" />}
            </form.AppField>
            <form.AppField name="operatingSystem">
              {(field) => <field.TextField label="Operating system" />}
            </form.AppField>
            <form.AppField name="description">
              {(field) => <field.TextareaField label="Description" />}
            </form.AppField>
            {serverError && (
              <Alert variant="destructive">
                <AlertDescription>{serverError}</AlertDescription>
              </Alert>
            )}
          </form>
          <DialogFooter showCloseButton>
            <form.SubmitButton formId={formId} label="Save server" />
          </DialogFooter>
        </form.AppForm>
      </DialogContent>
    </Dialog>
  );
}
