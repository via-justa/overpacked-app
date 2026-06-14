<script setup lang="ts">
import { computed } from 'vue'
import { VueQueryDevtools } from '@tanstack/vue-query-devtools'
import Toast from 'primevue/toast'
import AppIcon from './components/icons/AppIcon.vue'
import { useAuthStore } from './stores/auth'

const isDev = import.meta.env.DEV
const authStore = useAuthStore()
const isBootstrapping = computed(() => authStore.isBootstrapping)
</script>

<template>
  <div v-if="isBootstrapping" data-component="app-bootstrap-screen"
    class="app-shell-gradient-soft grid min-h-screen place-items-center px-6">
    <div class="surface-chip text-copy flex items-center gap-3 px-5 py-3 text-sm font-semibold">
      <AppIcon category="feedback" name="spinner" spin />
      <span>Restoring session...</span>
    </div>
  </div>
  <div v-else data-component="app-router-view">
    <RouterView />
  </div>
  <Toast data-component="app-toast" position="bottom-center" aria-live="polite" aria-atomic="true" />
  <VueQueryDevtools v-if="isDev" />
</template>
