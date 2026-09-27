import { useQuery } from "@tanstack/react-query";
import { Activity, ArrowRight, Server, ShieldAlert } from "lucide-react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { PublicLayout } from "@/components/layout/public-layout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { statusQuery } from "@/lib/queries";

export const Route = createFileRoute("/")({ component: PublicStatusPage });

function PublicStatusPage() {
  const query = useQuery(statusQuery());
  const metrics = [
    { key: "servers", label: "Servers", icon: Server },
    { key: "monitors", label: "Monitors", icon: Activity },
    { key: "openIncidents", label: "Open incidents", icon: ShieldAlert },
  ] as const;

  return (
    <PublicLayout>
      <section className="space-y-10">
        <div className="grid items-center gap-8 lg:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)] lg:gap-12">
          <div className="max-w-3xl space-y-6">
            <h1 className="text-4xl font-semibold tracking-tight sm:text-6xl">
              Infrastructure health without noise.
            </h1>
            <p className="max-w-2xl text-lg leading-8 text-muted-foreground">
              Easily monitor your infrastructure and get notified when something goes wrong.&nbsp;
              <span className="font-bold">No more guessing.</span>
            </p>

            <Button nativeButton={false} size="lg" render={<Link to="/login" />}>
              Open console
              <ArrowRight />
            </Button>
          </div>
          <img
            src="/assets/system-topology.svg"
            alt="Diagram of servers connected to monitoring checks"
            width={640}
            height={440}
            className="mx-auto h-auto w-full max-w-xl"
          />
        </div>
        {query.isPending ? (
          <LoadingState label="Loading public status" />
        ) : query.isError ? (
          <ErrorState message={query.error.message} retry={() => void query.refetch()} />
        ) : (
          <div className="grid gap-4 sm:grid-cols-3">
            {metrics.map(({ key, label, icon: Icon }) => (
              <Card key={key}>
                <CardHeader>
                  <div className="mb-4 grid size-9 place-items-center rounded-lg border bg-muted/40">
                    <Icon className="size-4" />
                  </div>
                  <CardTitle className="text-4xl">{query.data[key]}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground">{label}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>
    </PublicLayout>
  );
}
