<script setup lang="ts">
import { RouterView, useRoute, useRouter } from 'vue-router'
import AppTopNav from '../components/layout/AppTopNav.vue'
import AppActionRail from '../components/layout/AppActionRail.vue'
import { useActionRail } from '../composables/useActionRail'
import { logoutAuth } from '../lib/api/auth'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const { pinned } = useActionRail()

const navItems = [
  { to: '/trips', label: 'Trips', iconCategory: 'navigation' as const, iconName: 'trips' as const },
  { to: '/planner', label: 'Planner', iconCategory: 'navigation' as const, iconName: 'planner' as const },
  { to: '/gear', label: 'Gear', iconCategory: 'navigation' as const, iconName: 'gear' as const },
  { to: '/settings', label: 'Settings', iconCategory: 'navigation' as const, iconName: 'settings' as const },
]

const onLogout = async () => {
  const token = authStore.accessToken

  if (token) {
    try {
      await logoutAuth(token)
    } catch {
      // Clear local session regardless of backend logout outcome.
    }
  }

  authStore.clearSession()
  await router.replace('/login')
}
</script>

<template>
  <div data-component="app-layout-view"
    class="app-shell-gradient text-ink min-h-screen overflow-x-hidden transition-[padding] duration-200 motion-reduce:transition-none"
    :class="pinned ? 'md:pl-60' : 'md:pl-16'">
    <AppTopNav :nav-items="navItems" :current-path="route.path" @logout="onLogout" />
    <AppActionRail @logout="onLogout" />

    <main data-element="app-layout-content" class="w-full px-4 pb-6 pt-32 sm:px-6 sm:pb-8 md:pt-28 lg:px-10">
      <RouterView />
    </main>
  </div>
</template>
