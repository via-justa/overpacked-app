<script setup lang="ts">
import { computed } from 'vue'
import Dialog from 'primevue/dialog'

const props = withDefaults(defineProps<{
  modelValue: boolean
  width?: string
  modal?: boolean
  dismissibleMask?: boolean
  draggable?: boolean
  dataElement?: string
}>(), {
  width: 'min(36rem, calc(100vw - 2rem))',
  modal: true,
  dismissibleMask: true,
  draggable: false,
  dataElement: 'template-dialog',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  hide: []
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const onHide = () => {
  emit('hide')
}
</script>

<template>
  <Dialog v-model:visible="visible" :data-element="dataElement" :modal="modal" :dismissible-mask="dismissibleMask"
    :show-header="false" :draggable="draggable" :closable="false" :style="{ width }" :pt="{
      root: { class: '!bg-transparent !shadow-none !border-0' },
      content: { class: '!bg-transparent !p-0' },
    }" @hide="onHide">
    <slot />
  </Dialog>
</template>