import type {
  AuthenticationResponseJSON,
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
  RegistrationResponseJSON,
} from "@simplewebauthn/browser";

const baseURL = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";
let accessToken = "";
let refreshInFlight: Promise<boolean> | null = null;
let sessionExpiredHandler: (() => void) | undefined;

export type Role = "viewer" | "operator" | "admin";

export type ServerStatus = "up" | "down" | "paused";

export type MonitorType = "http" | "tcp" | "icmp";

export type IncidentSeverity = "low" | "medium" | "high" | "critical";

export type IncidentStatus = "open" | "acknowledged" | "resolved";

export type Links = { self: string; next: string | null; previous: string | null };

export type Pagination = { page: number; pageSize: number; total: number; totalPages: number };

export type ListResponse<T> = { data: T[]; pagination: Pagination; links: Links };

export type ResourceResponse<T> = { data: T; _links: Record<string, string> };

export type User = { id: string; email: string; role: Role; createdAt: string; updatedAt: string };

export type Server = {
  id: string;
  ownerId: string;
  name: string;
  address: string;
  operatingSystem: string;
  description: string;
  status: ServerStatus;
  createdAt: string;
  updatedAt: string;
};

export type Monitor = {
  id: string;
  serverId: string;
  name: string;
  type: MonitorType;
  target: string;
  intervalSeconds: number;
  expectedHealth: string;
  status: ServerStatus;
  lastCheckedAt: string | null;
  createdAt: string;
  updatedAt: string;
};

export type Incident = {
  id: string;
  monitorId: string;
  title: string;
  description: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  startedAt: string;
  resolvedAt: string | null;
  createdAt: string;
  updatedAt: string;
};

export type AuthResponse = { accessToken: string; user: Pick<User, "id" | "email" | "role"> };

export type LoginResponse = AuthResponse | { twoFactorRequired: true };

export type TwoFactorEnrollment = { secret: string; uri: string; expiresAt: string };

export type BackupCodesResponse = { backupCodes: string[] };

export type OAuthProvider = { id: string; name: string; loginUrl: string };

export type OAuthConnection = { id: string; name: string; available: boolean; connected: boolean };

export type Passkey = { id: string; name: string; createdAt: string; lastUsedAt: string | null };

export type PasskeyOptions<T> = { ceremonyToken: string; optionsJSON: T };

export type AccountSettings = {
  email: string;
  hasPassword: boolean;
  providers: OAuthConnection[];
  twoFactorEnabled: boolean;
  backupCodesRemaining: number;
};

export type StatusResponse = { servers: number; monitors: number; openIncidents: number };

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

function withQuery(path: string, params: Record<string, string | number | undefined> = {}) {
  const query = new URLSearchParams();

  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") query.set(key, String(value));
  }

  return query.size ? `${path}?${query}` : path;
}

async function refreshAccessToken() {
  if (!refreshInFlight) {
    refreshInFlight = fetch(`${baseURL}/auth/refresh`, {
      method: "POST",
      credentials: "include",
      headers: { Accept: "application/json" },
    })
      .then(async (response) => {
        if (!response.ok) return false;

        const result = (await response.json()) as AuthResponse;
        accessToken = result.accessToken;

        return true;
      })
      .catch(() => false)
      .finally(() => {
        refreshInFlight = null;
      });
  }

  return refreshInFlight;
}

export async function api<T>(path: string, init: RequestInit = {}, retried = false): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");

  if (init.body) headers.set("Content-Type", "application/json");

  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);

  let response: Response;

  try {
    response = await fetch(`${baseURL}${path}`, { ...init, credentials: "include", headers });
  } catch (cause) {
    throw new ApiError(
      0,
      "network_error",
      cause instanceof Error ? cause.message : "Network unavailable",
    );
  }

  // First-factor and challenge failures must not trigger normal session restoration.
  const canRefresh =
    !["/auth/login", "/auth/register", "/auth/refresh"].includes(path) &&
    !path.startsWith("/auth/two-factor/") &&
    !path.startsWith("/auth/passkeys/");

  if (response.status === 401 && canRefresh) {
    if (!retried && (await refreshAccessToken())) return api<T>(path, init, true);

    accessToken = "";
    sessionExpiredHandler?.();
  }

  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(
      response.status,
      body?.error?.code ?? "request_failed",
      body?.error?.message ?? `Request failed with status ${response.status}`,
    );
  }

  return (response.status === 204 ? undefined : await response.json()) as T;
}

