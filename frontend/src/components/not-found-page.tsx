import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";

export function NotFoundPage() {
  return (
    <div className="dark grid min-h-svh place-items-center bg-background p-6 text-center text-foreground">
      <div>
        <p className="font-mono text-sm text-primary">404</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">Page not found</h1>
        <p className="mt-2 text-muted-foreground">Requested page does not exist.</p>
        <Button nativeButton={false} className="mt-6" render={<Link to="/" />}>
          Return to status
        </Button>
      </div>
    </div>
  );
}
