import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { incidentsApi } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogTrigger,
} from "@/components/ui/dialog";
import { IncidentFormDialog } from "@/components/incident-form-dialog";
import { ErrorState } from "@/components/error-state";
import { LoadingState } from "@/components/loading-state";

export function IncidentDetailsDialog({
  serverId,
  monitorId,
  id,
  title,
  canEdit,
}: {
  serverId: string;
  monitorId: string;
  id: string;
  title: string;
  canEdit: boolean;
}) {
  const [open, setOpen] = useState(false);
  const query = useQuery({
    queryKey: ["incidents", serverId, monitorId, id],
    queryFn: ({ signal }) => incidentsApi.get(serverId, monitorId, id, signal),
    enabled: open,
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        render={<Button variant="link" className="h-auto p-0 text-left whitespace-normal" />}
      >
        {title}
      </DialogTrigger>
      <DialogContent>
        <DialogTitle>Incident details</DialogTitle>
        <DialogDescription>Recorded impact and resolution.</DialogDescription>
        {query.isPending ? (
          <LoadingState />
        ) : query.isError ? (
          <ErrorState message={query.error.message} />
        ) : (
          <div className="space-y-4">
            <h3 className="font-semibold wrap-break-word">{query.data.data.title}</h3>
            <p className="whitespace-pre-wrap wrap-break-word">
              {query.data.data.description || "No description"}
            </p>
            <p>
              {query.data.data.severity} · {query.data.data.status}
            </p>
            <p>Started: {new Date(query.data.data.startedAt).toLocaleString()}</p>
            {query.data.data.resolvedAt && (
              <p>Resolved: {new Date(query.data.data.resolvedAt).toLocaleString()}</p>
            )}
            {canEdit && (
              <IncidentFormDialog
                key={query.data.data.updatedAt}
                serverId={serverId}
                monitorId={monitorId}
                incident={query.data.data}
              />
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
