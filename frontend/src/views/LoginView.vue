<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import { useForm } from 'vee-validate'
import { useMutation } from '@tanstack/vue-query'
import { z } from 'zod'
import { iconRegistry } from '../lib/icons'
import { loginAuth } from '../lib/api/auth'
import { buildTypedSchema } from '../lib/validation/schema'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const submitError = ref('')
const toast = useToast()

const authNoticeReason = computed(() => {
  const reason = route.query.reason
  if (reason === 'session_expired') {
    return reason
  }

  return null
})

const authNoticeMessage = computed(() => {
  if (authNoticeReason.value === 'session_expired') {
    return 'Your session expired. Please sign in again.'
  }

  return ''
})

watch(
  authNoticeReason,
  (reason) => {
    if (reason !== 'session_expired') {
      return
    }

    toast.add({
      severity: 'warn',
      summary: 'Session expired',
      detail: 'Please sign in again to continue.',
      life: 4500,
    })
  },
  { immediate: true },
)

const loginMutation = useMutation({
  mutationFn: loginAuth,
})

const isSubmitting = computed(() => loginMutation.isPending.value)

const redirectPath = computed(() => {
  const redirect = route.query.redirect

  if (typeof redirect === 'string' && redirect.startsWith('/')) {
    return redirect
  }

  return '/dashboard'
})

const { defineField, handleSubmit, errors, meta } = useForm({
  validationSchema: buildTypedSchema(
    z.object({
      username: z.string().min(1, 'Username is required'),
      password: z.string().min(1, 'Password is required'),
    }),
  ),
})

const [username, usernameProps] = defineField('username')
const [password, passwordProps] = defineField('password')

const onSubmit = handleSubmit(async (values) => {
  submitError.value = ''

  try {
    const response = await loginMutation.mutateAsync(values)
    authStore.setSessionFromTokens(response)
    await router.replace(redirectPath.value)
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : 'Unable to sign in'
  }
})
</script>

<template>
  <main data-component="login-view"
    class="relative grid min-h-screen place-items-center bg-cover bg-center bg-no-repeat px-5 py-10"
    style="background-image: url('/login.jpg')">
    <!-- Dark overlay to keep the form readable over the photo -->
    <div class="absolute inset-0 bg-black/40" aria-hidden="true"></div>

    <form data-element="login-form" class="surface-panel relative z-10 w-full max-w-md p-7" @submit.prevent="onSubmit">
      <img src="/logo.png" alt="Overpacked" class="mx-auto h-15 w-auto object-contain" />

      <h1 class="text-ink mt-4 text-3xl font-bold">Sign in</h1>

      <Message v-if="authNoticeMessage" severity="warn" class="mt-4" :closable="false">
        {{ authNoticeMessage }}
      </Message>

      <div class="mt-6 space-y-4">
        <div>
          <label class="text-copy mb-1 block text-sm font-medium" for="username">Username</label>
          <input id="username" data-element="login-username" v-model="username" v-bind="usernameProps"
            class="input-shell w-full" type="text" name="username" autocomplete="username"
            :aria-invalid="Boolean(errors.username)"
            :aria-describedby="errors.username ? 'username-error' : undefined" />
          <p v-if="errors.username" id="username-error" class="text-danger-500 mt-1 text-xs font-medium">{{
            errors.username }}</p>
        </div>

        <div>
          <label class="text-copy mb-1 block text-sm font-medium" for="password">Password</label>
          <input id="password" data-element="login-password" v-model="password" v-bind="passwordProps"
            class="input-shell w-full" type="password" name="password" autocomplete="current-password"
            :aria-invalid="Boolean(errors.password)"
            :aria-describedby="errors.password ? 'password-error' : undefined" />
          <p v-if="errors.password" id="password-error" class="text-danger-500 mt-1 text-xs font-medium">{{
            errors.password }}</p>
        </div>
      </div>

      <p v-if="submitError" class="text-danger-500 mt-3 text-sm font-medium">{{ submitError }}</p>

      <Button type="submit" data-element="login-submit" class="mt-6 w-full" label="Sign in"
        :icon="`pi ${iconRegistry.navigation.login}`" :disabled="!meta.valid || isSubmitting" :loading="isSubmitting" />
    </form>

    <!-- Photo credit watermark -->
    <a href="https://unsplash.com/@francescafrann" target="_blank" rel="noopener noreferrer"
      class="text-ink-inverse/70 hover:text-ink-inverse absolute bottom-3 right-3 z-10 text-xs">
      Photo by Francesca Pieleanu @ Unsplash
    </a>
  </main>
</template>
