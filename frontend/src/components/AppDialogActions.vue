<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'

const props = withDefaults(defineProps<{
  mode: 'create' | 'edit'
  submitLabel?: string
  canSubmit: boolean
  isPending?: boolean
  isCreating?: boolean
  isUpdating?: boolean
  isDeleting?: boolean
  showDelete?: boolean
  submitDataElement?: string
  cancelDataElement?: string
  deleteDataElement?: string
}>(), {
  isPending: false,
  isCreating: false,
  isUpdating: false,
  isDeleting: false,
  showDelete: true,
  submitDataElement: 'dialog-submit',
  cancelDataElement: 'dialog-cancel',
  deleteDataElement: 'dialog-delete',
})

const submitText = computed(() => props.submitLabel ?? (props.mode === 'create' ? 'Create' : 'Save'))

defineEmits<{
  submit: []
  cancel: []
  delete: []
}>()
</script>

<template>
  <div data-element="dialog-actions" class="mt-4 flex shrink-0 items-center justify-between gap-3">
    <div class="flex flex-wrap items-center gap-2">
      <Button :data-element="submitDataElement" :label="submitText" icon="pi pi-check"
        :disabled="!canSubmit || isPending" :loading="isCreating || isUpdating" @click="$emit('submit')" />
      <Button :data-element="cancelDataElement" label="Cancel" icon="pi pi-times" severity="secondary" outlined
        :disabled="isPending" @click="$emit('cancel')" />
    </div>
    <Button v-if="mode === 'edit' && showDelete" :data-element="deleteDataElement" label="Delete" icon="pi pi-trash"
      severity="danger" outlined class="ml-auto" :disabled="isPending" :loading="isDeleting" @click="$emit('delete')" />
  </div>
</template>
