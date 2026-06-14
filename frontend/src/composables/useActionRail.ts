import { ref } from 'vue'
import { getStoredValue, setStoredValue } from '../lib/storage/localStorage'

const STORAGE_KEY = 'overpacked-app.ui.actionRailPinned'

function isBooleanString(value: string): value is 'true' | 'false' {
  return value === 'true' || value === 'false'
}

// Module-level state so the burger button, the rail, and the layout offset all
// share a single source of truth for the pinned-open state.
const pinned = ref(getStoredValue(STORAGE_KEY, isBooleanString, 'false') === 'true')

export function useActionRail() {
  const togglePinned = () => {
    pinned.value = !pinned.value
    setStoredValue(STORAGE_KEY, String(pinned.value))
  }

  return { pinned, togglePinned }
}
