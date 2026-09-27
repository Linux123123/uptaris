import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
  EmptyDescription,
  EmptyContent,
} from "@/components/ui/empty";

export function ErrorState({ message, retry }: { message: string; retry?: () => void }) {
  return (
    <Empty role="alert" className="min-h-40">
      <EmptyHeader>
        <EmptyMedia variant="icon" className="text-destructive">
          <AlertTriangle aria-hidden="true" />
        </EmptyMedia>
        <EmptyTitle role="heading" aria-level={3}>
          Unable to load data
        </EmptyTitle>
        <EmptyDescription>{message}</EmptyDescription>
      </EmptyHeader>
      {retry && (
        <EmptyContent>
          <Button variant="outline" onClick={retry}>
            Try again
          </Button>
        </EmptyContent>
      )}
    </Empty>
  );
}
