import { useRouter } from 'vue-router'
import type { ActionTarget } from './actionOptions'

// Maps every navigable action target to its destination route. Logout is
// handled by the caller via the onLogout callback.
const actionRoutes: Record<Exclude<ActionTarget, 'logout'>, { path: string; query?: Record<string, string> }> = {
  'add-trip': { path: '/trips/new' },
  'add-item': { path: '/gear', query: { action: 'create-item' } },
  'add-set': { path: '/sets', query: { create: '1' } },
  'add-person': { path: '/persons', query: { create: '1' } },
  'add-packing-list': { path: '/lists', query: { create: '1' } },
  'manage-manufacturers': { path: '/gear', query: { action: 'manufacturers' } },
  'manage-categories': { path: '/gear', query: { action: 'categories' } },
  'import-csv': { path: '/gear', query: { action: 'import' } },
  'settings': { path: '/settings' },
  'trips': { path: '/trips' },
  'planner': { path: '/planner' },
  'sets': { path: '/sets' },
  'lists': { path: '/lists' },
  'gear': { path: '/gear' },
  'persons': { path: '/persons' },
}

export function useActionRouter(onLogout: () => void) {
  const router = useRouter()

  const selectAction = async (target: ActionTarget) => {
    if (target === 'logout') {
      onLogout()
      return
    }

    await router.push(actionRoutes[target])
  }

  return { selectAction }
}
