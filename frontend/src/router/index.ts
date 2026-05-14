import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { pinia } from '../stores'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/',
      component: () => import('../views/AppLayoutView.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: { name: 'dashboard' },
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('../features/dashboard/views/DashboardPage.vue'),
        },
        {
          path: 'persons',
          name: 'persons',
          component: () => import('../features/persons/views/PersonsPage.vue'),
        },
        {
          path: 'packs',
          name: 'packs',
          component: () => import('../features/packs/views/PacksPage.vue'),
        },
        {
          path: 'items',
          redirect: { name: 'gear' },
        },
        {
          path: 'gear',
          name: 'gear',
          component: () => import('../features/items/views/ItemsPage.vue'),
        },
        {
          path: 'sets',
          name: 'sets',
          component: () => import('../features/sets/views/SetsPage.vue'),
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('../features/settings/views/SettingsPage.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore(pinia)
  await authStore.ensureBootstrapped()
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)
  const guestOnly = to.matched.some((record) => record.meta.guestOnly)

  if (requiresAuth && !authStore.isAuthenticated) {
    const notice = authStore.consumeAuthNotice()

    return {
      name: 'login',
      query: {
        redirect: to.fullPath,
        ...(notice ? { reason: notice } : {}),
      },
    }
  }

  if (guestOnly && authStore.isAuthenticated) {
    const redirectQuery = to.query.redirect
    if (typeof redirectQuery === 'string' && redirectQuery.startsWith('/')) {
      return redirectQuery
    }

    return { name: 'dashboard' }
  }

  return true
})

export default router