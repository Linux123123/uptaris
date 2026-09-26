import { QueryClientProvider } from "@tanstack/react-query";
import { lazy, Suspense, useEffect, type ReactNode } from "react";
import { restoreSession } from "@/lib/api";
import { setAnonymous, setAuthenticated } from "@/lib/auth-store";
import { queryClient } from "@/lib/query-client";

const ReactQueryDevtools = import.meta.env.DEV
  ? lazy(() =>
      import("@tanstack/react-query-devtools").then((module) => ({
        default: module.ReactQueryDevtools,
      })),
    )
  : null;
export function AppProviders({ children }: { children: ReactNode }) {
  useEffect(() => {
    void restoreSession().then((user) => (user ? setAuthenticated(user) : setAnonymous()));
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {ReactQueryDevtools && (
        <Suspense>
          <ReactQueryDevtools initialIsOpen={false} />
        </Suspense>
      )}
    </QueryClientProvider>
  );
}
