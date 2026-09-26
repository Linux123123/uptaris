export const resourceStatusOptions = [
  { value: "up", label: "Up" },
  { value: "down", label: "Down" },
  { value: "paused", label: "Paused" },
] as const;

export const monitorTypeOptions = [
  { value: "http", label: "HTTP" },
  { value: "tcp", label: "TCP" },
  { value: "icmp", label: "ICMP" },
] as const;

export const incidentStatusOptions = [
  { value: "open", label: "Open" },
  { value: "acknowledged", label: "Acknowledged" },
  { value: "resolved", label: "Resolved" },
] as const;

export const incidentSeverityOptions = [
  { value: "low", label: "Low" },
  { value: "medium", label: "Medium" },
  { value: "high", label: "High" },
  { value: "critical", label: "Critical" },
] as const;
