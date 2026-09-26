import { QueryClient } from "@tanstack/react-query";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: (count, error) =>
        ![400, 401, 403, 404, 409, 422].includes((error as { status?: number }).status ?? 0) &&
        count < 2,
    },
  },
});
