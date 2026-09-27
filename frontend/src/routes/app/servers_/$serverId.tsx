import { createFileRoute, Link, stripSearchParams } from "@tanstack/react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { ArrowLeft, Clock3, ExternalLink, MonitorIcon, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DataTable } from "@/components/data-table";
import { DeleteResourceDialog } from "@/components/dialogs/delete-resource-dialog";
import { DetailCard } from "@/components/detail-card";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { MonitorFormDialog } from "@/components/dialogs/monitor-form-dialog";
import { PageHeader } from "@/components/page-header";
import { SelectFilter } from "@/components/select-filter";
import { ServerFormDialog } from "@/components/dialogs/server-form-dialog";
import { StatusBadge } from "@/components/status-badge";
import { authStore } from "@/lib/auth-store";
import { monitorsApi, serversApi, type Monitor } from "@/lib/api";
import { invalidateInventory } from "@/lib/invalidate-inventory";
import { monitorsQuery, serverQuery } from "@/lib/queries";
import { paginationDefaults } from "@/lib/pagination";
import { monitorTypeOptions, resourceStatusOptions } from "@/lib/resource-options";

export const Route = createFileRoute("/app/servers_/$serverId")({
  params: {
    parse: (params) => ({ serverId: z.string().regex(/^\d+$/).parse(params.serverId) }),
  },
  validateSearch: z.object({
    page: z.coerce.number().int().positive().catch(paginationDefaults.page),
    pageSize: z.coerce.number().int().min(1).max(100).catch(paginationDefaults.pageSize),
    status: z.enum(["up", "down", "paused"]).optional().catch(undefined),
    type: z.enum(["http", "tcp", "icmp"]).optional().catch(undefined),
  }),
  search: { middlewares: [stripSearchParams(paginationDefaults)] },
  loaderDeps: ({ search }) => search,
  loader: async ({ context, params, deps }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(serverQuery(params.serverId)),
      context.queryClient.ensureQueryData(monitorsQuery(params.serverId, deps)),
    ]);
  },
  component: ServerDetailPage,
});

