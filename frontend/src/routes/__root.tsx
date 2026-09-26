import { createRootRouteWithContext } from "@tanstack/react-router";
import type { QueryClient } from "@tanstack/react-query";
import type { AuthState } from "@/lib/auth-store";
import { RootLayout } from "@/components/layout/root-layout";
import { NotFoundPage } from "@/components/not-found-page";
import { ApplicationErrorPage } from "@/components/application-error-page";

export const Route = createRootRouteWithContext<{ queryClient: QueryClient; auth: AuthState }>()({
  component: RootLayout,
  notFoundComponent: NotFoundPage,
  errorComponent: ApplicationErrorPage,
});
