import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useId, useState } from "react";
import { toast } from "sonner";
import { z } from "zod";
import { LoadingState } from "@/components/feedback/loading-state";
import { PasskeySettings } from "@/components/forms/passkey-settings";
import { PasswordInput } from "@/components/forms/password-input";
import { TwoFactorSettings } from "@/components/forms/two-factor-settings";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldLabel } from "@/components/ui/field";
import { accountApi, authApi } from "@/lib/api";
import { validPassword } from "@/lib/password-policy";

export const Route = createFileRoute("/app/settings")({
  validateSearch: z.object({ oauth_error: z.string().optional().catch(undefined) }),
  component: AccountSettingsPage,
});

function AccountSettingsPage() {
  const client = useQueryClient();
  const currentPasswordId = useId();
  const newPasswordId = useId();
  const account = useQuery({ queryKey: ["account"], queryFn: accountApi.settings });
  const passkeys = useQuery({ queryKey: ["passkeys"], queryFn: accountApi.passkeys });
  const settings = account.data;
  const availableConnections =
    settings?.providers.filter((provider) => provider.connected && provider.available).length ?? 0;
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const oauthError = Route.useSearch().oauth_error;
  const link = useMutation({
    mutationFn: authApi.linkOAuth,
    onSuccess: ({ authorizeUrl }) => window.location.assign(authorizeUrl),
    onError: (error) => toast.error(error.message),
  });
  const unlink = useMutation({
    mutationFn: authApi.unlinkOAuth,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ["account"] });
      toast.success("Provider disconnected");
    },
    onError: (error) => toast.error(error.message),
  });
  const password = useMutation({
    mutationFn: () => accountApi.setPassword(currentPassword, newPassword),
    onSuccess: async () => {
      setCurrentPassword("");
      setNewPassword("");
      setPasswordError("");
      await client.invalidateQueries({ queryKey: ["account"] });
      toast.success("Password saved. Other sessions signed out.");
    },
    onError: (error) => setPasswordError(error.message),
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Account settings" description="Manage sign-in methods and password." />
      {account.isLoading ? (
        <LoadingState label="Loading account…" />
      ) : account.error ? (
        <Alert variant="destructive">
          <AlertDescription>{account.error.message}</AlertDescription>
        </Alert>
      ) : settings ? (
        <>
          {oauthError && (
            <Alert variant="destructive">
              <AlertDescription>OAuth connection failed. Try again.</AlertDescription>
            </Alert>
          )}
          {settings.providers.map((provider) => {
            const isLastSignInMethod =
              !settings.hasPassword &&
              availableConnections <= (provider.available ? 1 : 0) &&
              (passkeys.data?.length ?? 0) === 0;

            return (
              <Card key={provider.id}>
                <CardHeader>
                  <CardTitle>{provider.name}</CardTitle>
                  <CardDescription>
                    {provider.connected
                      ? `${provider.name} account connected.`
                      : `Connect ${provider.name} for sign-in.`}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {provider.connected ? (
                    <Button
                      variant="outline"
                      disabled={unlink.isPending || link.isPending || isLastSignInMethod}
                      onClick={() => unlink.mutate(provider.id)}
                    >
                      Disconnect {provider.name}
                    </Button>
                  ) : provider.available ? (
                    <Button
                      variant="outline"
                      disabled={link.isPending || unlink.isPending}
                      onClick={() => link.mutate(provider.id)}
                    >
                      Connect {provider.name}
                    </Button>
                  ) : (
                    <p className="text-sm text-muted-foreground">Provider is not configured.</p>
                  )}
                  {provider.connected && isLastSignInMethod && (
                    <p className="mt-3 text-sm text-muted-foreground">
                      Add a password, passkey, or another provider before disconnecting.
                    </p>
                  )}
                </CardContent>
              </Card>
            );
          })}
          <PasskeySettings
            account={settings}
            passkeys={passkeys.data ?? []}
            loading={passkeys.isLoading}
            error={passkeys.error?.message}
          />
          <TwoFactorSettings account={settings} />
          <Card>
            <CardHeader>
              <CardTitle>{settings.hasPassword ? "Update password" : "Add password"}</CardTitle>
              <CardDescription>
                {settings.hasPassword
                  ? "Enter current password before choosing a new one."
                  : "Use password as another way to sign in."}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form
                className="grid max-w-lg gap-4"
                onSubmit={(event) => {
                  event.preventDefault();
                  setPasswordError("");

                  if (!password.isPending && validPassword(newPassword)) password.mutate();
                }}
              >
                {settings.hasPassword && (
                  <Field>
                    <FieldLabel htmlFor={currentPasswordId} className="text-sm font-medium">
                      Current password
                    </FieldLabel>
                    <PasswordInput
                      id={currentPasswordId}
                      autoComplete="current-password"
                      value={currentPassword}
                      onChange={(event) => setCurrentPassword(event.target.value)}
                      required
                      disabled={password.isPending}
                    />
                  </Field>
                )}
                <Field>
                  <FieldLabel htmlFor={newPasswordId} className="text-sm font-medium">
                    New password
                  </FieldLabel>
                  <PasswordInput
                    id={newPasswordId}
                    autoComplete="new-password"
                    showRequirements
                    value={newPassword}
                    onChange={(event) => setNewPassword(event.target.value)}
                    required
                    disabled={password.isPending}
                  />
                </Field>
                {passwordError && (
                  <Alert variant="destructive">
                    <AlertDescription>{passwordError}</AlertDescription>
                  </Alert>
                )}
                <Button
                  className="w-fit"
                  disabled={
                    password.isPending ||
                    !validPassword(newPassword) ||
                    (settings.hasPassword && !currentPassword)
                  }
                  type="submit"
                >
                  {password.isPending
                    ? "Saving…"
                    : settings.hasPassword
                      ? "Update password"
                      : "Set password"}
                </Button>
              </form>
            </CardContent>
          </Card>
        </>
      ) : null}
    </div>
  );
}
