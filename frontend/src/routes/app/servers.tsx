import { FilterButton } from "@/components/filter-button";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { ChevronRight, ServerIcon } from "lucide-react";
import { z } from "zod";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DataTable } from "@/components/data-table";
import { ErrorState } from "@/components/error-state";
import { LoadingState } from "@/components/loading-state";
import { PageHeader } from "@/components/page-header";
import { ServerFormDialog } from "@/components/server-form-dialog";
import { StatusBadge } from "@/components/status-badge";
import { authStore } from "@/lib/auth-store";
import { serversQuery } from "@/lib/queries";
import type { Server, ServerStatus } from "@/lib/api";

const searchSchema = z.object({
  page: z.coerce.number().int().positive().catch(1),
  pageSize: z.coerce.number().int().min(10).max(100).catch(20),
  status: z.enum(["up", "down", "paused"]).optional().catch(undefined),
});
export const Route = createFileRoute("/app/servers")({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => search,
  loader: ({ context, deps }) => context.queryClient.ensureQueryData(serversQuery(deps)),
  component: ServersPage,
});

const columns: ColumnDef<Server>[] = [
  {
    accessorKey: "name",
    header: "Server",
    cell: ({ row }) => (
      <div>
        <Link
          className="font-medium hover:underline"
          to="/app/servers/$serverId"
          params={{ serverId: row.original.id }}
          search={{ page: 1, pageSize: 20 }}
        >
          {row.original.name}
        </Link>
        <p className="text-xs text-muted-foreground md:hidden">{row.original.address}</p>
      </div>
    ),
  },
  {
    accessorKey: "address",
    header: "Address",
    cell: ({ row }) => <span className="font-mono text-xs">{row.original.address}</span>,
  },
  { accessorKey: "operatingSystem", header: "Operating system" },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => <StatusBadge value={row.original.status} />,
  },
  {
    id: "open",
    header: "",
    cell: ({ row }) => (
      <Link
        aria-label={`Open ${row.original.name}`}
        to="/app/servers/$serverId"
        params={{ serverId: row.original.id }}
        search={{ page: 1, pageSize: 20 }}
      >
        <ChevronRight className="size-4 text-muted-foreground" />
      </Link>
    ),
  },
];

function ServersPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const query = useQuery(serversQuery(search));
  const canEdit = authStore.state.user?.role !== "viewer";

  return (
    <div className="space-y-6">
      <PageHeader
        title="Servers"
        description="Inventory and current state of accessible infrastructure."
        action={canEdit ? <ServerFormDialog /> : undefined}
      />
      <div className="flex flex-wrap gap-2" aria-label="Filter by status">
        {([undefined, "up", "down", "paused"] as (ServerStatus | undefined)[]).map((status) => (
          <FilterButton
            key={status ?? "all"}
            label={status ?? "all"}
            active={search.status === status}
            onClick={() => void navigate({ search: { ...search, page: 1, status } })}
          />
        ))}
      </div>
      <Card>
        <CardContent className="p-1">
          {query.isPending ? (
            <LoadingState label="Loading servers" />
          ) : query.isError ? (
            <ErrorState message={query.error.message} retry={() => void query.refetch()} />
          ) : (
            <DataTable
              columns={columns}
              data={query.data.data}
              page={query.data.pagination.page}
              totalPages={query.data.pagination.totalPages}
              onPageChange={(page) => void navigate({ search: { ...search, page } })}
              emptyTitle="No servers found"
              emptyDescription="Add first server or change current status filter."
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
