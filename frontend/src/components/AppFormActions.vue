<script setup lang="ts">
import Button from 'primevue/button'

withDefaults(defineProps<{
  submitLabel: string
  canSubmit: boolean
  loading?: boolean
  showCancel?: boolean
  showDelete?: boolean
  isDeleting?: boolean
  submitDataElement?: string
  cancelDataElement?: string
  deleteDataElement?: string
}>(), {
  loading: false,
  showCancel: false,
  showDelete: false,
  isDeleting: false,
  submitDataElement: 'form-submit',
  cancelDataElement: 'form-cancel',
  deleteDataElement: 'form-delete',
})

defineEmits<{
  submit: []
  cancel: []
  delete: []
}>()
</script>

<template>
  <footer data-element="form-actions" class="mt-4 flex flex-wrap items-center gap-2">
    <Button :data-element="submitDataElement" :label="submitLabel" icon="pi pi-check"
      :disabled="!canSubmit || loading" :loading="loading" @click="$emit('submit')" />
    <Button v-if="showCancel" :data-element="cancelDataElement" label="Cancel" icon="pi pi-times"
      severity="secondary" outlined :disabled="loading" @click="$emit('cancel')" />
    <Button v-if="showDelete" :data-element="deleteDataElement" label="Delete" icon="pi pi-trash"
      severity="danger" outlined class="ml-auto" :disabled="loading" :loading="isDeleting" @click="$emit('delete')" />
  </footer>
</template>
