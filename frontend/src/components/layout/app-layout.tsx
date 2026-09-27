import { toast } from "sonner";
import { Outlet, useNavigate, useRouterState } from "@tanstack/react-router";
import { useSelector } from "@tanstack/react-store";
import { LayoutDashboard, LogOut, Server, Settings, ShieldAlert, Users } from "lucide-react";
import { authApi, setAccessToken } from "@/lib/api";
import { authStore, setAnonymous } from "@/lib/auth-store";
import { queryClient } from "@/lib/query-client";
import { Button } from "@/components/ui/button";
import { SidebarNavigationLink } from "@/components/sidebar-navigation-link";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar";

const navigation = [
  { to: "/app/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { to: "/app/servers", label: "Servers", icon: Server },
  { to: "/app/incidents", label: "Incidents", icon: ShieldAlert },
] as const;

function pageName(pathname: string) {
  if (pathname.includes("/admin/users")) return "Users";

  if (pathname.endsWith("/settings")) return "Account settings";

  if (pathname.includes("/monitors/")) return "Monitor details";

  if (pathname.includes("/servers/")) return "Server details";

  if (pathname.endsWith("/servers")) return "Servers";

  if (pathname.endsWith("/incidents")) return "Incidents";

  return "Dashboard";
}

export function AppLayout() {
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const user = useSelector(authStore, (state) => state.user);
  const links =
    user?.role === "admin"
      ? [...navigation, { to: "/app/admin/users" as const, label: "Users", icon: Users }]
      : navigation;

  async function logout() {
    try {
      await authApi.logout();
    } catch {
      toast.error("Sign out failed. Retry to revoke your session.");

      return;
    }

    setAccessToken("");
    queryClient.clear();
    setAnonymous();
    await navigate({ to: "/" });
  }

  return (
    <div className="dark min-h-svh bg-background text-foreground">
      <SidebarProvider>
        <Sidebar variant="inset" collapsible="icon">
          <SidebarHeader>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarNavigationLink size="lg" to="/app/dashboard" aria-label="Uptaris dashboard">
                  <span className="relative size-8 shrink-0" aria-hidden="true">
                    <img
                      src="/assets/logo-mark-white.svg"
                      alt=""
                      className="absolute inset-0 size-full object-contain"
                    />
                    <img
                      src="/assets/logo-mark-signal.svg"
                      alt=""
                      className="absolute inset-0 size-full object-contain group-hover/menu-button:animate-[logo-signal_1.4s_linear_infinite] group-hover/menu-button:drop-shadow-[0_0_6px_#02e595] group-focus-visible/menu-button:animate-[logo-signal_1.4s_linear_infinite] group-focus-visible/menu-button:drop-shadow-[0_0_6px_#02e595] motion-reduce:animate-none"
                    />
                  </span>
                  <span className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-semibold">Uptaris</span>
                    <span className="truncate text-xs text-muted-foreground">
                      Operations console
                    </span>
                  </span>
                </SidebarNavigationLink>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarHeader>
          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupLabel>Monitor</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {links.map(({ to, label, icon: Icon }) => (
                    <SidebarMenuItem key={to}>
                      <SidebarNavigationLink
                        to={to}
                        isActive={pathname === to || pathname.startsWith(`${to}/`)}
                        tooltip={label}
                      >
                        <Icon />
                        <span>{label}</span>
                      </SidebarNavigationLink>
                    </SidebarMenuItem>
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarFooter>
            <SidebarMenu>
              <SidebarMenuItem className="mb-2">
                <SidebarNavigationLink
                  to="/app/settings"
                  isActive={pathname === "/app/settings"}
                  tooltip="Account settings"
                >
                  <Settings />
                  <span>Account settings</span>
                </SidebarNavigationLink>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <div className="flex min-w-0 items-center gap-2 px-2 py-1 group-data-[collapsible=icon]:hidden">
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-sm font-medium">{user?.email}</span>
                    <span className="block text-xs capitalize text-muted-foreground">
                      {user?.role}
                    </span>
                  </span>
                  <Button variant="ghost" size="icon-sm" onClick={logout} aria-label="Sign out">
                    <LogOut />
                  </Button>
                </div>
                <SidebarMenuButton
                  className="hidden group-data-[collapsible=icon]:flex"
                  onClick={logout}
                  tooltip="Sign out"
                >
                  <LogOut />
                  <span>Sign out</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarFooter>
          <SidebarRail />
        </Sidebar>
        <SidebarInset>
          <header className="sticky top-0 z-20 flex h-14 shrink-0 items-center gap-2 border-b bg-background/90 px-4 backdrop-blur supports-backdrop-filter:bg-background/75">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbPage>{pageName(pathname)}</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </header>
          <main className="flex flex-1 flex-col p-4 md:p-6 lg:p-8">
            <div className="mx-auto w-full max-w-7xl">
              <Outlet />
            </div>
          </main>
          <footer className="border-t px-6 py-4 text-sm text-muted-foreground">
            Uptaris Incident Management
          </footer>
        </SidebarInset>
      </SidebarProvider>
    </div>
  );
}
