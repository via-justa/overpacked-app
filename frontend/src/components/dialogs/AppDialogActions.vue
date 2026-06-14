<script setup lang="ts">
import { computed } from 'vue'
import AppActionButton from '../actions/AppActionButton.vue'
import AppActionCluster from '../actions/AppActionCluster.vue'

// Reusable dialog action buttons (submit/cancel/delete) rendered as a top-right
// icon cluster. Keeps the original prop/emit surface so existing dialogs migrate
// without template changes; the host panel must be `relative` (see wrappers).
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

const submitAction = computed(() => (props.mode === 'create' ? 'create' : 'save'))
const loading = computed(() => props.isPending || props.isCreating || props.isUpdating || props.isDeleting)

defineEmits<{
  submit: []
  cancel: []
  delete: []
}>()
</script>

<template>
  <AppActionCluster data-element="dialog-actions">
    <AppActionButton v-if="mode === 'edit' && showDelete" action="delete" :data-element="deleteDataElement"
      :disabled="loading" :loading="isDeleting" @click="$emit('delete')" />
    <AppActionButton action="cancel" :data-element="cancelDataElement" :disabled="loading" @click="$emit('cancel')" />
    <AppActionButton :action="submitAction" :label="submitLabel" :data-element="submitDataElement"
      :disabled="!canSubmit || loading" :loading="loading" @click="$emit('submit')" />
  </AppActionCluster>
</template>
