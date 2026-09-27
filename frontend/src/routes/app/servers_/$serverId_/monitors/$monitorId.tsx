import { createFileRoute, Link, stripSearchParams } from "@tanstack/react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { ArrowLeft, ShieldAlert, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DataTable } from "@/components/data-table";
import { DeleteResourceDialog } from "@/components/dialogs/delete-resource-dialog";
import { DetailCard } from "@/components/detail-card";
import { ErrorState } from "@/components/feedback/error-state";
import { IncidentDetailsDialog } from "@/components/dialogs/incident-details-dialog";
import { IncidentFormDialog } from "@/components/dialogs/incident-form-dialog";
import { LoadingState } from "@/components/feedback/loading-state";
import { MonitorFormDialog } from "@/components/dialogs/monitor-form-dialog";
import { PageHeader } from "@/components/page-header";
import { SelectFilter } from "@/components/select-filter";
import { StatusBadge } from "@/components/status-badge";
import { incidentsApi, type Incident } from "@/lib/api";
import { authStore } from "@/lib/auth-store";
import { invalidateInventory } from "@/lib/invalidate-inventory";
import { incidentsQuery, monitorQuery } from "@/lib/queries";
import { paginationDefaults } from "@/lib/pagination";
import { incidentSeverityOptions, incidentStatusOptions } from "@/lib/resource-options";

const incidentSearchSchema = z.object({
  page: z.coerce.number().int().positive().catch(paginationDefaults.page),
  pageSize: z.coerce.number().int().min(1).max(100).catch(paginationDefaults.pageSize),
  status: z.enum(["open", "acknowledged", "resolved"]).optional().catch(undefined),
  severity: z.enum(["low", "medium", "high", "critical"]).optional().catch(undefined),
});

export const Route = createFileRoute("/app/servers_/$serverId_/monitors/$monitorId")({
  params: {
    parse: (params) => ({
      serverId: z.string().regex(/^\d+$/).parse(params.serverId),
      monitorId: z.string().regex(/^\d+$/).parse(params.monitorId),
    }),
  },
  validateSearch: incidentSearchSchema,
  search: { middlewares: [stripSearchParams(paginationDefaults)] },
  loaderDeps: ({ search }) => search,
  loader: async ({ context, params, deps }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(monitorQuery(params.serverId, params.monitorId)),
      context.queryClient.ensureQueryData(incidentsQuery(params.serverId, params.monitorId, deps)),
    ]);
  },
  component: MonitorDetailPage,
});

function MonitorDetailPage() {
  const { serverId, monitorId } = Route.useParams();
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const queryClient = useQueryClient();
  const monitor = useQuery(monitorQuery(serverId, monitorId));
  const incidents = useQuery(incidentsQuery(serverId, monitorId, search));
  const canEdit = authStore.state.user?.role !== "viewer";
  const removeIncident = useMutation({
    onError: (error) => toast.error(error.message),
    mutationFn: (id: string) => incidentsApi.remove(serverId, monitorId, id),
    onSuccess: async () => {
      await invalidateInventory(queryClient);
      await queryClient.invalidateQueries({ queryKey: ["incidents", serverId, monitorId] });
      toast.success("Incident deleted");
    },
  });

  if (monitor.isPending) return <LoadingState label="Loading monitor" />;

  if (monitor.isError)
    return <ErrorState message={monitor.error.message} retry={() => void monitor.refetch()} />;

  const value = monitor.data.data;
  const columns: ColumnDef<Incident>[] = [
    {
      accessorKey: "title",
      header: "Incident",
      cell: ({ row }) => (
        <div>
          <IncidentDetailsDialog
            serverId={serverId}
            monitorId={monitorId}
            id={row.original.id}
            title={row.original.title}
            canEdit={canEdit}
          />
          <p className="max-w-md truncate text-xs text-muted-foreground">
            {row.original.description || "No description"}
          </p>
        </div>
      ),
    },
    {
      accessorKey: "severity",
      header: "Severity",
      cell: ({ row }) => <StatusBadge value={row.original.severity} />,
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => <StatusBadge value={row.original.status} />,
    },
    {
      accessorKey: "startedAt",
      header: "Started",
      cell: ({ row }) => new Date(row.original.startedAt).toLocaleString(),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) =>
        canEdit ? (
          <DeleteResourceDialog
            trigger={
              <Button variant="ghost" size="icon-sm" aria-label={`Delete ${row.original.title}`}>
                <Trash2 />
              </Button>
            }
            title={<>Delete incident?</>}
            description={<>This removes the incident from active history.</>}
            actionLabel="Delete incident"
            onConfirm={() => removeIncident.mutateAsync(row.original.id)}
          />
        ) : null,
    },
  ];

  return (
    <div className="space-y-6">
      <Button
        nativeButton={false}
        variant="ghost"
        size="sm"
        className="-ml-2"
        render={
          <Link
            to="/app/servers/$serverId"
            params={{ serverId }}
            search={{ page: 1, pageSize: 20 }}
          />
        }
      >
        <ArrowLeft />
        {`Server #${serverId}`}
      </Button>
      <PageHeader
        title={value.name}
        description={`${value.type.toUpperCase()} · ${value.target}`}
        action={canEdit ? <MonitorFormDialog serverId={serverId} monitor={value} /> : undefined}
      />
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <DetailCard label="Status">
          <StatusBadge value={value.status} />
        </DetailCard>
        <DetailCard label="Type">{value.type.toUpperCase()}</DetailCard>
        <DetailCard label="Interval">{value.intervalSeconds} seconds</DetailCard>
        <DetailCard label="Expected">{value.expectedHealth}</DetailCard>
      </div>
      <Card>
        <CardHeader className="border-b">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex flex-wrap items-center gap-3">
              <CardTitle className="flex items-center gap-2">
                <ShieldAlert className="size-4" />
                Incidents
              </CardTitle>
              <SelectFilter
                label="All statuses"
                value={search.status}
                options={incidentStatusOptions}
                onChange={(status) => void navigate({ search: { ...search, page: 1, status } })}
              />
              <SelectFilter
                label="All severities"
                value={search.severity}
                options={incidentSeverityOptions}
                onChange={(severity) => void navigate({ search: { ...search, page: 1, severity } })}
              />
            </div>
            {canEdit && <IncidentFormDialog serverId={serverId} monitorId={monitorId} />}
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {incidents.isPending ? (
            <LoadingState />
          ) : incidents.isError ? (
            <ErrorState message={incidents.error.message} retry={() => void incidents.refetch()} />
          ) : (
            <DataTable
              columns={columns}
              data={incidents.data.data}
              page={search.page}
              pageSize={search.pageSize}
              totalPages={incidents.data.pagination.totalPages}
              onPageChange={(page) => void navigate({ search: { ...search, page } })}
              onPageSizeChange={(pageSize) =>
                void navigate({ search: { ...search, page: 1, pageSize } })
              }
              emptyTitle="No incidents"
              emptyDescription="This monitor has no recorded incidents."
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
