import type { ComponentProps } from "react";
import { REGEXP_ONLY_DIGITS } from "input-otp";
import { Input } from "@/components/ui/input";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp";

type SecondFactorInputProps = Pick<
  ComponentProps<typeof Input>,
  "id" | "disabled" | "autoFocus"
> & {
  value: string;
  onChange: (value: string) => void;
  backup?: boolean;
  large?: boolean;
};

export function SecondFactorInput({
  backup = false,
  large = false,
  value,
  onChange,
  ...props
}: SecondFactorInputProps) {
  if (backup) {
    return (
      <Input
        {...props}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        autoComplete="one-time-code"
        placeholder="xxxxx-xxxxx-xxxxx-xxxxx"
        maxLength={32}
        className={
          large
            ? "h-12 text-center font-mono text-base tracking-widest"
            : "font-mono tracking-widest"
        }
        required
      />
    );
  }

  return (
    <InputOTP
      {...props}
      value={value}
      onChange={onChange}
      maxLength={6}
      pattern={REGEXP_ONLY_DIGITS}
      autoComplete="one-time-code"
      inputMode="numeric"
      containerClassName={large ? "w-full justify-center" : undefined}
      required
    >
      <InputOTPGroup className={large ? "gap-1 sm:gap-2" : undefined}>
        {Array.from({ length: 6 }, (_, index) => (
          <InputOTPSlot
            key={index}
            index={index}
            className={
              large
                ? "size-9 rounded-lg border border-input bg-background text-lg font-semibold shadow-sm first:rounded-lg last:rounded-lg sm:size-12"
                : undefined
            }
          />
        ))}
      </InputOTPGroup>
    </InputOTP>
  );
}
