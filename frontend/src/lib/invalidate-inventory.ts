import type { QueryClient, QueryKey } from "@tanstack/react-query";

export async function invalidateInventory(queryClient: QueryClient, ...resourceKeys: QueryKey[]) {
  const queryKeys: QueryKey[] = [["dashboard"], ["status"], ["incident-overview"], ...resourceKeys];

  await Promise.all(queryKeys.map((queryKey) => queryClient.invalidateQueries({ queryKey })));
}
