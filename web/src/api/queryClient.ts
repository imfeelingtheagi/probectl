import { QueryClient } from '@tanstack/react-query'

/**
 * Server state lives in TanStack Query; component/UI state stays in React. This
 * is the data-fetching pattern every later screen follows: a typed apiFetch
 * wrapped in useQuery/useMutation. A factory keeps each app mount (and each test)
 * on an isolated cache.
 */
export function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  })
}
