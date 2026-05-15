<script setup lang="ts">
/**
 * Generic card component for person/item/set cards.
 * Provides consistent styling with slots for image, content, and actions.
 */
import { normalizeTitleWords } from '../lib/text/normalize'

interface Props {
  title: string
  imageSrc?: string
  imgAlt?: string
  showImagePlaceholder?: boolean
}

withDefaults(defineProps<Props>(), {
  showImagePlaceholder: false,
})
</script>

<template>
  <article class="surface-panel p-4">
    <!-- Optional image slot -->
    <div v-if="imageSrc" class="bg-surface-soft -m-4 mb-4 aspect-4/3 overflow-hidden rounded-t-2xl">
      <img :src="imageSrc" :alt="imgAlt ?? title" class="h-full w-full object-cover" />
    </div>
    <div v-else-if="showImagePlaceholder"
      class="text-copy-subtle -m-4 mb-4 flex aspect-4/3 items-center justify-center rounded-t-2xl bg-[linear-gradient(135deg,color-mix(in_srgb,var(--color-surface-soft)_92%,var(--color-surface-base)),color-mix(in_srgb,var(--color-line-subtle)_95%,var(--color-surface-base)))]">
      <div
        class="border-line bg-surface-elevated flex flex-col items-center gap-2 rounded-2xl border px-5 py-4 shadow-soft">
        <span
          class="border-line text-copy-subtle bg-surface-base flex h-10 w-10 items-center justify-center rounded-full border">
          <slot name="placeholder-icon">
            <i class="pi pi-image text-base" />
          </slot>
        </span>
        <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.12em]">
          <slot name="placeholder-text">No Image</slot>
        </span>
      </div>
    </div>

    <!-- Title -->
    <h3 class="text-ink text-lg font-semibold">{{ normalizeTitleWords(title) }}</h3>

    <!-- Metadata/content slot -->
    <div v-if="$slots.default" class="text-copy-muted mt-2 text-sm">
      <slot />
    </div>

    <!-- Actions slot -->
    <div v-if="$slots.actions" class="mt-4">
      <slot name="actions" />
    </div>
  </article>
</template>
