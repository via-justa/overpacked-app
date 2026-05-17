<script setup lang="ts">
import AppNotSetValue from './AppNotSetValue.vue'

const props = defineProps<{
  value?: boolean | null
  label?: string
}>()
</script>

<template>
  <span v-if="typeof props.value === 'boolean'" class="inline-flex items-center"
    :aria-label="`${label ?? 'Value'}: ${props.value ? 'Yes' : 'No'}`" :title="props.value ? 'Yes' : 'No'">
    <i :class="props.value ? 'pi pi-check-circle text-brand-500' : 'pi pi-times-circle text-copy-muted'"
      aria-hidden="true"></i>
    <span class="sr-only">{{ props.value ? 'Yes' : 'No' }}</span>
  </span>
  <AppNotSetValue v-else :label="label" />
</template>
