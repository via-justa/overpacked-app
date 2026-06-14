<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { globalSearch, type SearchEntityType, type SearchResult } from '../../features/search/api/searchApi'
import { AppIcon } from '../icons'
import type { IconCategory } from '../../lib/icons'

const SEARCH_DEBOUNCE_MS = 250
const SEARCH_MIN_LENGTH = 2
const SEARCH_LIMIT = 20

type EntityDisplay = {
    label: string
    iconCategory: IconCategory
    iconName: string
    clickable: boolean
}

const entityConfig: Record<SearchEntityType, EntityDisplay> = {
    item: { label: 'Item', iconCategory: 'navigation', iconName: 'gear', clickable: true },
    set: { label: 'Set', iconCategory: 'navigation', iconName: 'sets', clickable: true },
    packing_list: { label: 'Packing List', iconCategory: 'navigation', iconName: 'lists', clickable: true },
    person: { label: 'Person', iconCategory: 'navigation', iconName: 'person', clickable: true },
    manufacturer: { label: 'Manufacturer', iconCategory: 'content', iconName: 'building', clickable: true },
    trip: { label: 'Trip', iconCategory: 'navigation', iconName: 'planner', clickable: false },
}

const filterChips: { value: SearchEntityType; label: string }[] = [
    { value: 'item', label: 'Items' },
    { value: 'set', label: 'Sets' },
    { value: 'packing_list', label: 'Lists' },
    { value: 'person', label: 'Persons' },
    { value: 'manufacturer', label: 'Manufacturers' },
    { value: 'trip', label: 'Trips' },
]

const router = useRouter()

const queryText = ref('')
const debouncedQuery = ref('')
const selectedTypes = ref<SearchEntityType[]>([])
const isOpen = ref(false)
const focusedIndex = ref(-1)
const containerRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

// Debounce the raw input so we only query after the user pauses typing
watch(queryText, (value) => {
    if (debounceTimer) {
        clearTimeout(debounceTimer)
    }
    debounceTimer = setTimeout(() => {
        debouncedQuery.value = value.trim()
    }, SEARCH_DEBOUNCE_MS)
})

const isQueryValid = computed(() => debouncedQuery.value.length >= SEARCH_MIN_LENGTH)

const searchQuery = useQuery({
    queryKey: computed(() => ['global-search', debouncedQuery.value, selectedTypes.value.join(',')]),
    queryFn: () => globalSearch(debouncedQuery.value, selectedTypes.value, SEARCH_LIMIT),
    enabled: isQueryValid,
})

const results = computed<SearchResult[]>(() => searchQuery.data.value ?? [])
const showDropdown = computed(() => isOpen.value && isQueryValid.value)
const isLoading = computed(() => searchQuery.isFetching.value)
const hasNoResults = computed(() => isQueryValid.value && !isLoading.value && results.value.length === 0)

// Reset keyboard focus whenever the result set changes
watch(results, () => {
    focusedIndex.value = -1
})

const isTypeSelected = (type: SearchEntityType) => selectedTypes.value.includes(type)

const toggleType = (type: SearchEntityType) => {
    const next = new Set(selectedTypes.value)
    if (next.has(type)) {
        next.delete(type)
    } else {
        next.add(type)
    }
    selectedTypes.value = [...next]
}

// Map a result to its in-place navigation target; trips have no destination
const navTargetFor = (result: SearchResult): { path: string; query?: Record<string, string> } | null => {
    switch (result.entity_type) {
        case 'item':
            return { path: '/gear', query: { open: result.id } }
        case 'set':
            return { path: '/sets', query: { open: result.id } }
        case 'packing_list':
            return { path: '/lists', query: { open: result.id } }
        case 'person':
            return { path: '/persons', query: { open: result.id } }
        case 'manufacturer':
            return { path: '/gear', query: { action: 'manufacturers' } }
        case 'trip':
            return null
    }
}

const closeDropdown = () => {
    isOpen.value = false
    focusedIndex.value = -1
}

const onSelectResult = async (result: SearchResult) => {
    const target = navTargetFor(result)
    if (!target) {
        return
    }
    closeDropdown()
    queryText.value = ''
    debouncedQuery.value = ''
    await router.push(target)
}

const onInputFocus = () => {
    isOpen.value = true
}

