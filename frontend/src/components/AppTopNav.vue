<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import AppCreateButton from './AppCreateButton.vue'

type NavItem = {
  to: string
  label: string
  icon: string
}

const props = defineProps<{
  navItems: NavItem[]
  currentPath: string
}>()

const emit = defineEmits<{
  logout: []
}>()

const onLogout = () => {
  emit('logout')
}

const router = useRouter()
const route = useRoute()

const createTargetPath = computed(() => {
  if (props.currentPath.startsWith('/persons')) {
    return '/persons'
  }

  if (props.currentPath.startsWith('/gear')) {
    return '/gear'
  }

  if (props.currentPath.startsWith('/sets')) {
    return '/sets'
  }

  return null
})

const canCreateFromCurrentPage = computed(() => createTargetPath.value !== null)

const onCreatePerson = async () => {
  if (!createTargetPath.value) {
    return
  }

  await router.push({
    path: createTargetPath.value,
    query: {
      ...route.query,
      create: '1',
    },
  })
}

const settingsItem = props.navItems.find((item) => item.to === '/settings')
const primaryNavItems = props.navItems.filter((item) => item.to !== '/settings')
</script>

<template>
  <header data-component="app-top-nav"
    class="border-line-subtle bg-surface-elevated fixed inset-x-0 top-0 z-40 border-b backdrop-blur">
    <div class="flex w-full flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-6 lg:px-10">
      <div class="flex items-center gap-3">
        <p class="text-brand-500 text-xs font-semibold uppercase tracking-[0.18em]">Packing List</p>
      </div>

      <nav data-element="top-nav-primary" class="flex flex-wrap items-center gap-1">
        <RouterLink v-for="item in primaryNavItems" :key="item.to" :to="item.to"
          :data-element="`nav-link-${item.to.replace('/', '') || 'home'}`"
          :aria-current="currentPath.startsWith(item.to) ? 'page' : undefined"
          class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition" :class="currentPath.startsWith(item.to)
            ? 'bg-surface-inverse text-ink-inverse'
            : 'text-copy hover:bg-surface-soft hover:text-ink'">
          <i :class="item.icon" aria-hidden="true"></i>
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="flex items-center gap-2">
        <AppCreateButton data-element="nav-create-person" :visible="canCreateFromCurrentPage" @click="onCreatePerson" />
        <RouterLink v-if="settingsItem" :to="settingsItem.to" data-element="nav-link-settings"
          class="border-line text-copy hover:bg-surface-soft hover:text-ink flex items-center gap-2 rounded-lg border px-3 py-1.5 text-sm font-medium transition"
          :class="currentPath.startsWith(settingsItem.to) ? 'bg-surface-soft' : ''">
          <i :class="settingsItem.icon" aria-hidden="true"></i>
          <span>{{ settingsItem.label }}</span>
        </RouterLink>
        <Button data-element="nav-logout" label="Logout" icon="pi pi-sign-out" size="small" outlined
          @click="onLogout" />
      </div>
    </div>
  </header>
</template>
