<script setup lang="ts">
import { computed } from 'vue'
import type { Label } from '../types'

const props = defineProps<{
  label: Label
  size?: 'sm' | 'md'
  removable?: boolean
}>()

const emit = defineEmits<{
  remove: []
}>()

const sizeClass = computed(() => {
  return props.size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm'
})

const getContrastColor = (color?: string | null): 'light' | 'dark' => {
  if (!color) {
    return 'light'
  }

  // Handle HSL colors
  if (color.startsWith('hsl')) {
    const match = color.match(/hsl\((\d+),\s*(\d+)%,\s*(\d+)%\)/)
    if (match) {
      const lightness = Number.parseInt(match[3], 10)
      return lightness > 55 ? 'dark' : 'light'
    }
  }

  // Handle hex colors
  const hex = color.replace('#', '')
  const r = Number.parseInt(hex.substring(0, 2), 16)
  const g = Number.parseInt(hex.substring(2, 4), 16)
  const b = Number.parseInt(hex.substring(4, 6), 16)

  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255

  return luminance > 0.5 ? 'dark' : 'light'
}

const backgroundColor = computed(() => {
  return props.label.color ?? '#6b7280'
})

const textColorClass = computed(() => {
  const contrast = getContrastColor(props.label.color)
  return contrast === 'light' ? 'text-ink-inverse' : 'text-ink'
})

const borderColor = computed(() => {
  const contrast = getContrastColor(props.label.color)
  return contrast === 'light'
    ? 'rgba(255, 255, 255, 0.2)'
    : 'rgba(0, 0, 0, 0.1)'
})
</script>

<template>
  <component :is="removable ? 'button' : 'span'" data-component="item-label"
    class="inline-flex items-center gap-1 rounded-full font-medium transition-all"
    :class="[sizeClass, textColorClass, removable ? 'hover:opacity-80 cursor-pointer' : '']"
    :type="removable ? 'button' : undefined" :aria-label="removable ? `Remove ${label.name} label` : undefined" :style="{
      backgroundColor: backgroundColor,
      border: `1px solid ${borderColor}`
    }" @click="removable ? emit('remove') : undefined">
    <span class="truncate">{{ label.name }}</span>
  </component>
</template>
