import { createApp } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'
import { makeTestQueryClient } from './queryClient'

export interface WithSetupResult<T> {
  result: T
  unmount: () => void
}

/**
 * Runs a composable inside a real (throwaway) component instance so that
 * lifecycle hooks, `provide`/`inject`, and `useQuery` have a valid setup
 * context. The Vue Query plugin is installed with a fresh test QueryClient.
 *
 * Use this for composables that call `useQuery`/`useMutation` or `provide`.
 * Pair with MSW to serve the underlying HTTP calls and `flushPromises()` to
 * let queries resolve before asserting.
 */
export function withSetup<T>(composable: () => T): WithSetupResult<T> {
  let result!: T
  const app = createApp({
    setup() {
      result = composable()
      return () => null
    },
  })
  app.use(VueQueryPlugin, { queryClient: makeTestQueryClient() })
  app.mount(document.createElement('div'))
  return { result, unmount: () => app.unmount() }
}
