import { BrandLink } from "@/components/brand-link";
import { Link } from "@tanstack/react-router";
import { Activity, LogIn, Menu, UserPlus } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTitle, SheetTrigger } from "@/components/ui/sheet";

const links = [
  { to: "/", label: "System status", icon: Activity },
  { to: "/login", label: "Sign in", icon: LogIn },
  { to: "/register", label: "Create account", icon: UserPlus },
] as const;

export function PublicLayout({ children }: { children: ReactNode }) {
  return (
    <div className="dark grid min-h-svh grid-rows-[auto_1fr_auto] bg-background text-foreground">
      <header className="sticky top-0 z-30 border-b bg-background/90 backdrop-blur supports-[backdrop-filter]:bg-background/75">
        <div className="mx-auto flex h-16 w-full max-w-7xl items-center px-4 sm:px-6 lg:px-8">
          <BrandLink />
          <nav className="ml-auto hidden items-center gap-1 md:flex" aria-label="Main navigation">
            {links.map((link) => (
              <Button
                nativeButton={false}
                key={link.to}
                variant="ghost"
                size="sm"
                render={<Link to={link.to} />}
              >
                <link.icon />
                {link.label}
              </Button>
            ))}
          </nav>
          <Sheet>
            <SheetTrigger
              render={
                <Button
                  className="ml-auto md:hidden"
                  variant="ghost"
                  size="icon"
                  aria-label="Open menu"
                />
              }
            >
              <Menu />
            </SheetTrigger>
            <SheetContent side="right" className="w-72 p-6">
              <SheetTitle className="sr-only">Navigation</SheetTitle>
              <BrandLink />
              <nav className="mt-8 grid gap-2" aria-label="Mobile navigation">
                {links.map((link) => (
                  <Button
                    nativeButton={false}
                    key={link.to}
                    variant="ghost"
                    className="justify-start"
                    render={<Link to={link.to} />}
                  >
                    <link.icon />
                    {link.label}
                  </Button>
                ))}
              </nav>
            </SheetContent>
          </Sheet>
        </div>
      </header>
      <main className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 sm:py-14 lg:px-8 lg:py-16">
        {children}
      </main>
      <footer className="border-t py-6 text-center text-sm text-muted-foreground">
        Uptaris Infrastructure monitoring
      </footer>
    </div>
  );
}
