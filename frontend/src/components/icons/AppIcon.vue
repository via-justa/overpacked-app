<script setup lang="ts">
import { computed } from 'vue'
import { getIconName, getIconSizeClass, type IconCategory, type IconSize } from '../../lib/icons'

/**
 * AppIcon - Semantic icon component with consistent styling and accessibility
 * 
 * Automatically handles:
 * - Icon class generation from registry
 * - Size standardization
 * - Accessibility (aria-hidden)
 * - Animation support (spin, pulse)
 * 
 * @example
 * <AppIcon category="action" name="delete" size="sm" />
 * <AppIcon category="navigation" name="dashboard" size="lg" color="text-brand-500" />
 * <AppIcon category="feedback" name="spinner" spin />
 */

const props = withDefaults(defineProps<{
  /** Icon category from registry */
  category: IconCategory
  /** Icon name within category */
  name: string
  /** Standardized size (xs, sm, md, lg, xl, 2xl) */
  size?: IconSize
  /** Custom Tailwind color class (e.g., 'text-brand-500') */
  color?: string
  /** Additional CSS classes */
  class?: string
  /** Enable spin animation */
  spin?: boolean
  /** Enable pulse animation */
  pulse?: boolean
  /** Manual aria-hidden override (default: true for decorative icons) */
  ariaHidden?: boolean | 'false' | 'true'
}>(), {
  size: 'md',
  color: '',
  class: '',
  spin: false,
  pulse: false,
  ariaHidden: true,
})

// Get icon name from registry
const iconName = computed(() => getIconName(props.category, props.name))

// Build complete icon class string
const iconClass = computed(() => {
  const classes = ['pi', iconName.value]

  // Add size
  if (props.size) {
    classes.push(getIconSizeClass(props.size))
  }

  // Add color
  if (props.color) {
    classes.push(props.color)
  }

  // Add animations
  if (props.spin) {
    classes.push('pi-spin')
  }
  if (props.pulse) {
    classes.push('animate-pulse')
  }

  // Add custom classes
  if (props.class) {
    classes.push(props.class)
  }

  return classes.filter(Boolean).join(' ')
})

// Normalize aria-hidden value
const ariaHiddenValue = computed(() => {
  if (typeof props.ariaHidden === 'boolean') {
    return props.ariaHidden
  }
  return props.ariaHidden === 'true'
})
</script>

<template>
  <i :class="iconClass" :aria-hidden="ariaHiddenValue"></i>
</template>
