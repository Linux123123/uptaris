import type { ComponentProps } from "react";
import { Link } from "@tanstack/react-router";
import { SidebarMenuButton, useSidebar } from "@/components/ui/sidebar";

type SidebarNavigationLinkProps = Omit<ComponentProps<typeof SidebarMenuButton>, "render"> & {
  to: ComponentProps<typeof Link>["to"];
};

export function SidebarNavigationLink({ to, ...props }: SidebarNavigationLinkProps) {
  const { setOpenMobile } = useSidebar();

  return (
    <SidebarMenuButton {...props} render={<Link to={to} onClick={() => setOpenMobile(false)} />} />
  );
}
