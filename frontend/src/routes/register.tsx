import { createFileRoute, redirect } from "@tanstack/react-router";
import { AuthPage } from "@/components/auth-page";
import { PublicLayout } from "@/components/layout/public-layout";

export const Route = createFileRoute("/register")({
  beforeLoad: ({ context }) => {
    if (context.auth.status === "authenticated") throw redirect({ to: "/app/dashboard" });
  },
  component: RegisterPage,
});

function RegisterPage() {
  return (
    <PublicLayout>
      <AuthPage register />
    </PublicLayout>
  );
}