// Keyboard navigation over the flat result list
const onKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
        closeDropdown()
        inputRef.value?.blur()
        return
    }

    if (!showDropdown.value || results.value.length === 0) {
        return
    }

    switch (event.key) {
        case 'ArrowDown':
            event.preventDefault()
            focusedIndex.value = (focusedIndex.value + 1) % results.value.length
            break
        case 'ArrowUp':
            event.preventDefault()
            focusedIndex.value = (focusedIndex.value - 1 + results.value.length) % results.value.length
            break
        case 'Enter': {
            event.preventDefault()
            const result = results.value[focusedIndex.value]
            if (result) {
                void onSelectResult(result)
            }
            break
        }
    }
}

// Close the dropdown when clicking outside the component
const onDocumentClick = (event: MouseEvent) => {
    if (!containerRef.value) {
        return
    }
    if (!containerRef.value.contains(event.target as Node)) {
        closeDropdown()
    }
}

onMounted(() => {
    globalThis.document?.addEventListener('click', onDocumentClick)
})

onBeforeUnmount(() => {
    globalThis.document?.removeEventListener('click', onDocumentClick)
    if (debounceTimer) {
        clearTimeout(debounceTimer)
    }
})
</script>

<template>
    <div ref="containerRef" data-component="app-global-search" class="relative w-full max-w-md">
        <div class="relative">
            <span class="text-copy-muted pointer-events-none absolute inset-y-0 left-3 flex items-center">
                <AppIcon category="action" name="search" size="sm" />
            </span>
            <input ref="inputRef" v-model="queryText" type="search" data-element="global-search-input"
                placeholder="Search gear, sets, people…" aria-label="Global search"
                class="border-line-subtle bg-surface-soft text-copy placeholder:text-copy-muted focus:border-brand-500 focus:ring-brand-500 w-full rounded-lg border py-1.5 pl-9 pr-3 text-sm outline-none transition focus:ring-1"
                @focus="onInputFocus" @keydown="onKeyDown" />
        </div>

        <div v-if="showDropdown" data-element="global-search-dropdown"
            class="border-line-subtle bg-surface-elevated shadow-panel absolute left-0 right-0 top-full z-50 mt-2 overflow-hidden rounded-xl border">
            <div class="border-line-subtle flex flex-wrap gap-1 border-b px-3 py-2">
                <button v-for="chip in filterChips" :key="chip.value" type="button" :data-filter-chip="chip.value"
                    :aria-pressed="isTypeSelected(chip.value)"
                    class="rounded-full px-2.5 py-0.5 text-xs font-medium transition" :class="isTypeSelected(chip.value)
                        ? 'bg-surface-inverse text-ink-inverse'
                        : 'bg-surface-soft text-copy hover:bg-surface-muted'" @click="toggleType(chip.value)">
                    {{ chip.label }}
                </button>
            </div>

            <div class="max-h-80 overflow-y-auto py-1">
                <output v-if="isLoading" class="text-copy-muted block px-3 py-3 text-sm">Searching…</output>

                <p v-else-if="hasNoResults" class="text-copy-muted px-3 py-3 text-sm">No results found.</p>

                <button v-for="(result, index) in results" v-else :key="result.entity_type + result.id" type="button"
                    :aria-current="index === focusedIndex ? 'true' : undefined"
                    :disabled="!entityConfig[result.entity_type].clickable" :data-result-type="result.entity_type"
                    class="flex w-full items-center gap-3 px-3 py-2 text-left transition disabled:cursor-default disabled:opacity-60"
                    :class="index === focusedIndex ? 'bg-surface-soft' : 'hover:bg-surface-soft'"
                    @mouseenter="focusedIndex = index" @click="onSelectResult(result)">
                    <span
                        class="bg-brand-50 text-brand-500 inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-full"
                        aria-hidden="true">
                        <AppIcon :category="entityConfig[result.entity_type].iconCategory"
                            :name="entityConfig[result.entity_type].iconName" size="sm" />
                    </span>
                    <span class="min-w-0 flex-1">
                        <span class="text-ink block truncate text-sm font-medium">{{ result.title }}</span>
                        <span v-if="result.subtitle" class="text-copy-muted block truncate text-xs">{{ result.subtitle
                            }}</span>
                    </span>
                    <span class="text-copy-subtle shrink-0 text-xs font-semibold uppercase tracking-[0.08em]">
                        {{ entityConfig[result.entity_type].label }}
                    </span>
                </button>
            </div>
        </div>
    </div>
</template>
