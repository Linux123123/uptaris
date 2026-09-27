import { browserSupportsWebAuthn, startRegistration } from "@simplewebauthn/browser";
import { useQueryClient } from "@tanstack/react-query";
import { useId, useState } from "react";
import { toast } from "sonner";
import { DeleteResourceDialog } from "@/components/dialogs/delete-resource-dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { accountApi, type AccountSettings, type Passkey } from "@/lib/api";

export function PasskeySettings({
  account,
  passkeys,
  loading,
  error,
}: {
  account: AccountSettings;
  passkeys: Passkey[];
  loading: boolean;
  error?: string;
}) {
  const client = useQueryClient();
  const nameId = useId();
  const [name, setName] = useState("");
  const [pending, setPending] = useState(false);
  const [actionError, setActionError] = useState("");
  const supported = browserSupportsWebAuthn();
  const hasOtherSignInMethod =
    account.hasPassword ||
    account.providers.some((provider) => provider.connected && provider.available);

  const isLastSignInMethod = passkeys.length === 1 && !hasOtherSignInMethod;

  async function addPasskey() {
    setPending(true);
    setActionError("");

    try {
      const { ceremonyToken, optionsJSON } = await accountApi.beginPasskeyRegistration();
      const credential = await startRegistration({ optionsJSON });
      await accountApi.finishPasskeyRegistration(ceremonyToken, credential, name.trim());
      setName("");
      await client.invalidateQueries({ queryKey: ["passkeys"] });
      toast.success("Passkey added");
    } catch (cause) {
      setActionError(
        cause instanceof Error && cause.name === "NotAllowedError"
          ? "Passkey request cancelled or timed out."
          : cause instanceof Error
            ? cause.message
            : "Could not add passkey.",
      );
    } finally {
      setPending(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Passkeys</CardTitle>
        <CardDescription>Sign in with device PIN, fingerprint, or face unlock.</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-5">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading passkeys…</p>
        ) : !error && passkeys.length === 0 ? (
          <p className="text-sm text-muted-foreground">No passkeys added yet.</p>
        ) : (
          <div className="divide-y rounded-lg border">
            {passkeys.map((passkey) => {
              return (
                <div key={passkey.id} className="flex items-center gap-4 px-4 py-3">
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">{passkey.name}</p>
                    <p className="text-xs text-muted-foreground">
                      Added {new Date(passkey.createdAt).toLocaleDateString()}
                      {passkey.lastUsedAt &&
                        ` · Used ${new Date(passkey.lastUsedAt).toLocaleDateString()}`}
                    </p>
                  </div>
                  <DeleteResourceDialog
                    trigger={
                      <Button variant="outline" size="sm" disabled={isLastSignInMethod || pending}>
                        Remove
                      </Button>
                    }
                    title="Remove passkey?"
                    description="This passkey will stop working for sign-in."
                    actionLabel="Remove passkey"
                    onConfirm={async () => {
                      await accountApi.deletePasskey(passkey.id);
                      await client.invalidateQueries({ queryKey: ["passkeys"] });
                      toast.success("Passkey removed");
                    }}
                  />
                </div>
              );
            })}
          </div>
        )}
        {isLastSignInMethod && (
          <p className="text-sm text-muted-foreground">
            Add another sign-in method before removing your last passkey.
          </p>
        )}
        {supported ? (
          <div className="grid max-w-sm gap-3">
            <Field>
              <FieldLabel htmlFor={nameId}>Passkey name (optional)</FieldLabel>
              <Input
                id={nameId}
                value={name}
                maxLength={80}
                placeholder="Work laptop"
                onChange={(event) => setName(event.target.value)}
                disabled={pending}
              />
            </Field>
            <Button className="w-fit" disabled={pending} onClick={() => void addPasskey()}>
              {pending ? "Waiting for passkey…" : "Add passkey"}
            </Button>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">Passkeys are unavailable in this browser.</p>
        )}
        {actionError && (
          <Alert variant="destructive">
            <AlertDescription>{actionError}</AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  );
}
