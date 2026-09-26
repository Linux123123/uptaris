import { createFileRoute } from "@tanstack/react-router";
import { StatusPage } from "@/components/status-page";
import { PublicLayout } from "@/components/layout/public-layout";

export const Route = createFileRoute("/")({
  component: PublicStatusPage,
});

function PublicStatusPage() {
  return (
    <PublicLayout>
      <StatusPage />
    </PublicLayout>
  );
}
