import { createFileRoute, redirect } from "@tanstack/react-router";
import { z } from "zod";
import { AuthPage } from "@/components/auth-page";
import { PublicLayout } from "@/components/layout/public-layout";

const searchSchema = z.object({
  redirect: z.string().startsWith("/app").optional().catch(undefined),
  registered: z.coerce.boolean().optional().catch(false),
});
export const Route = createFileRoute("/login")({
  validateSearch: searchSchema,
  beforeLoad: ({ context }) => {
    if (context.auth.status === "authenticated") throw redirect({ to: "/app/dashboard" });
  },
  component: LoginPage,
});
function LoginPage() {
  const { redirect, registered } = Route.useSearch();

  return (
    <PublicLayout>
      <AuthPage register={false} registered={registered} redirectTo={redirect} />
    </PublicLayout>
  );
}