export const authApi = {
  register: (email: string, password: string) =>
    api<Pick<User, "id" | "email" | "role">>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  login: (email: string, password: string) =>
    api<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  beginPasskeyLogin: () =>
    api<PasskeyOptions<PublicKeyCredentialRequestOptionsJSON>>("/auth/passkeys/login/options", {
      method: "POST",
    }),
  finishPasskeyLogin: (ceremonyToken: string, credential: AuthenticationResponseJSON) =>
    api<AuthResponse>("/auth/passkeys/login/verify", {
      method: "POST",
      body: JSON.stringify({ ceremonyToken, credential }),
    }),
  verifyTwoFactor: (code: string, backupCode: boolean) =>
    api<AuthResponse>("/auth/two-factor/verify", {
      method: "POST",
      body: JSON.stringify({ code, backupCode }),
    }),
  cancelTwoFactor: () => api<void>("/auth/two-factor/cancel", { method: "POST" }),
  logout: () => api<void>("/auth/logout", { method: "POST" }),
  me: () => api<Pick<User, "id" | "email" | "role">>("/auth/me"),
  providers: () => api<OAuthProvider[]>("/auth/providers"),
  oauthLoginURL: (provider: OAuthProvider) => `${baseURL}${provider.loginUrl}`,
  linkOAuth: (provider: string) =>
    api<{ authorizeUrl: string }>(`/auth/oauth/${encodeURIComponent(provider)}/link`, {
      method: "POST",
    }),
  unlinkOAuth: (provider: string) =>
    api<void>(`/auth/oauth/${encodeURIComponent(provider)}/link`, { method: "DELETE" }),
};

export const accountApi = {
  settings: () => api<AccountSettings>("/account"),
  passkeys: () => api<Passkey[]>("/account/passkeys"),
  beginPasskeyRegistration: () =>
    api<PasskeyOptions<PublicKeyCredentialCreationOptionsJSON>>("/account/passkeys/options", {
      method: "POST",
    }),
  finishPasskeyRegistration: (
    ceremonyToken: string,
    credential: RegistrationResponseJSON,
    name: string,
  ) =>
    api<Passkey>("/account/passkeys/verify", {
      method: "POST",
      body: JSON.stringify({ ceremonyToken, credential, name }),
    }),
  deletePasskey: (id: string) =>
    api<void>(`/account/passkeys/${encodeURIComponent(id)}`, { method: "DELETE" }),
  setupTwoFactor: (currentPassword: string) =>
    api<TwoFactorEnrollment>("/account/two-factor/setup", {
      method: "POST",
      body: JSON.stringify({ currentPassword }),
    }),
  confirmTwoFactor: (code: string) =>
    api<BackupCodesResponse>("/account/two-factor/confirm", {
      method: "POST",
      body: JSON.stringify({ code }),
    }),
  manageTwoFactor: (disable: boolean, currentPassword: string, code: string, backupCode: boolean) =>
    api<BackupCodesResponse | undefined>(
      `/account/two-factor/${disable ? "disable" : "backup-codes"}`,
      { method: "POST", body: JSON.stringify({ currentPassword, code, backupCode }) },
    ),
  setPassword: (currentPassword: string, newPassword: string) =>
    api<void>("/account/password", {
      method: "PUT",
      body: JSON.stringify({ currentPassword, newPassword }),
    }),
};

export const systemApi = {
  status: (signal?: AbortSignal) => api<StatusResponse>("/status", { signal }),
  dashboard: (signal?: AbortSignal) => api<StatusResponse>("/dashboard", { signal }),
};

export const serversApi = {
  list: (
    params: { page?: number; pageSize?: number; status?: ServerStatus } = {},
    signal?: AbortSignal,
  ) => api<ListResponse<Server>>(withQuery("/servers", params), { signal }),
  get: (id: string, signal?: AbortSignal) =>
    api<ResourceResponse<Server>>(`/servers/${id}`, { signal }),
  create: (
    body: Pick<Server, "name" | "address" | "operatingSystem"> &
      Partial<Pick<Server, "description" | "status">>,
  ) => api<Server>("/servers", { method: "POST", body: JSON.stringify(body) }),
  update: (
    id: string,
    body: Partial<Pick<Server, "name" | "address" | "operatingSystem" | "description" | "status">>,
  ) => api<Server>(`/servers/${id}`, { method: "PATCH", body: JSON.stringify(body) }),
  remove: (id: string) => api<void>(`/servers/${id}`, { method: "DELETE" }),
};

