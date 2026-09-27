import { createFileRoute, useNavigate, useRouter } from "@tanstack/react-router";
import { useId, useState } from "react";
import { z } from "zod";
import { SecondFactorInput } from "@/components/forms/second-factor-input";
import { PublicLayout } from "@/components/layout/public-layout";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldLabel } from "@/components/ui/field";
import { ApiError, authApi, setAccessToken } from "@/lib/api";
import { setAuthenticated } from "@/lib/auth-store";
import { validSecondFactor } from "@/lib/two-factor";

export const Route = createFileRoute("/two-factor")({
  validateSearch: z.object({ redirect: z.string().optional().catch(undefined) }),
  component: TwoFactorPage,
});

function TwoFactorPage() {
  const id = useId();
  const navigate = useNavigate();
  const router = useRouter();
  const { redirect } = Route.useSearch();
  const [code, setCode] = useState("");
  const [backup, setBackup] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  const [expired, setExpired] = useState(false);
  const valid = validSecondFactor(code, backup);

  async function verifySignIn() {
    if (!valid || pending || expired) return;

    setPending(true);
    setError("");

    try {
      const result = await authApi.verifyTwoFactor(code.trim(), backup);
      setCode("");
      setAccessToken(result.accessToken);
      setAuthenticated(result.user);
      await router.invalidate();
      // Keep redirects within the authenticated application.
      await navigate({
        to: redirect?.startsWith("/app/") ? redirect : "/app/dashboard",
        replace: true,
      });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Verification failed. Try again.");

      if (cause instanceof ApiError && cause.code === "invalid_two_factor_challenge")
        setExpired(true);

      setCode("");
    } finally {
      setPending(false);
    }
  }

  async function cancelSignIn() {
    setPending(true);
    setError("");

    try {
      await authApi.cancelTwoFactor();
      await navigate({ to: "/login", search: { redirect }, replace: true });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not cancel sign-in. Try again.");
    } finally {
      setPending(false);
    }
  }

  return (
    <PublicLayout>
      <div className="mx-auto grid min-h-full w-full max-w-5xl items-center gap-10 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] lg:gap-16">
        <div className="max-w-lg space-y-6">
          <div className="space-y-3">
            <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">One more step</h1>
            <p className="max-w-md text-base leading-7 text-muted-foreground">
              Confirm this sign-in with your authenticator app or a saved backup code.
            </p>
          </div>
          <img
            src="/assets/access-flow.svg"
            alt=""
            width={460}
            height={136}
            className="mx-auto h-auto w-full max-w-xs lg:mx-0 lg:max-w-sm"
          />
        </div>
        <Card className="w-full max-w-lg justify-self-center px-2 py-8 shadow-2xl shadow-black/20 sm:px-6 sm:py-10">
          <CardHeader className="gap-2">
            <CardTitle className="text-xl">Verify sign-in</CardTitle>
            <CardDescription>
              {backup
                ? "Enter one of your single-use backup codes."
                : "Enter the six-digit code from your authenticator app."}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form
              className="grid gap-5"
              onSubmit={(event) => {
                event.preventDefault();
                void verifySignIn();
              }}
            >
              <Field className="gap-3">
                <FieldLabel htmlFor={id} className="w-full justify-center text-center">
                  {backup ? "Backup code" : "Authenticator code"}
                </FieldLabel>
                <SecondFactorInput
                  id={id}
                  value={code}
                  onChange={setCode}
                  backup={backup}
                  large
                  disabled={pending || expired}
                  autoFocus
                />
              </Field>
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}
              <Button type="submit" size="lg" disabled={!valid || pending || expired}>
                {pending ? "Verifying…" : "Verify sign-in"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                disabled={pending || expired}
                onClick={() => {
                  setBackup(!backup);
                  setCode("");
                  setError("");
                }}
              >
                {backup ? "Use authenticator app" : "Use a backup code"}
              </Button>
              <Button type="button" variant="link" disabled={pending} onClick={cancelSignIn}>
                Back to sign-in
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </PublicLayout>
  );
}
