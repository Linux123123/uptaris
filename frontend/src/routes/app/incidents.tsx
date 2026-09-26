import { FilterButton } from "@/components/filter-button";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { ShieldAlert } from "lucide-react";
import { z } from "zod";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { DataTable } from "@/components/data-table";
import { PageHeader } from "@/components/page-header";
import { ErrorState } from "@/components/error-state";
import { LoadingState } from "@/components/loading-state";
import { StatusBadge } from "@/components/status-badge";
import { incidentOverviewQuery } from "@/lib/queries";
import type { IncidentRow, IncidentSeverity, IncidentStatus } from "@/lib/api";

const searchSchema = z.object({
  page: z.coerce.number().int().positive().catch(1),
  pageSize: z.coerce.number().int().min(1).max(100).catch(20),
  status: z.enum(["open", "acknowledged", "resolved"]).optional().catch(undefined),
  severity: z.enum(["low", "medium", "high", "critical"]).optional().catch(undefined),
});

export const Route = createFileRoute("/app/incidents")({
  validateSearch: searchSchema,
  component: IncidentsPage,
});

function IncidentsPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const query = useQuery(incidentOverviewQuery(search));
  const rows = query.data?.data ?? [];
  const pending = query.isPending;
  const failed = query.error;
  const columns: ColumnDef<IncidentRow>[] = [
    {
      accessorKey: "title",
      header: "Incident",
      cell: ({ row }) => (
        <div>
          <p className="font-medium">{row.original.title}</p>
          <p className="max-w-xs truncate text-xs text-muted-foreground">
            {row.original.description || "No description"}
          </p>
        </div>
      ),
    },
    {
      accessorKey: "monitorName",
      header: "Monitor",
      cell: ({ row }) => (
        <Link
          className="hover:underline"
          to="/app/monitors/$monitorId"
          params={{ monitorId: row.original.monitorId }}
          search={{ serverId: row.original.serverId, page: 1, pageSize: 20 }}
        >
          {row.original.monitorName}
        </Link>
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
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Incidents"
        description="Incidents aggregated from every server and monitor you can access."
      />
      <div className="flex flex-wrap gap-2">
        <FilterButton
          label="All statuses"
          active={!search.status}
          onClick={() => void navigate({ search: { ...search, page: 1, status: undefined } })}
        />
        {(["open", "acknowledged", "resolved"] as IncidentStatus[]).map((status) => (
          <FilterButton
            key={status}
            label={status}
            active={search.status === status}
            onClick={() => void navigate({ search: { ...search, page: 1, status } })}
          />
        ))}
        <span className="mx-1 h-8 border-l" />
        {(["low", "medium", "high", "critical"] as IncidentSeverity[]).map((severity) => (
          <FilterButton
            key={severity}
            label={severity}
            active={search.severity === severity}
            onClick={() =>
              void navigate({
                search: {
                  ...search,
                  page: 1,
                  severity: search.severity === severity ? undefined : severity,
                },
              })
            }
          />
        ))}
      </div>
      <Card>
        <CardContent className="p-1">
          {pending ? (
            <LoadingState label="Loading scoped incidents" />
          ) : failed ? (
            <ErrorState message={failed.message} />
          ) : (
            <DataTable
              columns={columns}
              data={rows}
              page={search.page}
              totalPages={query.data?.pagination.totalPages ?? 0}
              onPageChange={(page) => void navigate({ search: { ...search, page } })}
              emptyTitle="No incidents found"
              emptyDescription="No incidents match current scope and filters."
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
