import { Circle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import type { IncidentSeverity, IncidentStatus, ServerStatus } from "@/lib/api";

const colors: Record<string, string> = {
  up: "border-emerald-500/25 bg-emerald-500/10 text-emerald-400",
  resolved: "border-emerald-500/25 bg-emerald-500/10 text-emerald-400",
  down: "border-red-500/25 bg-red-500/10 text-red-400",
  critical: "border-red-500/25 bg-red-500/10 text-red-400",
  high: "border-orange-500/25 bg-orange-500/10 text-orange-400",
  paused: "border-amber-500/25 bg-amber-500/10 text-amber-300",
  acknowledged: "border-amber-500/25 bg-amber-500/10 text-amber-300",
  medium: "border-amber-500/25 bg-amber-500/10 text-amber-300",
  open: "border-blue-500/25 bg-blue-500/10 text-blue-300",
  low: "border-slate-500/25 bg-slate-500/10 text-slate-300",
};

export function StatusBadge({
  value,
}: {
  value: ServerStatus | IncidentStatus | IncidentSeverity;
}) {
  return (
    <Badge variant="outline" className={colors[value]}>
      <Circle className="size-2 fill-current" />
      {value}
    </Badge>
  );
}
