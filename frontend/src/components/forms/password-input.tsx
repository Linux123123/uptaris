import { useId, useState, type ComponentProps } from "react";
import { Check, Circle, Eye, EyeOff, ShieldCheck } from "lucide-react";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@/components/ui/input-group";
import { Progress } from "@/components/ui/progress";
import { Card, CardContent } from "@/components/ui/card";
import { passwordChecks } from "@/lib/password-policy";

type PasswordInputProps = Omit<ComponentProps<typeof InputGroupInput>, "type" | "value"> & {
  value: string;
  showRequirements?: boolean;
};

export function PasswordInput({ value, showRequirements = false, ...props }: PasswordInputProps) {
  const [visible, setVisible] = useState(false);
  const requirementsId = useId();
  const checks = passwordChecks(value);
  const passed = checks.filter((check) => check.met).length;
  const ready = passed === checks.length;
  const description =
    [props["aria-describedby"], showRequirements && requirementsId].filter(Boolean).join(" ") ||
    undefined;

  return (
    <div className="grid gap-3">
      <InputGroup>
        <InputGroupInput
          {...props}
          value={value}
          type={visible ? "text" : "password"}
          aria-describedby={description}
        />
        <InputGroupAddon align="inline-end">
          <InputGroupButton
            size="icon-xs"
            onClick={() => setVisible((current) => !current)}
            disabled={props.disabled}
            aria-label={visible ? "Hide password" : "Show password"}
            aria-pressed={visible}
          >
            {visible ? <EyeOff aria-hidden="true" /> : <Eye aria-hidden="true" />}
          </InputGroupButton>
        </InputGroupAddon>
      </InputGroup>
      {showRequirements && (
        <Card id={requirementsId} size="sm" className="bg-muted/30">
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between gap-3 text-xs">
              <span className="flex items-center gap-1.5 font-medium">
                <ShieldCheck className="size-3.5" aria-hidden="true" />
                Password requirements
              </span>
              <span
                aria-live="polite"
                className={
                  ready
                    ? "font-medium text-emerald-600 dark:text-emerald-400"
                    : "text-muted-foreground"
                }
              >
                {!value
                  ? "Choose a password"
                  : ready
                    ? "Ready to use"
                    : `${passed} of ${checks.length} met`}
              </span>
            </div>
            <Progress
              value={passed}
              max={checks.length}
              aria-label="Password requirements met"
              aria-valuetext={`${passed} of ${checks.length} requirements met`}
            />
            <ul className="grid gap-1.5 text-xs">
              {checks.map((check) => (
                <li
                  key={check.label}
                  className={`flex items-start gap-2 ${check.met ? "text-emerald-600 dark:text-emerald-400" : "text-muted-foreground"}`}
                >
                  {check.met ? (
                    <Check className="mt-0.5 size-3 shrink-0" aria-hidden="true" />
                  ) : (
                    <Circle className="mt-0.5 size-3 shrink-0" aria-hidden="true" />
                  )}
                  <span>
                    <span className="sr-only">{check.met ? "Met: " : "Needed: "}</span>
                    {check.label}
                  </span>
                </li>
              ))}
            </ul>
            <p className="text-xs text-muted-foreground">
              Use a unique password you do not use for other accounts.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
