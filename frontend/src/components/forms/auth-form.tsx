import { browserSupportsWebAuthn, startAuthentication } from "@simplewebauthn/browser";
import { useQuery } from "@tanstack/react-query";
import { Link, useNavigate, useRouter } from "@tanstack/react-router";
import { useSelector } from "@tanstack/react-store";
import { KeyRound } from "lucide-react";
import { useState } from "react";
import { z } from "zod";
import { OAuthProviderLogo } from "@/components/oauth-provider-logo";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { useAppForm } from "@/hooks/use-app-form";
import { ApiError, authApi, setAccessToken, type AuthResponse } from "@/lib/api";
import { authStore, setAuthenticated } from "@/lib/auth-store";
import { passwordChecks } from "@/lib/password-policy";

const credentialsSchema = z
  .object({
    email: z.email("Enter valid email address"),
    password: z.string(),
  })
  .superRefine(({ password }, context) => {
    for (const check of passwordChecks(password)) {
      if (!check.met) {
        context.addIssue({ code: "custom", path: ["password"], message: check.label });
      }
    }
  });

export function AuthForm({
  register,
  registered = false,
  redirectTo,
  oauthError,
}: {
  register: boolean;
  registered?: boolean;
  redirectTo?: string;
  oauthError?: string;
}) {
  const navigate = useNavigate();
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const [passkeyPending, setPasskeyPending] = useState(false);
  const passkeySupported = browserSupportsWebAuthn();
  const { data: oauthProviders = [] } = useQuery({
    queryKey: ["oauth-providers"],
    queryFn: authApi.providers,
    staleTime: 5 * 60 * 1000,
  });
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

        if ("twoFactorRequired" in result) {
          await navigate({ to: "/two-factor", search: { redirect: redirectTo }, replace: true });

          return;
        }

        await completeSignIn(result);
      } catch (error) {
        setServerError(error instanceof ApiError ? error.message : "Request failed. Try again.");
      }
    },
  });

  async function completeSignIn(result: AuthResponse) {
    setAccessToken(result.accessToken);
    setAuthenticated(result.user);
    await router.invalidate();
    await navigate({ to: redirectTo ?? "/app/dashboard", replace: true });
  }

  async function signInWithPasskey() {
    setPasskeyPending(true);
    setServerError("");

    try {
      const { ceremonyToken, optionsJSON } = await authApi.beginPasskeyLogin();
      const credential = await startAuthentication({ optionsJSON });
      const result = await authApi.finishPasskeyLogin(ceremonyToken, credential);
      await completeSignIn(result);
    } catch (error) {
      setServerError(
        error instanceof Error && error.name === "NotAllowedError"
          ? "Passkey request cancelled or timed out."
          : error instanceof Error
            ? error.message
            : "Passkey sign-in failed. Try again.",
      );
    } finally {
      setPasskeyPending(false);
    }
  }

  return (
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
              <Alert>
                <AlertDescription>
                  {registered
                    ? "Account created. Sign in with your new credentials."
                    : "Session expired. Sign in again to continue."}
                </AlertDescription>
              </Alert>
            )}
            {oauthError && (
              <Alert variant="destructive">
                <AlertDescription>
                  {oauthError === "email_exists"
                    ? "An account already uses this email. Sign in with that account, then connect this provider in account settings."
                    : oauthError === "cancelled"
                      ? "Provider sign-in was cancelled."
                      : "Provider sign-in failed. Try again."}
                </AlertDescription>
              </Alert>
            )}
            {!register && !passkeySupported && (
              <p className="text-sm text-muted-foreground">
                Passkey sign-in is unavailable in this browser.
              </p>
            )}
            {(oauthProviders.length > 0 || (!register && passkeySupported)) && (
              <div className="grid gap-5">
                {!register && passkeySupported && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="w-full"
                    disabled={passkeyPending}
                    onClick={signInWithPasskey}
                  >
                    <KeyRound />
                    {passkeyPending ? "Waiting for passkey…" : "Sign in with passkey"}
                  </Button>
                )}
                {oauthProviders.length > 0 && (
                  <div className="grid auto-cols-fr grid-flow-col gap-2">
                    {oauthProviders.map((provider) => (
                      <Button
                        key={provider.id}
                        type="button"
                        variant="outline"
                        size="sm"
                        className="w-full min-w-0 gap-2 min-h-8 cursor-pointer"
                        aria-label={`Continue with ${provider.name}`}
                        onClick={() => window.location.assign(authApi.oauthLoginURL(provider))}
                      >
                        <OAuthProviderLogo provider={provider.id} />
                        {provider.name}
                      </Button>
                    ))}
                  </div>
                )}
                <div className="flex items-center gap-3">
                  <Separator className="flex-1" />
                  <span className="shrink-0 text-xs text-muted-foreground">
                    Or continue with email
                  </span>
                  <Separator className="flex-1" />
                </div>
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
                <field.PasswordField
                  label="Password"
                  showRequirements={register}
                  autoComplete={register ? "new-password" : "current-password"}
                />
              )}
            </form.AppField>
            {serverError && (
              <Alert variant="destructive">
                <AlertDescription>{serverError}</AlertDescription>
              </Alert>
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
  );
}
