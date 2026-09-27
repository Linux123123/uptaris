import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { Activity, ArrowRight, Server, ShieldAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { StatusBadge } from "@/components/status-badge";
import { serversQuery, dashboardQuery } from "@/lib/queries";

export const Route = createFileRoute("/app/dashboard")({
  loader: async ({ context }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(dashboardQuery()),
      context.queryClient.ensureQueryData(serversQuery({ page: 1, pageSize: 5 })),
    ]);
  },
  component: DashboardPage,
});

function DashboardPage() {
  const status = useQuery(dashboardQuery());
  const servers = useQuery(serversQuery({ page: 1, pageSize: 5 }));
  const metrics = [
    { label: "Servers", value: status.data?.servers, icon: Server, detail: "in inventory" },
    {
      label: "Monitors",
      value: status.data?.monitors,
      icon: Activity,
      detail: "configured checks",
    },
    {
      label: "Open incidents",
      value: status.data?.openIncidents,
      icon: ShieldAlert,
      detail: "need attention",
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Operations overview"
        description="Current infrastructure state and recent inventory at a glance."
        action={
          <Button
            nativeButton={false}
            render={<Link to="/app/servers" search={{ page: 1, pageSize: 20 }} />}
          >
            View servers
            <ArrowRight />
          </Button>
        }
      />
      {status.isError && (
        <ErrorState message={status.error.message} retry={() => void status.refetch()} />
      )}
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {metrics.map(({ label, value, icon: Icon, detail }) => (
          <Card key={label}>
            <CardHeader>
              <CardDescription>{label}</CardDescription>
              <CardTitle className="flex items-center justify-between text-3xl">
                <span>{status.isPending ? "—" : value}</span>
                <span className="grid size-9 place-items-center rounded-lg bg-muted">
                  <Icon className="size-4 text-muted-foreground" />
                </span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xs text-muted-foreground">{detail}</p>
            </CardContent>
          </Card>
        ))}
      </div>
      <Card>
        <CardHeader className="border-b">
          <CardTitle>Recent servers</CardTitle>
          <CardDescription>Most recently added accessible infrastructure.</CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          {servers.isPending ? (
            <LoadingState />
          ) : servers.isError ? (
            <ErrorState message={servers.error.message} retry={() => void servers.refetch()} />
          ) : servers.data.data.length === 0 ? (
            <div className="p-6 text-sm text-muted-foreground">No servers available.</div>
          ) : (
            <div className="divide-y">
              {servers.data.data.map((server) => (
                <Link
                  key={server.id}
                  to="/app/servers/$serverId"
                  params={{ serverId: server.id }}
                  search={{ page: 1, pageSize: 20 }}
                  className="flex items-center gap-4 px-4 py-3 transition-colors hover:bg-muted/50"
                >
                  <span className="grid size-9 place-items-center rounded-lg border bg-muted/30">
                    <Server className="size-4" />
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block truncate font-medium">{server.name}</span>
                    <span className="block truncate font-mono text-xs text-muted-foreground">
                      {server.address}
                    </span>
                  </span>
                  <StatusBadge value={server.status} />
                  <ArrowRight className="size-4 text-muted-foreground" />
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
