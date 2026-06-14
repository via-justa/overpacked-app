import { render, type RenderOptions, type RenderResult } from '@testing-library/vue'
import type { Component } from 'vue'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import Tooltip from 'primevue/tooltip'
import Aura from '@primeuix/themes/aura'
import { VueQueryPlugin } from '@tanstack/vue-query'
import { createTestingPinia, type TestingOptions } from '@pinia/testing'
import {
  createRouter,
  createMemoryHistory,
  type Router,
  type RouteRecordRaw,
} from 'vue-router'
import { makeTestQueryClient } from './queryClient'

export interface RenderWithProvidersOptions extends Omit<RenderOptions<Component>, 'global'> {
  /** Routes for the test router. Defaults to a minimal stub set. */
  routes?: RouteRecordRaw[]
  /** Initial path to navigate to before mount. Defaults to '/'. */
  initialRoute?: string
  /** Options forwarded to createTestingPinia (initialState, stubActions, ...). */
  pinia?: TestingOptions
}

const stubRoutes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: { template: '<div />' } },
  { path: '/login', name: 'login', component: { template: '<div />' } },
  { path: '/trips', name: 'trips', component: { template: '<div />' } },
]

/**
 * Renders a component with the same plugins the real app installs in main.ts:
 * PrimeVue (+ ToastService + tooltip directive), a fresh testing Pinia, a memory
 * router, and a fresh QueryClient. Returns Testing Library's result plus the
 * router so tests can assert navigation.
 */
export function renderWithProviders(
  component: Component,
  options: RenderWithProvidersOptions = {},
): RenderResult & { router: Router } {
  const { routes, initialRoute = '/', pinia, ...renderOptions } = options

  const router = createRouter({
    history: createMemoryHistory(),
    routes: routes ?? stubRoutes,
  })

  const utils = render(component, {
    ...renderOptions,
    global: {
      plugins: [
        [PrimeVue, { theme: { preset: Aura } }],
        ToastService,
        createTestingPinia({ stubActions: false, ...pinia }),
        router,
        [VueQueryPlugin, { queryClient: makeTestQueryClient() }],
      ],
      directives: { tooltip: Tooltip },
    },
  })

  // Navigate after the router is installed (render mounts the app). Callers
  // `await router.isReady()` (or a waitFor) before asserting on the route.
  void router.push(initialRoute)

  return { ...utils, router }
}
