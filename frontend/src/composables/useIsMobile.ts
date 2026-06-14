import { onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

// Tailwind `md` breakpoint — the single source of truth for "mobile" across the app.
export const MOBILE_BREAKPOINT = 768

/**
 * Reactive flag that is `true` while the viewport is narrower than the Tailwind
 * `md` breakpoint (768px). Tracks window resizes and is SSR-safe (defaults to
 * non-mobile when there is no `window`). Call from a component's `setup`.
 */
export function useIsMobile(): Ref<boolean> {
  const isMobile = ref(false)

  const update = () => {
    isMobile.value = (globalThis.window?.innerWidth ?? Number.POSITIVE_INFINITY) < MOBILE_BREAKPOINT
  }

  onMounted(() => {
    update()
    globalThis.window?.addEventListener('resize', update)
  })

  onBeforeUnmount(() => {
    globalThis.window?.removeEventListener('resize', update)
  })

  return isMobile
}
