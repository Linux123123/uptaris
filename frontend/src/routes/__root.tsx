import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyHeader,
  EmptyTitle,
  EmptyDescription,
  EmptyContent,
} from "@/components/ui/empty";
import { createRootRouteWithContext, Link, type ErrorComponentProps } from "@tanstack/react-router";
import type { QueryClient } from "@tanstack/react-query";
import type { AuthState } from "@/lib/auth-store";
import { RootLayout } from "@/components/layout/root-layout";

export const Route = createRootRouteWithContext<{ queryClient: QueryClient; auth: AuthState }>()({
  component: RootLayout,
  notFoundComponent: NotFoundPage,
  errorComponent: ApplicationErrorPage,
});

function NotFoundPage() {
  return (
    <div className="dark grid min-h-svh place-items-center bg-background p-6 text-center text-foreground">
      <Empty>
        <EmptyHeader>
          <p className="font-mono text-sm text-primary">404</p>
          <EmptyTitle>
            <h1 className="text-3xl">Page not found</h1>
          </EmptyTitle>
          <EmptyDescription>Requested page does not exist.</EmptyDescription>
        </EmptyHeader>
        <EmptyContent>
          <Button nativeButton={false} render={<Link to="/" />}>
            Return to status
          </Button>
        </EmptyContent>
      </Empty>
    </div>
  );
}

function ApplicationErrorPage({ error }: ErrorComponentProps) {
  return (
    <div className="dark grid min-h-svh place-items-center bg-background p-6 text-foreground">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>
            <h1>Something went wrong</h1>
          </CardTitle>
          <CardDescription>
            {error instanceof Error ? error.message : "Unexpected application error."}
          </CardDescription>
        </CardHeader>
        <CardFooter>
          <Button onClick={() => window.location.reload()}>Reload page</Button>
        </CardFooter>
      </Card>
    </div>
  );
}
