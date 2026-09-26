import { Link, useNavigate, useRouter } from "@tanstack/react-router";
import { useSelector } from "@tanstack/react-store";
import { useState } from "react";
import { z } from "zod";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAppForm } from "@/hooks/use-app-form";
import { ApiError, authApi, setAccessToken } from "@/lib/api";
import { authStore, setAuthenticated } from "@/lib/auth-store";

const credentialsSchema = z.object({
  email: z.email("Enter valid email address"),
  password: z.string().min(8, "Password must contain at least 8 characters").max(128),
});

export function AuthPage({
  register,
  registered = false,
  redirectTo,
}: {
  register: boolean;
  registered?: boolean;
  redirectTo?: string;
}) {
  const navigate = useNavigate();
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const sessionExpired = useSelector(authStore, (state) => state.expired);
  const form = useAppForm({
    defaultValues: { email: "", password: "" },
    validators: { onChange: credentialsSchema },
    onSubmit: async ({ value }) => {
      setServerError("");
      try {
        if (register) {
          await authApi.register(value.email, value.password);
          await navigate({ to: "/login", search: { registered: true }, replace: true });

          return;
        }
        const result = await authApi.login(value.email, value.password);
        setAccessToken(result.accessToken);
        setAuthenticated(result.user);
        await router.invalidate();
        await navigate({ to: redirectTo ?? "/app/dashboard", replace: true });
      } catch (error) {
        setServerError(error instanceof ApiError ? error.message : "Request failed. Try again.");
      }
    },
  });

  return (
    <div className="mx-auto grid w-full max-w-5xl items-center gap-12 lg:min-h-full lg:grid-cols-[minmax(0,1fr)_minmax(0,1.15fr)] xl:gap-20">
      <div className="hidden min-w-0 lg:block">
        <h1 className="text-4xl font-semibold tracking-tight">
          Keep operations calm when systems are not.
        </h1>
        <p className="mt-4 text-lg leading-8 text-muted-foreground">
          One focused console for server health, checks, incidents, and access control.
        </p>
      </div>
      <Card className="w-full min-w-0 max-w-xl justify-self-center px-3 py-6 shadow-2xl shadow-black/20 sm:px-6 sm:py-10 lg:max-w-none">
        <CardHeader>
          <CardTitle>{register ? "Create account" : "Welcome back"}</CardTitle>
          <CardDescription>
            {register ? "Start a new monitoring workspace." : "Sign in to your operations console."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form.AppForm>
            <form
              className="relative flex flex-col gap-7 justify-center h-full"
              onSubmit={(event) => {
                event.preventDefault();
                void form.handleSubmit();
              }}
            >
              {!register && (registered || sessionExpired) && (
                <div className="rounded-md border border-primary/25 bg-primary/10 p-3 text-sm text-foreground">
                  {registered
                    ? "Account created. Sign in with your new credentials."
                    : "Session expired. Sign in again to continue."}
                </div>
              )}
              <form.AppField name="email">
                {(field) => (
                  <field.TextField
                    label="Email"
                    type="email"
                    autoComplete="email"
                    reserveErrorSpace
                  />
                )}
              </form.AppField>
              <form.AppField name="password">
                {(field) => (
                  <field.TextField
                    label="Password"
                    type="password"
                    autoComplete={register ? "new-password" : "current-password"}
                    reserveErrorSpace
                  />
                )}
              </form.AppField>
              {serverError && (
                <div
                  role="alert"
                  aria-live="polite"
                  className="absolute inset-x-0 bottom-36 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive"
                >
                  {serverError}
                </div>
              )}
              <form.SubmitButton
                label={register ? "Create account" : "Sign in"}
                pendingLabel="Please wait…"
              />
              <p className="text-center text-sm text-muted-foreground">
                {register ? "Already have an account?" : "New to Uptaris?"}{" "}
                <Link
                  className="font-medium text-foreground underline underline-offset-4"
                  to={register ? "/login" : "/register"}
                >
                  {register ? "Sign in" : "Create account"}
                </Link>
              </p>
            </form>
          </form.AppForm>
        </CardContent>
      </Card>
    </div>
  );
}