function ServerDetailPage() {
  const { serverId } = Route.useParams();
  const navigate = Route.useNavigate();
  const search = Route.useSearch();
  const queryClient = useQueryClient();
  const server = useQuery(serverQuery(serverId));
  const monitors = useQuery(monitorsQuery(serverId, search));
  const canEdit = authStore.state.user?.role !== "viewer";
  const removeServer = useMutation({
    onError: (error) => toast.error(error.message),
    mutationFn: () => serversApi.remove(serverId),
    onSuccess: async () => {
      await invalidateInventory(queryClient);
      await queryClient.invalidateQueries({ queryKey: ["servers"] });
      toast.success("Server deleted");
      await navigate({ to: "/app/servers", search: { page: 1, pageSize: 20 } });
    },
  });
  const removeMonitor = useMutation({
    onError: (error) => toast.error(error.message),
    mutationFn: (monitorId: string) => monitorsApi.remove(serverId, monitorId),
    onSuccess: async () => {
      await invalidateInventory(queryClient);
      await queryClient.invalidateQueries({ queryKey: ["monitors", serverId] });
      toast.success("Monitor deleted");
    },
  });

  if (server.isPending) return <LoadingState label="Loading server" />;

  if (server.isError)
    return <ErrorState message={server.error.message} retry={() => void server.refetch()} />;

  const value = server.data.data;
  const columns: ColumnDef<Monitor>[] = [
    {
      accessorKey: "name",
      header: "Monitor",
      cell: ({ row }) => (
        <div>
          <Link
            className="font-medium hover:underline"
            to="/app/servers/$serverId/monitors/$monitorId"
            params={{ serverId, monitorId: row.original.id }}
            search={{ page: 1, pageSize: 20 }}
          >
            {row.original.name}
          </Link>
          <p className="font-mono text-xs text-muted-foreground">{row.original.target}</p>
        </div>
      ),
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <span className="uppercase text-muted-foreground">{row.original.type}</span>
      ),
    },
    {
      accessorKey: "intervalSeconds",
      header: "Interval",
      cell: ({ row }) => (
        <span className="flex items-center gap-1">
          <Clock3 className="size-3" />
          {row.original.intervalSeconds}s
        </span>
      ),
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => <StatusBadge value={row.original.status} />,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <div className="flex justify-end gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            nativeButton={false}
            render={
              <Link
                to="/app/servers/$serverId/monitors/$monitorId"
                params={{ serverId, monitorId: row.original.id }}
                search={{ page: 1, pageSize: 20 }}
                aria-label={`Open ${row.original.name}`}
              />
            }
          >
            <ExternalLink />
          </Button>
          {canEdit && (
            <DeleteResourceDialog
              trigger={
                <Button variant="ghost" size="icon-sm" aria-label={`Delete ${row.original.name}`}>
                  <Trash2 />
                </Button>
              }
              title={<>Delete monitor?</>}
              description={<>This removes the monitor and its incidents.</>}
              actionLabel="Delete monitor"
              onConfirm={() => removeMonitor.mutateAsync(row.original.id)}
            />
          )}
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <Button
        nativeButton={false}
        variant="ghost"
        size="sm"
        className="-ml-2"
        render={<Link to="/app/servers" search={{ page: 1, pageSize: 20 }} />}
      >
        <ArrowLeft />
        All servers
      </Button>
      <PageHeader
        title={value.name}
        description={value.description || `${value.address} · ${value.operatingSystem}`}
        action={
          canEdit ? (
            <div className="flex gap-2">
              <ServerFormDialog server={value} />
              <DeleteResourceDialog
                trigger={
                  <Button variant="destructive">
                    <Trash2 />
                    Delete
                  </Button>
                }
                title={<>Delete {value.name}?</>}
                description={
                  <>
                    This removes server, all monitors, and all incidents. Action cannot be undone.
                  </>
                }
                actionLabel="Delete server"
                onConfirm={() => removeServer.mutateAsync()}
              />
            </div>
          ) : undefined
        }
      />
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <DetailCard label="Status">
          <StatusBadge value={value.status} />
        </DetailCard>
        <DetailCard label="Address">
          <span className="font-mono text-xs">{value.address}</span>
        </DetailCard>
        <DetailCard label="Operating system">{value.operatingSystem}</DetailCard>
        <DetailCard label="Owner ID">#{value.ownerId}</DetailCard>
      </div>
      <Card>
        <CardHeader className="border-b">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex flex-wrap items-center gap-3">
              <CardTitle className="flex items-center gap-2">
                <MonitorIcon className="size-4" />
                Monitors
              </CardTitle>
              <SelectFilter
                label="All statuses"
                value={search.status}
                options={resourceStatusOptions}
                onChange={(status) => void navigate({ search: { ...search, page: 1, status } })}
              />
              <SelectFilter
                label="All types"
                value={search.type}
                options={monitorTypeOptions}
                onChange={(type) => void navigate({ search: { ...search, page: 1, type } })}
              />
            </div>
            {canEdit && <MonitorFormDialog serverId={serverId} />}
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {monitors.isPending ? (
            <LoadingState />
          ) : monitors.isError ? (
            <ErrorState message={monitors.error.message} retry={() => void monitors.refetch()} />
          ) : (
            <DataTable
              columns={columns}
              data={monitors.data.data}
              page={search.page}
              pageSize={search.pageSize}
              totalPages={monitors.data.pagination.totalPages}
              onPageChange={(page) => void navigate({ search: { ...search, page } })}
              onPageSizeChange={(pageSize) =>
                void navigate({ search: { ...search, page: 1, pageSize } })
              }
              emptyTitle="No monitors configured"
              emptyDescription="Add first HTTP, TCP, or ICMP monitor."
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
