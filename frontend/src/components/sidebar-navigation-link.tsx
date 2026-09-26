import type { ComponentProps } from "react";
import { Link } from "@tanstack/react-router";
import { SidebarMenuButton, useSidebar } from "@/components/ui/sidebar";

type SidebarNavigationLinkProps = Omit<ComponentProps<typeof SidebarMenuButton>, "render"> & {
  to: "/app/dashboard" | "/app/servers" | "/app/incidents" | "/app/admin/users";
};

export function SidebarNavigationLink({ to, ...props }: SidebarNavigationLinkProps) {
  const { setOpenMobile } = useSidebar();

  return (
    <SidebarMenuButton {...props} render={<Link to={to} onClick={() => setOpenMobile(false)} />} />
  );
}
