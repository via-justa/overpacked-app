import { QueryClient } from '@tanstack/vue-query'

// Every test client is tracked so it can be drained in afterEach (see setup.ts).
const activeClients = new Set<QueryClient>()

// A QueryClient for tests: never retry (failures surface immediately instead of
// hanging the test), never cache between tests (gcTime 0), and never refetch on
// focus/reconnect (avoids spurious fetches during teardown). Do NOT reuse the
// app's client — its retry/staleTime defaults make tests slow and flaky.
export const makeTestQueryClient = (): QueryClient => {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, refetchOnWindowFocus: false, refetchOnReconnect: false },
      mutations: { retry: false },
    },
  })
  activeClients.add(client)
  return client
}

// Cancels in-flight queries and clears every client created during the test, so
// no mocked request is left in flight when the happy-dom window tears down.
// Left-over requests would otherwise complete after MSW closes and surface as
// benign-but-noisy AbortError / ECONNREFUSED unhandled rejections.
export const drainTestQueryClients = async (): Promise<void> => {
  for (const client of activeClients) {
    await client.cancelQueries()
    client.clear()
  }
  activeClients.clear()
}
