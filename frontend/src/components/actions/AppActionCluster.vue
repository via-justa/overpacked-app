<script setup lang="ts">
/**
 * AppActionCluster - positions a group of AppActionButtons in the top-right
 * corner of its parent (macOS-style window controls).
 *
 * The parent must be `position: relative` and reserve right padding on its
 * header so the title does not sit under the cluster. Place AppActionButtons
 * in the default slot in left-to-right reading order (e.g. Delete, Cancel, Save).
 */
withDefaults(defineProps<{ dataElement?: string }>(), {
  dataElement: 'action-cluster',
})
</script>

<template>
  <div :data-element="dataElement" class="absolute right-2 top-2 z-10 flex items-center gap-1">
    <slot />
  </div>
</template>
