<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import { iconRegistry, type ActionIcon } from '../../lib/icons'

/**
 * AppActionButton - a small, round, icon-only action button.
 *
 * Single source of truth for the app's action verbs (save/create/cancel/close/
 * delete/confirm/edit). Each preset maps to a registry icon, a tonal color (a soft
 * same-hue background that deepens on hover), and default tooltip + aria-label text.
 * Because the button is icon-only, `aria-label` (cross-platform) and a hover tooltip
 * (desktop) are always applied, and the coarse-pointer hit area is bumped to >=44px
 * for touch.
 *
 * @example <AppActionButton action="save" :disabled="!canSubmit" :loading="isSaving" @click="onSave" />
 */

type ActionPreset = 'save' | 'create' | 'confirm' | 'cancel' | 'close' | 'delete' | 'edit' | 'reset'
type Tone = 'primary' | 'danger' | 'secondary'

const props = withDefaults(defineProps<{
  /** Which action this button represents (selects icon, tone and default label). */
  action: ActionPreset
  /** Overrides the default tooltip + aria-label text (e.g. "Remove item from set"). */
  label?: string
  /** Overrides the preset tone (used e.g. for a danger-toned confirm). */
  tone?: Tone
  loading?: boolean
  disabled?: boolean
  /** PrimeVue size passthrough; defaults to a compact button. */
  size?: 'small'
}>(), {
  label: undefined,
  tone: undefined,
  loading: false,
  disabled: false,
  size: 'small',
})

defineEmits<{ click: [event: MouseEvent] }>()

const PRESETS: Record<ActionPreset, { icon: ActionIcon, tone: Tone, label: string }> = {
  save: { icon: 'confirm', tone: 'primary', label: 'Save' },
  create: { icon: 'create', tone: 'primary', label: 'Create' },
  confirm: { icon: 'confirm', tone: 'primary', label: 'Confirm' },
  delete: { icon: 'delete', tone: 'danger', label: 'Delete' },
  cancel: { icon: 'cancel', tone: 'secondary', label: 'Cancel' },
  close: { icon: 'close', tone: 'secondary', label: 'Close' },
  edit: { icon: 'edit', tone: 'secondary', label: 'Edit' },
  reset: { icon: 'reset', tone: 'secondary', label: 'Reset' },
}

const preset = computed(() => PRESETS[props.action])
const iconClass = computed(() => `pi ${iconRegistry.action[preset.value.icon]}`)
const text = computed(() => props.label ?? preset.value.label)
const tone = computed<Tone>(() => props.tone ?? preset.value.tone)
</script>

<template>
  <Button v-tooltip.top="text" rounded :size="size" :icon="iconClass" :aria-label="text" :loading="loading"
    :disabled="disabled" :class="['app-action-btn', `app-action-btn-${tone}`, 'pointer-coarse:min-h-11 pointer-coarse:min-w-11']"
    @click="$emit('click', $event)" />
</template>

<style>
/* Tonal styling for the icon-only action button. Not scoped: the rules must reach
   the PrimeVue Button root element (the fallthrough class lands there). Colors come
   only from theme tokens, so the look stays themeable. */
.p-button.app-action-btn {
  background: var(--action-btn-bg);
  border-color: transparent;
  color: var(--action-btn-fg);
}

.p-button.app-action-btn:enabled:hover {
  background: var(--action-btn-bg-hover);
  border-color: transparent;
  color: var(--action-btn-fg);
}

.p-button.app-action-btn .p-button-icon {
  color: inherit;
}

.app-action-btn-primary {
  --action-btn-bg: var(--color-action-primary-bg);
  --action-btn-bg-hover: var(--color-action-primary-bg-hover);
  --action-btn-fg: var(--color-action-primary-fg);
}

.app-action-btn-secondary {
  --action-btn-bg: var(--color-action-secondary-bg);
  --action-btn-bg-hover: var(--color-action-secondary-bg-hover);
  --action-btn-fg: var(--color-action-secondary-fg);
}

.app-action-btn-danger {
  --action-btn-bg: var(--color-action-danger-bg);
  --action-btn-bg-hover: var(--color-action-danger-bg-hover);
  --action-btn-fg: var(--color-action-danger-fg);
}
</style>
