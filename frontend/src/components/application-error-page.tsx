import type { ErrorComponentProps } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";

export function ApplicationErrorPage({ error }: ErrorComponentProps) {
  return (
    <div className="dark grid min-h-svh place-items-center bg-background p-6 text-foreground">
      <div className="max-w-md rounded-xl border bg-card p-6 shadow-sm">
        <h1 className="font-semibold">Something went wrong</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          {error instanceof Error ? error.message : "Unexpected application error."}
        </p>
        <Button className="mt-6" onClick={() => window.location.reload()}>
          Reload page
        </Button>
      </div>
    </div>
  );
}
