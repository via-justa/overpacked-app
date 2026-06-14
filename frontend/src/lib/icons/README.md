# Icon Registry System

Centralized icon management for consistent, type-safe icon usage across the application.

## Overview

The icon registry provides:
- **Type-safe** icon references via TypeScript
- **Semantic naming** for better code clarity
- **Centralized management** for easy updates
- **Size standardization** for consistent visual hierarchy
- **Accessibility helpers** built-in

## Usage

### Basic Import

```typescript
import { iconRegistry, getIconClass, getIconSizeClass } from '@/lib/icons/registry'
```

### In Templates (Current Pattern)

```vue
<template>
  <!-- Using registry for consistency -->
  <i :class="`pi ${iconRegistry.action.create}`" aria-hidden="true"></i>
  
  <!-- With helper function -->
  <i :class="getIconClass('action', 'delete')" aria-hidden="true"></i>
</template>
```

### With PrimeVue Button

```vue
<template>
  <Button 
    :icon="`pi ${iconRegistry.action.submit}`" 
    label="Save"
  />
</template>
```

### Type-Safe References

```typescript
import type { IconCategory, ActionIcon } from '@/lib/icons/registry'

// Type-safe icon name
const iconName: ActionIcon = 'create' // ✅ Autocomplete works
const invalid: ActionIcon = 'invalid' // ❌ TypeScript error

// Type-safe category
const category: IconCategory = 'action' // ✅
```

## Icon Categories

### Action Icons
User interactions: create, delete, edit, cancel, confirm, upload, etc.

```typescript
iconRegistry.action.create    // pi-plus
iconRegistry.action.delete    // pi-trash
iconRegistry.action.edit      // pi-pencil
iconRegistry.action.confirm   // pi-check
```

### Navigation Icons
App sections and routing: dashboard, gear, sets, persons, settings

```typescript
iconRegistry.navigation.dashboard  // pi-home
iconRegistry.navigation.gear       // pi-box
iconRegistry.navigation.sets       // pi-sitemap
iconRegistry.navigation.persons    // pi-users
```

### Status Icons
States and feedback: success, error, notSet, active, inactive

```typescript
iconRegistry.status.success   // pi-check-circle
iconRegistry.status.error     // pi-times-circle
iconRegistry.status.notSet    // pi-minus-circle
```

### Content Icons
Media and content: image, tag, externalLink

```typescript
iconRegistry.content.image         // pi-image
iconRegistry.content.tag           // pi-tag
iconRegistry.content.externalLink  // pi-external-link
```

### Directional Icons
Navigation indicators: chevrons, arrows

```typescript
iconRegistry.directional.chevronRight  // pi-chevron-right
iconRegistry.directional.arrowUp       // pi-arrow-up
```

### Feedback Icons
Loading and info states

```typescript
iconRegistry.feedback.spinner  // pi-spinner
iconRegistry.feedback.loading  // pi-spinner
```

## Size Standards

Use consistent size classes for visual hierarchy:

```typescript
import { getIconSizeClass } from '@/lib/icons/registry'

getIconSizeClass('xs')   // text-xs  (12px)
getIconSizeClass('sm')   // text-sm  (14px)
getIconSizeClass('md')   // text-base (16px) - default
getIconSizeClass('lg')   // text-lg  (18px)
getIconSizeClass('xl')   // text-xl  (20px)
getIconSizeClass('2xl')  // text-2xl (24px)
```

## Accessibility

Always include `aria-hidden="true"` for decorative icons:

```vue
<!-- ✅ Correct: Icon with adjacent label -->
<button>
  <i :class="getIconClass('action', 'delete')" aria-hidden="true"></i>
  <span>Delete</span>
</button>

<!-- ✅ Correct: Icon with screen reader text -->
<button aria-label="Delete item">
  <i :class="getIconClass('action', 'delete')" aria-hidden="true"></i>
  <span class="sr-only">Delete item</span>
</button>

<!-- ❌ Wrong: Icon without accessibility consideration -->
<i class="pi pi-trash"></i>
```

## Migration Guide

### Before (Direct PrimeIcons)
```vue
<template>
  <Button icon="pi pi-trash" label="Delete" />
  <i class="pi pi-check-circle text-brand-500"></i>
</template>
```

### After (Icon Registry)
```vue
<script setup lang="ts">
import { iconRegistry } from '@/lib/icons/registry'
</script>

<template>
  <Button :icon="`pi ${iconRegistry.action.delete}`" label="Delete" />
  <i :class="`pi ${iconRegistry.status.success} text-brand-500`" aria-hidden="true"></i>
</template>
```

## Adding New Icons

1. **Determine category** (action, navigation, status, content, directional, feedback)
2. **Add to appropriate category object** in `registry.ts`
3. **Use semantic name** (e.g., `confirmDelete` not `checkIcon`)
4. **Update TypeScript types** (automatic via `as const`)

Example:
```typescript
// In registry.ts
export const actionIcons = {
  // ... existing icons
  archive: 'pi-archive',  // Add new icon
} as const
```

## Best Practices

1. **Always use registry** instead of hardcoding `pi pi-*` classes
2. **Use semantic names** from registry, not raw PrimeIcons names
3. **Include aria-hidden** for decorative icons
4. **Use size constants** from `iconSizes` for consistency
5. **Document new icons** with clear semantic meaning

## Future Enhancements

Phase 2 will introduce:
- `AppIcon.vue` component for automatic accessibility and styling
- Color variants mapped to semantic tokens
- Animation helpers (spin, pulse, etc.)
- Icon composition utilities
