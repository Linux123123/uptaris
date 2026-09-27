import { Button } from "@/components/ui/button";

export function FilterButton({
  label,
  active,
  onClick,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <Button
      type="button"
      aria-pressed={active}
      variant={active ? "default" : "outline"}
      size="sm"
      className="rounded-full capitalize"
      onClick={onClick}
    >
      {label}
    </Button>
  );
}
