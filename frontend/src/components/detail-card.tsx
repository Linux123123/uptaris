import type { ReactNode } from "react";
import { Card, CardContent } from "@/components/ui/card";

export function DetailCard({ label, children }: { label: string; children: ReactNode }) {
  return (
    <Card size="sm">
      <CardContent>
        <p className="mb-2 text-xs text-muted-foreground">{label}</p>
        <div className="font-medium">{children}</div>
      </CardContent>
    </Card>
  );
}
