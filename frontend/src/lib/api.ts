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
export type User = { id: number; email: string; role: Role; createdAt: string; updatedAt: string };
export type Server = {
  id: number;
  ownerId: number;
  name: string;
  address: string;
  operatingSystem: string;
  description: string;
  status: ServerStatus;
  createdAt: string;
  updatedAt: string;
};
export type Monitor = {
  id: number;
  serverId: number;
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
  id: number;
  monitorId: number;
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

  const canRefresh = !["/auth/login", "/auth/register", "/auth/refresh"].includes(path);
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
    api<AuthResponse>("/auth/login", { method: "POST", body: JSON.stringify({ email, password }) }),
  logout: () => api<void>("/auth/logout", { method: "POST" }),
  me: () => api<Pick<User, "id" | "email" | "role">>("/auth/me"),
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
  get: (id: number, signal?: AbortSignal) =>
    api<ResourceResponse<Server>>(`/servers/${id}`, { signal }),
  create: (
    body: Pick<Server, "name" | "address" | "operatingSystem"> &
      Partial<Pick<Server, "description" | "status">>,
  ) => api<Server>("/servers", { method: "POST", body: JSON.stringify(body) }),
  update: (
    id: number,
    body: Partial<Pick<Server, "name" | "address" | "operatingSystem" | "description" | "status">>,
  ) => api<Server>(`/servers/${id}`, { method: "PATCH", body: JSON.stringify(body) }),
  remove: (id: number) => api<void>(`/servers/${id}`, { method: "DELETE" }),
};

export const monitorsApi = {
  list: (
    serverId: number,
    params: { page?: number; pageSize?: number; type?: MonitorType; status?: ServerStatus } = {},
    signal?: AbortSignal,
  ) => api<ListResponse<Monitor>>(withQuery(`/servers/${serverId}/monitors`, params), { signal }),
  get: (serverId: number, id: number, signal?: AbortSignal) =>
    api<ResourceResponse<Monitor>>(`/servers/${serverId}/monitors/${id}`, { signal }),
  create: (
    serverId: number,
    body: Pick<Monitor, "name" | "type" | "target" | "intervalSeconds" | "expectedHealth"> &
      Partial<Pick<Monitor, "status">>,
  ) =>
    api<Monitor>(`/servers/${serverId}/monitors`, { method: "POST", body: JSON.stringify(body) }),
  update: (
    serverId: number,
    id: number,
    body: Partial<
      Pick<Monitor, "name" | "type" | "target" | "intervalSeconds" | "expectedHealth" | "status">
    >,
  ) =>
    api<Monitor>(`/servers/${serverId}/monitors/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  remove: (serverId: number, id: number) =>
    api<void>(`/servers/${serverId}/monitors/${id}`, { method: "DELETE" }),
};

export type IncidentRow = Incident & { serverId: number; monitorName: string };
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
  get: (serverId: number, monitorId: number, id: number, signal?: AbortSignal) =>
    api<ResourceResponse<Incident>>(`/servers/${serverId}/monitors/${monitorId}/incidents/${id}`, {
      signal,
    }),
  list: (
    serverId: number,
    monitorId: number,
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
    serverId: number,
    monitorId: number,
    body: Pick<Incident, "title" | "severity"> &
      Partial<Pick<Incident, "description" | "status" | "startedAt" | "resolvedAt">>,
  ) =>
    api<Incident>(`/servers/${serverId}/monitors/${monitorId}/incidents`, {
      method: "POST",
      body: JSON.stringify(body),
    }),
  update: (
    serverId: number,
    monitorId: number,
    id: number,
    body: Partial<
      Pick<Incident, "title" | "description" | "severity" | "status" | "startedAt" | "resolvedAt">
    >,
  ) =>
    api<Incident>(`/servers/${serverId}/monitors/${monitorId}/incidents/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body),
    }),
  remove: (serverId: number, monitorId: number, id: number) =>
    api<void>(`/servers/${serverId}/monitors/${monitorId}/incidents/${id}`, { method: "DELETE" }),
};

export const adminApi = {
  users: (params: { page?: number; pageSize?: number } = {}, signal?: AbortSignal) =>
    api<ListResponse<User>>(withQuery("/admin/users", params), { signal }),
  removeUser: (id: number) => api<void>(`/admin/users/${id}`, { method: "DELETE" }),
  updateRole: (id: number, role: Role) =>
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
