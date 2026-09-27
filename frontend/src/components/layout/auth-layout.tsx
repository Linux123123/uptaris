import type { ReactNode } from "react";
import { PublicLayout } from "@/components/layout/public-layout";

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <PublicLayout>
      <div className="mx-auto grid w-full max-w-5xl items-center gap-12 lg:min-h-full lg:grid-cols-[minmax(0,1fr)_minmax(0,1.15fr)] xl:gap-20">
        <div className="min-w-0 max-w-lg">
          <h1 className="text-4xl font-semibold tracking-tight">
            Keep operations calm when systems are not.
          </h1>
          <p className="mt-4 text-lg leading-8 text-muted-foreground">
            One focused console for server health, checks, incidents, and access control.
          </p>
          <img
            src="/assets/access-flow.svg"
            alt=""
            width={460}
            height={136}
            className="mt-10 hidden h-auto w-full max-w-sm lg:block"
          />
        </div>
        {children}
      </div>
    </PublicLayout>
  );
}
