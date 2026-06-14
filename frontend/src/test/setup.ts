import '@testing-library/jest-dom/vitest'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { server } from './msw/server'
import { drainTestQueryClients } from './queryClient'

// MSW lifecycle: fail loudly on any request without a matching handler so tests
// can't silently hit the network or drift from the API contract.
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(async () => {
  // Drain in-flight vue-query fetches (while MSW is still up) before resetting
  // handlers, so nothing leaks to the network when the window tears down.
  await drainTestQueryClients()
  server.resetHandlers()
})
// Intentionally NOT calling server.close() in afterAll: a request that is still
// settling when the happy-dom window tears down would otherwise reach the real
// network (ECONNREFUSED) once MSW stopped intercepting. Each test file runs in
// its own worker that is discarded after the run, so leaving the interceptor in
// place for the worker's lifetime is safe and keeps the output clean.
afterAll(() => server.resetHandlers())

// happy-dom gaps that app code/PrimeVue rely on:
//   - ResizeObserver: PrimeVue overlays observe their target.
//   - matchMedia: useIsMobile reads it (and window resize).
if (!('ResizeObserver' in globalThis)) {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof ResizeObserver
}

if (!('matchMedia' in globalThis)) {
  globalThis.matchMedia = ((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof globalThis.matchMedia
}
