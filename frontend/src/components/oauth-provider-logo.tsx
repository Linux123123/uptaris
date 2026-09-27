import { FcGoogle } from "react-icons/fc";
import { SiGithub } from "react-icons/si";

export function OAuthProviderLogo({ provider }: { provider: string }) {
  if (provider === "github") {
    return <SiGithub aria-hidden="true" className="size-4" />;
  }

  if (provider === "google") {
    return <FcGoogle aria-hidden="true" className="size-4" />;
  }

  return null;
}
