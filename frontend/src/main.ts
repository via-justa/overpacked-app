import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import Aura from '@primeuix/themes/aura'
import { VueQueryPlugin } from '@tanstack/vue-query'
import './style.css'
import 'primeicons/primeicons.css'
import App from './App.vue'
import router from './router'
import { setApiAuthFailureHandler } from './lib/api/client'
import { queryClient } from './lib/query/client'
import { pinia } from './stores'
import { useAuthStore } from './stores/auth'

const app = createApp(App)

app.use(pinia)

const authStore = useAuthStore(pinia)
await authStore.ensureBootstrapped()

setApiAuthFailureHandler((reason) => {
	authStore.setAuthNotice(reason)
	authStore.clearSession()

	const redirectPath = `${globalThis.location.pathname}${globalThis.location.search}${globalThis.location.hash}`
	const canRedirectBack = redirectPath.startsWith('/') && !redirectPath.startsWith('/login')

	void router.replace({
		name: 'login',
		query: {
			reason,
			...(canRedirectBack ? { redirect: redirectPath } : {}),
		},
	})
})

app.use(router)
app.use(PrimeVue, {
	theme: {
		preset: Aura,
	},
})
app.use(ToastService)

app.use(VueQueryPlugin, {
	queryClient,
})

app.mount('#app')
