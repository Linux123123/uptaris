import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";

export function ErrorState({ message, retry }: { message: string; retry?: () => void }) {
  return (
    <div className="grid min-h-40 place-items-center p-6 text-center">
      <div>
        <AlertTriangle className="mx-auto size-8 text-destructive" />
        <h3 className="mt-3 font-medium">Unable to load data</h3>
        <p className="mt-1 text-sm text-muted-foreground">{message}</p>
        {retry && (
          <Button className="mt-4" variant="outline" onClick={retry}>
            Try again
          </Button>
        )}
      </div>
    </div>
  );
}
