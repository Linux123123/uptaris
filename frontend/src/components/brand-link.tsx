import { Link } from "@tanstack/react-router";

export function BrandLink() {
  return (
    <Link to="/" aria-label="Uptaris home" className="flex items-center">
      <img
        src="/assets/logo-full-dark.svg"
        alt="Uptaris"
        className="h-7 w-auto max-w-full sm:h-8"
      />
    </Link>
  );
}
