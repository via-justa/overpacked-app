<script setup lang="ts">
import ItemLabel from './ItemLabel.vue'
import { AppIcon } from '../../../components/icons'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { Item, Label } from '../types'

defineProps<{
  item: Item
  imageSrc: string
  itemLabels?: Label[]
}>()

defineEmits<{
  edit: [item: Item]
}>()

const formatType = (value: string) => {
  return value
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}
</script>

<template>
  <button data-element="item-card" :data-item-id="item.id" type="button"
    class="surface-panel hover:border-brand-300 flex sm:flex-col cursor-pointer overflow-hidden text-left transition"
    :class="{ 'opacity-50': !item.is_active }" @click="$emit('edit', item)">
    <div v-if="imageSrc"
      class="bg-surface-soft aspect-square w-24 shrink-0 self-center overflow-hidden sm:aspect-4/3 sm:w-auto sm:self-stretch">
      <img :src="imageSrc" :alt="normalizeTitleWords(item.name)" class="h-full w-full object-cover" />
    </div>
    <div v-else
      class="text-copy-subtle flex aspect-square w-24 shrink-0 self-center items-center justify-center sm:aspect-4/3 sm:w-auto sm:self-stretch bg-[linear-gradient(135deg,color-mix(in_srgb,var(--color-surface-soft)_92%,var(--color-surface-base)),color-mix(in_srgb,var(--color-line-subtle)_95%,var(--color-surface-base)))]">
      <div
        class="border-line bg-surface-elevated flex flex-col items-center gap-2 rounded-2xl px-0 py-0 sm:border sm:px-5 sm:py-4 sm:shadow-soft">
        <span
          class="border-line text-copy-subtle bg-surface-base flex h-10 w-10 items-center justify-center rounded-full border">
          <AppIcon category="content" name="image" size="md" />
        </span>
        <span
          class="hidden sm:inline text-copy-subtle text-xs font-semibold uppercase tracking-[0.12em]">No Image</span>
      </div>
    </div>
    <div class="min-w-0 flex flex-1 flex-col p-3 sm:p-4">
      <h3 class="text-ink text-base font-semibold line-clamp-2 sm:text-lg sm:line-clamp-none">
        {{ normalizeTitleWords(item.name) }}
      </h3>

      <div class="text-copy-muted mt-1.5 sm:mt-3 space-y-1 text-sm">
        <p class="leading-6">
          <span class="text-copy font-medium">Type:</span>
          <span class="ml-1">{{ formatType(item.type) }}</span>
        </p>

        <slot name="additional-info" />
      </div>

      <div v-if="itemLabels && itemLabels.length > 0" class="mt-auto flex flex-wrap gap-1.5 pt-3">
        <ItemLabel v-for="label in itemLabels" :key="label.id" :label="label" size="sm" />
      </div>
    </div>
  </button>
</template>
