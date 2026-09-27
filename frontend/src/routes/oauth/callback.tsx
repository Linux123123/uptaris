import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { authApi } from "@/lib/api";
import { setAuthenticated } from "@/lib/auth-store";
import { LoadingState } from "@/components/feedback/loading-state";

export const Route = createFileRoute("/oauth/callback")({ component: OAuthCallbackPage });

function OAuthCallbackPage() {
  const navigate = useNavigate();
  useEffect(() => {
    let active = true;
    void authApi
      .me()
      .then((user) => {
        if (!active) return;

        setAuthenticated(user);
        void navigate({ to: "/app/dashboard", replace: true });
      })
      .catch(() => {
        if (active)
          void navigate({ to: "/login", search: { oauth_error: "failed" }, replace: true });
      });

    return () => {
      active = false;
    };
  }, [navigate]);

  return <LoadingState label="Completing sign-in…" />;
}
