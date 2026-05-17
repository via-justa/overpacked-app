<script setup lang="ts">
import Button from 'primevue/button'
import { iconRegistry } from '../../lib/icons'

const props = withDefaults(defineProps<{
  visible?: boolean
  label?: string
  icon?: string
  size?: 'small' | 'large'
  reserveSpace?: boolean
  dataElement?: string
}>(), {
  visible: true,
  label: 'Create',
  icon: `pi ${iconRegistry.action.create}`,
  size: 'small',
  reserveSpace: true,
  dataElement: 'app-create-button',
})

const emit = defineEmits<{
  click: []
}>()

const onClick = () => {
  if (!props.visible) {
    return
  }

  emit('click')
}
</script>

<template>
  <Button :data-element="dataElement" :label="label" :icon="icon" :size="size" :disabled="!visible"
    :class="!visible && reserveSpace ? 'invisible pointer-events-none' : ''" @click="onClick" />
</template>