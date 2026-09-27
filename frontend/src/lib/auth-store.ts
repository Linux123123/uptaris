import { createStore } from "@tanstack/react-store";
import type { User } from "@/lib/api";

export type AuthStatus = "booting" | "authenticated" | "anonymous";

export type AuthUser = Pick<User, "id" | "email" | "role">;

export type AuthState = { status: AuthStatus; user: AuthUser | null; expired: boolean };

export const authStore = createStore<AuthState>({ status: "booting", user: null, expired: false });

export function setAuthenticated(user: AuthUser) {
  authStore.setState(() => ({ status: "authenticated", user, expired: false }));
}

export function setAnonymous(expired = false) {
  authStore.setState(() => ({ status: "anonymous", user: null, expired }));
}
