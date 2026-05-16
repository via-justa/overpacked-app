<script setup lang="ts">
import { RouterView, useRoute, useRouter } from 'vue-router'
import AppTopNav from '../components/AppTopNav.vue'
import { logoutAuth } from '../lib/api/auth'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const navItems = [
  { to: '/dashboard', label: 'Dashboard', icon: 'pi pi-home' },
  { to: '/packs', label: 'Packs', icon: 'pi pi-briefcase' },
  { to: '/sets', label: 'Sets', icon: 'pi pi-sitemap' },
  { to: '/gear', label: 'Gear', icon: 'pi pi-box' },
  { to: '/persons', label: 'Persons', icon: 'pi pi-users' },
  { to: '/settings', label: 'Settings', icon: 'pi pi-cog' },
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
