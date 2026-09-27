import { keepPreviousData, queryOptions } from "@tanstack/react-query";
import {
  adminApi,
  incidentsApi,
  monitorsApi,
  serversApi,
  systemApi,
  type IncidentSeverity,
  type IncidentStatus,
  type MonitorType,
  type ServerStatus,
} from "@/lib/api";

export const statusQuery = () =>
  queryOptions({
    queryKey: ["status"] as const,
    queryFn: ({ signal }) => systemApi.status(signal),
    staleTime: 30_000,
  });

export const serversQuery = (filters: { page: number; pageSize: number; status?: ServerStatus }) =>
  queryOptions({
    queryKey: ["servers", filters] as const,
    queryFn: ({ signal }) => serversApi.list(filters, signal),
    placeholderData: keepPreviousData,
  });

export const serverQuery = (serverId: string) =>
  queryOptions({
    queryKey: ["server", serverId] as const,
    queryFn: ({ signal }) => serversApi.get(serverId, signal),
  });

export const monitorsQuery = (
  serverId: string,
  filters: { page: number; pageSize: number; type?: MonitorType; status?: ServerStatus },
) =>
  queryOptions({
    queryKey: ["monitors", serverId, filters] as const,
    queryFn: ({ signal }) => monitorsApi.list(serverId, filters, signal),
    placeholderData: keepPreviousData,
  });

export const monitorQuery = (serverId: string, monitorId: string) =>
  queryOptions({
    queryKey: ["monitor", serverId, monitorId] as const,
    queryFn: ({ signal }) => monitorsApi.get(serverId, monitorId, signal),
  });

export const incidentsQuery = (
  serverId: string,
  monitorId: string,
  filters: { page: number; pageSize: number; severity?: IncidentSeverity; status?: IncidentStatus },
) =>
  queryOptions({
    queryKey: ["incidents", serverId, monitorId, filters] as const,
    queryFn: ({ signal }) => incidentsApi.list(serverId, monitorId, filters, signal),
    placeholderData: keepPreviousData,
  });

export const usersQuery = (filters: { page: number; pageSize: number }) =>
  queryOptions({
    queryKey: ["users", filters] as const,
    queryFn: ({ signal }) => adminApi.users(filters, signal),
    placeholderData: keepPreviousData,
  });

export const dashboardQuery = () =>
  queryOptions({ queryKey: ["dashboard"], queryFn: ({ signal }) => systemApi.dashboard(signal) });

export const incidentOverviewQuery = (filters: {
  page: number;
  pageSize: number;
  status?: IncidentStatus;
  severity?: IncidentSeverity;
}) =>
  queryOptions({
    queryKey: ["incident-overview", filters],
    queryFn: ({ signal }) => incidentsApi.overview(filters, signal),
    placeholderData: keepPreviousData,
  });
