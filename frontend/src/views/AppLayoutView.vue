<script setup lang="ts">
import { RouterView, useRoute, useRouter } from 'vue-router'
import AppTopNav from '../components/layout/AppTopNav.vue'
import { logoutAuth } from '../lib/api/auth'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const navItems = [
  { to: '/dashboard', label: 'Dashboard', iconCategory: 'navigation' as const, iconName: 'dashboard' as const },
  { to: '/planner', label: 'Planner', iconCategory: 'navigation' as const, iconName: 'planner' as const },
  { to: '/lists', label: 'Lists', iconCategory: 'navigation' as const, iconName: 'lists' as const },
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
  <div data-component="app-layout-view" class="app-shell-gradient text-ink min-h-screen overflow-x-hidden">
    <AppTopNav :nav-items="navItems" :current-path="route.path" @logout="onLogout" />

    <main data-element="app-layout-content" class="w-full px-4 pb-6 pt-24 sm:px-6 sm:pb-8 sm:pt-28 lg:px-10">
      <RouterView />
    </main>
  </div>
</template>
