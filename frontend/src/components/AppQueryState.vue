<script setup lang="ts">
import type { UseQueryReturnType } from '@tanstack/vue-query'
import AppQueryError from './AppQueryError.vue'
import AppLoadingState from './AppLoadingState.vue'
import AppEmptyState from './AppEmptyState.vue'

interface Props {
  query: UseQueryReturnType<unknown, Error>
  loadingMessage?: string
  emptyMessage?: string
  errorFallback?: string
}

const props = withDefaults(defineProps<Props>(), {
  loadingMessage: 'Loading...',
  emptyMessage: 'No data available',
  errorFallback: 'Unable to load data',
})

/**
 * Determines if the empty state should be shown.
 * Override this slot to customize empty state detection logic.
 */
const showEmpty = () => {
  if (props.query.isPending.value || props.query.isError.value) {
    return false
  }
  
  const data = props.query.data.value
  if (data === null || data === undefined) {
    return true
  }
  
  if (Array.isArray(data) && data.length === 0) {
    return true
  }
  
  return false
}
</script>

<template>
  <div>
    <AppQueryError :query="query" :fallback-message="errorFallback" />

    <AppLoadingState v-if="query.isPending.value" :message="loadingMessage" />

    <AppEmptyState v-else-if="showEmpty()" :message="emptyMessage" />

    <slot v-else />
  </div>
</template>
