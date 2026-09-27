import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider, createRouter } from "@tanstack/react-router";
import { useSelector } from "@tanstack/react-store";
import { queryClient } from "@/lib/query-client";
import { authStore } from "@/lib/auth-store";
import { onSessionExpired } from "@/lib/api";
import { Spinner } from "@/components/ui/spinner";
import { AppProviders } from "@/providers/app-providers";
import { routeTree } from "./routeTree.gen";
import "@/index.css";

const router = createRouter({
  routeTree,
  context: { queryClient, auth: authStore.state },
  defaultPreload: "intent",
  defaultPreloadStaleTime: 0,
  scrollRestoration: true,
});

onSessionExpired(() => {
  const redirect = router.state.location.href;
  queryClient.clear();
  authStore.setState(() => ({ status: "anonymous", user: null, expired: true }));
  void router.navigate({ to: "/login", search: { redirect }, replace: true });
});
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

function RouterHost() {
  const auth = useSelector(authStore, (state) => state);

  if (auth.status === "booting") {
    return (
      <div className="dark grid min-h-svh place-items-center bg-background text-foreground">
        <div className="flex items-center gap-4">
          <img src="/assets/logo-full-dark.svg" alt="Uptaris" className="h-7 w-auto sm:h-9" />
          <Spinner className="text-muted-foreground" />
        </div>
      </div>
    );
  }

  return <RouterProvider router={router} context={{ queryClient, auth }} />;
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppProviders>
      <RouterHost />
    </AppProviders>
  </StrictMode>,
);
