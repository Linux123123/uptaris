import { Inbox } from "lucide-react";

export function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <div className="grid min-h-48 place-items-center p-6 text-center">
      <div>
        <Inbox className="mx-auto size-8 text-muted-foreground" />
        <h3 className="mt-3 font-medium">{title}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}
