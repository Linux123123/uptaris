import { Spinner } from "@/components/ui/spinner";

export function LoadingState({ label = "Loading data" }: { label?: string }) {
  return (
    <div
      role="status"
      aria-live="polite"
      className="flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground"
    >
      <Spinner aria-hidden="true" />
      {label}
    </div>
  );
}
