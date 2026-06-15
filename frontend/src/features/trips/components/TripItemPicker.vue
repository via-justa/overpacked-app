<script setup lang="ts">
import { ref } from 'vue'
import Button from 'primevue/button'
import AppSelect from '../../../components/forms/AppSelect.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import type { Item } from '../../items/types'
import type { CarryStatus } from '../types'

defineProps<{
    items: Item[]
}>()

const emit = defineEmits<{
    add: [payload: { itemId: string; quantity: number; carryStatus: CarryStatus }]
}>()

const selectedItemId = ref('')
const quantity = ref('1')
const carryStatus = ref<CarryStatus>('packed')

const onAdd = (): void => {
    const parsedQuantity = Number.parseInt(quantity.value, 10)
    if (!selectedItemId.value || !Number.isFinite(parsedQuantity) || parsedQuantity < 1) {
        return
    }
    emit('add', { itemId: selectedItemId.value, quantity: parsedQuantity, carryStatus: carryStatus.value })
    selectedItemId.value = ''
    quantity.value = '1'
    carryStatus.value = 'packed'
}
</script>

<template>
    <div class="grid gap-2 sm:grid-cols-[1fr_6rem_8rem_auto]">
        <AppSelect v-model="selectedItemId">
            <option value="">Select item</option>
            <option v-for="item in items" :key="item.id" :value="item.id">{{ normalizeTitleWords(item.name) }}</option>
        </AppSelect>
        <input v-model="quantity" aria-label="Quantity" class="input-shell" type="number" min="1" step="1" />
        <AppSelect v-model="carryStatus">
            <option value="packed">Packed</option>
            <option value="worn">Worn</option>
        </AppSelect>
        <Button label="Add" size="small" :disabled="!selectedItemId" @click="onAdd" />
    </div>
</template>
