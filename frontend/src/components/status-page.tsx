import { useQuery } from "@tanstack/react-query";
import { Activity, ArrowRight, Server, ShieldAlert } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ErrorState } from "@/components/error-state";
import { LoadingState } from "@/components/loading-state";
import { statusQuery } from "@/lib/queries";

export function StatusPage() {
  const query = useQuery(statusQuery());
  const metrics = [
    { key: "servers", label: "Servers", icon: Server },
    { key: "monitors", label: "Monitors", icon: Activity },
    { key: "openIncidents", label: "Open incidents", icon: ShieldAlert },
  ] as const;

  return (
    <section className="space-y-10">
      <div className="grid items-end gap-8 lg:grid-cols-[1fr_auto]">
        <div className="max-w-3xl space-y-4">
          <h1 className="text-4xl font-semibold tracking-tight sm:text-6xl">
            Infrastructure health without noise.
          </h1>
          <p className="max-w-2xl text-lg leading-8 text-muted-foreground">
            Public aggregate availability. Addresses, targets, ownership, and operational detail
            remain private.
          </p>
        </div>
        <Button nativeButton={false} size="lg" render={<Link to="/login" />}>
          Open console
          <ArrowRight />
        </Button>
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
  );
}
