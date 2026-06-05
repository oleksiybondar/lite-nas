import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { PropsWithChildren, ReactElement } from "react";
import { useState } from "react";

/**
 * Provides TanStack Query state for gateway-backed server data.
 *
 * Query ownership stays at the domain-hook level, while this provider supplies
 * the shared cache, request deduplication, and polling infrastructure used by
 * those hooks.
 */
export const QueryProvider = ({ children }: PropsWithChildren): ReactElement => {
  const [queryClient] = useState(createQueryClient);

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
};

/**
 * Creates the shared query client used by the admin panel.
 */
const createQueryClient = (): QueryClient => {
  return new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnWindowFocus: false,
        retry: false,
      },
    },
  });
};
