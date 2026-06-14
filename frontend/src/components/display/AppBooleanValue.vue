<script setup lang="ts">
import { computed } from 'vue'
import AppIcon from '../icons/AppIcon.vue'
import AppNotSetValue from './AppNotSetValue.vue'

const props = defineProps<{
  value?: boolean | null
  label?: string
}>()

// Compute icon name and color based on boolean value
const iconName = computed(() => props.value ? 'success' : 'error')
const iconColor = computed(() => props.value ? 'text-brand-500' : 'text-copy-muted')
</script>

<template>
  <span v-if="typeof props.value === 'boolean'" class="inline-flex items-center"
    :aria-label="`${label ?? 'Value'}: ${props.value ? 'Yes' : 'No'}`" :title="props.value ? 'Yes' : 'No'">
    <AppIcon category="status" :name="iconName" :color="iconColor" />
    <span class="sr-only">{{ props.value ? 'Yes' : 'No' }}</span>
  </span>
  <AppNotSetValue v-else :label="label" />
</template>
