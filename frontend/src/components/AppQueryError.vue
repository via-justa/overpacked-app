<script setup lang="ts">
import type { UseQueryReturnType } from '@tanstack/vue-query'
import Message from 'primevue/message'

const props = defineProps<{
  query: UseQueryReturnType<unknown, Error>
  fallbackMessage: string
  dataElement?: string
}>()

const errorMessage = props.query.error.value instanceof Error 
  ? props.query.error.value.message 
  : props.fallbackMessage
</script>

<template>
  <Message v-if="query.isError.value" :data-element="dataElement" severity="error" :closable="false">
    {{ errorMessage }}
  </Message>
</template>
