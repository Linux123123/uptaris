import { createFileRoute, redirect } from "@tanstack/react-router";
import { AuthForm } from "@/components/forms/auth-form";
import { AuthLayout } from "@/components/layout/auth-layout";

export const Route = createFileRoute("/register")({
  beforeLoad: ({ context }) => {
    if (context.auth.status === "authenticated") throw redirect({ to: "/app/dashboard" });
  },
  component: RegisterPage,
});

function RegisterPage() {
  return (
    <AuthLayout>
      <AuthForm register />
    </AuthLayout>
  );
}
