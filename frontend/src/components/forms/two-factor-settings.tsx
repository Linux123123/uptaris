import { useQueryClient } from "@tanstack/react-query";
import { Download, ShieldCheck } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { useId, useState } from "react";
import { toast } from "sonner";
import { PasswordInput } from "@/components/forms/password-input";
import { SecondFactorInput } from "@/components/forms/second-factor-input";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from "@/components/ui/collapsible";
import { Field, FieldLabel } from "@/components/ui/field";
import { accountApi, type AccountSettings, type TwoFactorEnrollment } from "@/lib/api";
import { validSecondFactor } from "@/lib/two-factor";

export function TwoFactorSettings({ account }: { account: AccountSettings }) {
  const client = useQueryClient();
  const id = useId();
  // Enrollment secrets and newly issued backup codes stay in component memory only.
  const [enrollment, setEnrollment] = useState<TwoFactorEnrollment | null>(null);
  const [codes, setCodes] = useState<string[]>([]);
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [backup, setBackup] = useState(false);
  const [action, setAction] = useState<"disable" | "regenerate" | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  const validCode = validSecondFactor(code, backup);

  async function submit() {
    if (pending) return;

    setError("");
    setPending(true);

    try {
      if (action) {
        const result = await accountApi.manageTwoFactor(
          action === "disable",
          password,
          code.trim(),
          backup,
        );
        setCodes(result?.backupCodes ?? []);
        setAction(null);
        toast.success(
          action === "disable"
            ? "Two-factor authentication disabled. Other sessions signed out."
            : "Backup codes refreshed. Previous codes no longer work.",
        );
      } else if (enrollment) {
        const result = await accountApi.confirmTwoFactor(code.trim());
        setCodes(result.backupCodes);
        setEnrollment(null);
        toast.success("Two-factor authentication enabled. Other sessions signed out.");
      } else {
        setEnrollment(await accountApi.setupTwoFactor(password));
      }

      setPassword("");
      setCode("");
      setBackup(false);
      await client.invalidateQueries({ queryKey: ["account"] });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Request failed. Try again.");
    } finally {
      setPending(false);
    }
  }

  function download() {
    const blob = new Blob(
      [
        "Uptaris two-factor backup codes\n" +
          account.email +
          "\n\n" +
          codes.join("\n") +
          "\n\nEach code works once. Store privately, separate from your password.\n",
      ],
      { type: "text/plain" },
    );
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "uptaris-backup-codes.txt";
    link.click();
    URL.revokeObjectURL(url);
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <ShieldCheck className="size-5" aria-hidden="true" />
          Two-factor authentication
        </CardTitle>
        <CardDescription>
          {account.twoFactorEnabled
            ? `Enabled for password and provider sign-in. ${account.backupCodesRemaining} backup codes remaining.`
            : "Add an authenticator code to protect password and provider sign-in."}
        </CardDescription>
      </CardHeader>
      <CardContent className="grid max-w-lg gap-4">
        {codes.length > 0 ? (
          <div className="grid gap-4">
            <Alert>
              <AlertTitle>Save your backup codes</AlertTitle>
              <AlertDescription>
                <p className="mt-1 text-sm text-muted-foreground">
                  Shown once. Each code can sign you in once if your authenticator is unavailable.
                  Keep them private and separate from your password.
                </p>
                <ul className="mt-4 grid gap-2 font-mono text-sm sm:grid-cols-2">
                  {codes.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              </AlertDescription>
            </Alert>
            <div className="flex flex-wrap gap-2">
              <Button onClick={download}>
                <Download className="size-4" />
                Download codes
              </Button>
              <Button variant="outline" onClick={() => setCodes([])}>
                I saved my codes
              </Button>
            </div>
          </div>
        ) : account.twoFactorEnabled && !action ? (
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              onClick={() => {
                setAction("regenerate");
                setError("");
              }}
            >
              Refresh backup codes
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                setAction("disable");
                setError("");
              }}
            >
              Disable 2FA
            </Button>
          </div>
        ) : (
          <form
            className="grid gap-4"
            onSubmit={(event) => {
              event.preventDefault();
              void submit();
            }}
          >
            {action && (
              <p className="text-sm text-muted-foreground">
                {action === "disable"
                  ? "Disabling removes your authenticator and all backup codes."
                  : "Refreshing invalidates every old backup code."}{" "}
                Other sessions will be signed out.
              </p>
            )}
            {enrollment && (
              <div className="grid justify-items-start gap-3">
                <p className="text-sm">
                  Scan this QR code with your authenticator app, then enter its code. Enrollment
                  expires in ten minutes.
                </p>
                <div className="rounded-lg bg-white p-4">
                  <QRCodeSVG
                    value={enrollment.uri}
                    size={192}
                    className="size-36 max-w-full sm:size-48"
                    level="M"
                    title="Scan to add Uptaris to your authenticator app"
                  />
                </div>
                <Collapsible className="w-full text-sm">
                  <CollapsibleTrigger
                    render={
                      <Button
                        variant="ghost"
                        className="h-auto whitespace-normal justify-start px-0 text-left"
                      />
                    }
                  >
                    Can’t scan? Enter setup key manually
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <p className="mt-2 break-all rounded border p-3 font-mono select-all">
                      {enrollment.secret}
                    </p>
                    <p className="mt-2 text-muted-foreground">
                      Time based · 6 digits · 30 seconds · SHA-1
                    </p>
                  </CollapsibleContent>
                </Collapsible>
              </div>
            )}
            {account.hasPassword && !enrollment && (
              <Field>
                <FieldLabel htmlFor={`${id}-password`} className="text-sm font-medium">
                  Current password
                </FieldLabel>
                <PasswordInput
                  id={`${id}-password`}
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  autoComplete="current-password"
                  required
                  disabled={pending}
                />
              </Field>
            )}
            {(enrollment || action) && (
              <Field>
                <FieldLabel htmlFor={`${id}-code`} className="text-sm font-medium">
                  {backup ? "Backup code" : "Authenticator code"}
                </FieldLabel>
                <SecondFactorInput
                  id={`${id}-code`}
                  value={code}
                  onChange={setCode}
                  backup={backup}
                  disabled={pending}
                />

                {action && (
                  <Button
                    type="button"
                    variant="ghost"
                    className="w-fit"
                    disabled={pending}
                    onClick={() => {
                      setBackup(!backup);
                      setCode("");
                    }}
                  >
                    {backup ? "Use authenticator app" : "Use a backup code"}
                  </Button>
                )}
              </Field>
            )}
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            <div className="flex flex-wrap gap-2">
              <Button
                type="submit"
                variant={action === "disable" ? "destructive" : "default"}
                disabled={
                  pending ||
                  (account.hasPassword && !enrollment && !password) ||
                  (!!(enrollment || action) && !validCode)
                }
              >
                {pending
                  ? "Saving…"
                  : action === "disable"
                    ? "Disable 2FA"
                    : action === "regenerate"
                      ? "Refresh backup codes"
                      : enrollment
                        ? "Verify and enable"
                        : "Set up authenticator"}
              </Button>
              {(action || enrollment) && (
                <Button
                  type="button"
                  variant="outline"
                  disabled={pending}
                  onClick={() => {
                    setAction(null);
                    setEnrollment(null);
                    setCode("");
                    setPassword("");
                    setBackup(false);
                    setError("");
                  }}
                >
                  Cancel
                </Button>
              )}
            </div>
          </form>
        )}
      </CardContent>
    </Card>
  );
}
