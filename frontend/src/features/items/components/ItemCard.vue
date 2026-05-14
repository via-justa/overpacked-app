<script setup lang="ts">
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { Item } from '../types'

defineProps<{
  item: Item
  imageSrc: string
}>()

defineEmits<{
  openDetails: [item: Item]
}>()

const formatType = (value: string) => {
  return value
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}
</script>

<template>
  <article data-element="item-card" :data-item-id="item.id"
    class="surface-panel overflow-hidden">
    <div v-if="imageSrc" class="bg-surface-soft aspect-4/3 overflow-hidden">
      <img :src="imageSrc" :alt="normalizeTitleWords(item.name)" class="h-full w-full object-cover" />
    </div>
    <div v-else
      class="text-copy-subtle flex aspect-4/3 items-center justify-center bg-[linear-gradient(135deg,color-mix(in_srgb,var(--color-surface-soft)_92%,var(--color-surface-base)),color-mix(in_srgb,var(--color-line-subtle)_95%,var(--color-surface-base)))]">
      <div
        class="border-line bg-surface-elevated flex flex-col items-center gap-2 rounded-2xl border px-5 py-4 shadow-soft">
        <span
          class="border-line text-copy-subtle bg-surface-base flex h-10 w-10 items-center justify-center rounded-full border">
          <i class="pi pi-image text-base" />
        </span>
        <span class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.12em]">No Image</span>
      </div>
    </div>
    <div class="p-4">
      <button type="button"
        class="text-brand-500 decoration-brand-200 text-left text-lg font-semibold underline underline-offset-2"
        @click="$emit('openDetails', item)">
        {{ normalizeTitleWords(item.name) }}
      </button>

      <div class="text-copy-muted mt-3 space-y-1 text-sm">
        <p class="leading-6">
          <span class="text-copy font-medium">Type:</span>
          <span class="ml-1">{{ formatType(item.type) }}</span>
        </p>

        <slot name="additional-info" />
      </div>
    </div>
  </article>
</template>