export const monitorsApi = {
  list: (
    serverId: string,
    params: { page?: number; pageSize?: number; type?: MonitorType; status?: ServerStatus } = {},
    signal?: AbortSignal,
  ) => api<ListResponse<Monitor>>(withQuery(`/servers/${serverId}/monitors`, params), { signal }),
  get: (serverId: string, id: string, signal?: AbortSignal) =>
    api<ResourceResponse<Monitor>>(`/servers/${serverId}/monitors/${id}`, { signal }),
  create: (
    serverId: string,
    body: Pick<Monitor, "name" | "type" | "target" | "intervalSeconds" | "expectedHealth"> &
      Partial<Pick<Monitor, "status">>,
  ) =>
    api<Monitor>(`/servers/${serverId}/monitors`, { method: "POST", body: JSON.stringify(body) }),
  update: (
    serverId: string,
    id: string,
    body: Partial<
      Pick<Monitor, "name" | "type" | "target" | "intervalSeconds" | "expectedHealth" | "status">
    >,
  ) =>
    api<Monitor>(`/servers/${serverId}/monitors/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  remove: (serverId: string, id: string) =>
    api<void>(`/servers/${serverId}/monitors/${id}`, { method: "DELETE" }),
};

export type IncidentRow = Incident & { serverId: string; monitorName: string };

export const incidentsApi = {
  overview: (
    params: {
      page: number;
      pageSize: number;
      status?: IncidentStatus;
      severity?: IncidentSeverity;
    },
    signal?: AbortSignal,
  ) => api<ListResponse<IncidentRow>>(withQuery("/incidents", params), { signal }),
  get: (serverId: string, monitorId: string, id: string, signal?: AbortSignal) =>
    api<ResourceResponse<Incident>>(`/servers/${serverId}/monitors/${monitorId}/incidents/${id}`, {
      signal,
    }),
  list: (
    serverId: string,
    monitorId: string,
    params: {
      page?: number;
      pageSize?: number;
      severity?: IncidentSeverity;
      status?: IncidentStatus;
    } = {},
    signal?: AbortSignal,
  ) =>
    api<ListResponse<Incident>>(
      withQuery(`/servers/${serverId}/monitors/${monitorId}/incidents`, params),
      { signal },
    ),
  create: (
    serverId: string,
    monitorId: string,
    body: Pick<Incident, "title" | "severity"> &
      Partial<Pick<Incident, "description" | "status" | "startedAt" | "resolvedAt">>,
  ) =>
    api<Incident>(`/servers/${serverId}/monitors/${monitorId}/incidents`, {
      method: "POST",
      body: JSON.stringify(body),
    }),
  update: (
    serverId: string,
    monitorId: string,
    id: string,
    body: Partial<
      Pick<Incident, "title" | "description" | "severity" | "status" | "startedAt" | "resolvedAt">
    >,
  ) =>
    api<Incident>(`/servers/${serverId}/monitors/${monitorId}/incidents/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  remove: (serverId: string, monitorId: string, id: string) =>
    api<void>(`/servers/${serverId}/monitors/${monitorId}/incidents/${id}`, { method: "DELETE" }),
};

export const adminApi = {
  users: (params: { page?: number; pageSize?: number } = {}, signal?: AbortSignal) =>
    api<ListResponse<User>>(withQuery("/admin/users", params), { signal }),
  removeUser: (id: string) => api<void>(`/admin/users/${id}`, { method: "DELETE" }),
  updateRole: (id: string, role: Role) =>
    api<User>(`/admin/users/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ role }),
    }),
};

export function setAccessToken(value: string) {
  accessToken = value;
}

export function onSessionExpired(handler: () => void) {
  sessionExpiredHandler = handler;
}

export async function restoreSession() {
  try {
    if (!(await refreshAccessToken())) return null;

    return await authApi.me();
  } catch {
    accessToken = "";

    return null;
  }
}
