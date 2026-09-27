import { createFileRoute, redirect } from "@tanstack/react-router";
import { z } from "zod";
import { AuthForm } from "@/components/forms/auth-form";
import { AuthLayout } from "@/components/layout/auth-layout";

const searchSchema = z.object({
  redirect: z.string().startsWith("/app").optional().catch(undefined),
  registered: z.coerce.boolean().optional().catch(false),
  oauth_error: z.string().optional().catch(undefined),
});

export const Route = createFileRoute("/login")({
  validateSearch: searchSchema,
  beforeLoad: ({ context }) => {
    if (context.auth.status === "authenticated") throw redirect({ to: "/app/dashboard" });
  },
  component: LoginPage,
});

function LoginPage() {
  const { redirect, registered, oauth_error } = Route.useSearch();

  return (
    <AuthLayout>
      <AuthForm
        register={false}
        registered={registered}
        redirectTo={redirect}
        oauthError={oauth_error}
      />
    </AuthLayout>
  );
}
