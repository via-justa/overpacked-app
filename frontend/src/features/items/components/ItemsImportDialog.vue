<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useMutation, useQuery } from '@tanstack/vue-query'
import Papa from 'papaparse'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import { iconRegistry } from '../../../lib/icons'
import AppSelect from '../../../components/forms/AppSelect.vue'
import AppTemplateDialog from '../../../components/dialogs/AppTemplateDialog.vue'
import { normalizeTitleWords } from '../../../lib/text/normalize'
import { queryClient } from '../../../lib/query/client'
import { createItem, createItemType, createManufacturer, listItemTypes, listManufacturers } from '../api/itemsApi'
import type { ItemCreate } from '../types'
import { slugifyCategoryId } from '../utils/itemUtils'

// ─── types ────────────────────────────────────────────────────────────────────

type ImportCsvRow = {
  name?: string
  manufacturer?: string
  product?: string
  item_type?: string
  weight_grams?: string
  unit?: string
  price?: string
  consumable?: string
  source_url?: string
  description?: string
}
type ImportFieldKey = keyof ImportCsvRow
type ImportFieldMapping = Record<ImportFieldKey, string>
type ImportColumnTarget = ImportFieldKey | ''
type ImportFieldConfig = { key: ImportFieldKey; label: string; required: boolean }
type ParsedImportCsvFile = { rows: Array<Record<string, string | undefined>>; columns: string[] }
type ImportItemDraft = {
  rowNumber: number
  itemName: string
  manufacturerName: string
  categoryName: string
  typeId: string
  weightGrams?: number
  value?: number
  sourceUrl?: string
  description?: string
}
type ImportPreviewRow = {
  rowNumber: number
  itemName: string
  manufacturerName: string
  categoryName: string
  typeId: string
}
type ParsedImportRow = { draft?: ImportItemDraft; skippedReason?: string }
type ImportRuntimeContext = {
  manufacturerIdByName: Map<string, string>
  knownTypeIds: Set<string>
}

// ─── constants ────────────────────────────────────────────────────────────────

const GRAMS_PER_OUNCE = 28.349523125
const IMPORT_PREVIEW_LIMIT = 20
const importSummaryNoFile = 'Choose a CSV file to preview importable rows.'
const importSummaryNoRows = 'No valid rows found. Fix CSV values and try again.'
const importInvalidCsvError = 'Unable to parse CSV file.'
const importUnsupportedWeightUnitError = 'Unsupported weight unit. Use g, kg, oz, lb, lbs, or gram.'
const importMissingNameError = 'missing item name'
const importMissingManufacturerError = 'missing manufacturer'
const importMissingCategoryError = 'missing category'
const importInvalidPriceError = 'invalid price value'
const importUnexpectedError = 'unexpected error'
const importMissingMappingErrorPrefix = 'Map required fields before import:'

const importFieldConfigs: ImportFieldConfig[] = [
  { key: 'name', label: 'name', required: true },
  { key: 'manufacturer', label: 'manufacturer', required: true },
  { key: 'item_type', label: 'item_type', required: true },
  { key: 'product', label: 'name (fallback)', required: false },
  { key: 'weight_grams', label: 'weight_grams', required: false },
  { key: 'unit', label: 'weight unit (for conversion)', required: false },
  { key: 'price', label: 'price', required: false },
  { key: 'consumable', label: 'consumable (type hint)', required: false },
  { key: 'source_url', label: 'source_url', required: false },
  { key: 'description', label: 'description', required: false },
]

const importedCategoryAliases: Record<string, 'consumable' | 'electronics' | 'shelter' | 'sleep' | 'wearable'> = {
  consumable: 'consumable',
  electronics: 'electronics',
  shelter: 'shelter',
  sleep: 'sleep',
  sleep_system: 'sleep',
  wearable: 'wearable',
}

// ─── props / emits ────────────────────────────────────────────────────────────

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const toast = useToast()

// ─── queries ──────────────────────────────────────────────────────────────────

const manufacturersQuery = useQuery({
  queryKey: ['manufacturers'],
  queryFn: listManufacturers,
})

const itemTypesQuery = useQuery({
  queryKey: ['item-types'],
  queryFn: listItemTypes,
})

// ─── state ────────────────────────────────────────────────────────────────────

const importFileName = ref('')
const importParseError = ref('')
const importPreviewRows = ref<ImportPreviewRow[]>([])
const importDraftRows = ref<ImportItemDraft[]>([])
const importSkippedRows = ref<string[]>([])
const importAvailableColumns = ref<string[]>([])
const importRawRows = ref<Array<Record<string, string | undefined>>>([])
const importColumnTargets = reactive<Record<string, ImportColumnTarget>>({})

// ─── helpers ─────────────────────────────────────────────────────────────────

const createEmptyImportFieldMapping = (): ImportFieldMapping => ({
  name: '',
  manufacturer: '',
  product: '',
  item_type: '',
  weight_grams: '',
  unit: '',
  price: '',
  consumable: '',
  source_url: '',
  description: '',
})

const normalizeImportValue = (value: string | undefined): string => (value ?? '').trim()

const parseImportBool = (value: string | undefined): boolean => {
  const n = normalizeImportValue(value).toLowerCase()
  return n === 'true' || n === 'yes' || n === '1' || n === 'y'
}

const parseImportNumber = (value: string | undefined): number | undefined => {
  const n = normalizeImportValue(value)
  if (!n) return undefined
  const parsed = Number(n)
  return Number.isNaN(parsed) ? undefined : parsed
}

const toImportedWeightGrams = (value: number, unitValue: string | undefined): number | undefined => {
  const unit = normalizeImportValue(unitValue).toLowerCase()
  if (!unit || unit === 'g' || unit === 'gram' || unit === 'grams') return value
  if (unit === 'kg' || unit === 'kilogram' || unit === 'kilograms') return value * 1000
  if (unit === 'oz' || unit === 'ounce' || unit === 'ounces') return value * GRAMS_PER_OUNCE
  if (unit === 'lb' || unit === 'lbs' || unit === 'pound' || unit === 'pounds') return value * GRAMS_PER_OUNCE * 16
  return undefined
}

const parseImportSourceUrl = (value: string | undefined): string | undefined => {
  const normalized = normalizeImportValue(value)
  if (!normalized) return undefined
  try {
    const parsed = new URL(normalized)
    if (parsed.protocol === 'http:' || parsed.protocol === 'https:') return normalized
  } catch {
    return undefined
  }
  return undefined
}

const resolveImportedTypeId = (categoryName: string, consumable: boolean): string => {
  if (consumable) return 'consumable'
  const categorySlug = slugifyCategoryId(categoryName)
  const aliasType = importedCategoryAliases[categorySlug]
  if (aliasType) return aliasType
  return categorySlug
}

// ─── computed ────────────────────────────────────────────────────────────────

const importFieldMapping = computed<ImportFieldMapping>(() => {
  const mapping = createEmptyImportFieldMapping()
  for (const [column, target] of Object.entries(importColumnTargets)) {
    if (!target || mapping[target]) continue
    mapping[target] = column
  }
  return mapping
})

const importMissingRequiredMappings = computed(() =>
  importFieldConfigs
    .filter((c) => c.required && !importFieldMapping.value[c.key])
    .map((c) => c.label),
)

const importSummaryText = computed(() => {
  if (importParseError.value) return importParseError.value
  if (!importFileName.value) return importSummaryNoFile
  if (importMissingRequiredMappings.value.length > 0) {
    return `${importMissingMappingErrorPrefix} ${importMissingRequiredMappings.value.join(', ')}`
  }
  if (importDraftRows.value.length === 0) return importSummaryNoRows
  return `Ready to import ${importDraftRows.value.length} gear items from ${importFileName.value}.`
})

// ─── row parsing ─────────────────────────────────────────────────────────────

const mapRawRowToImportRow = (rawRow: Record<string, string | undefined>, mapping: ImportFieldMapping): ImportCsvRow => {
  const mappedRow: ImportCsvRow = {}
  for (const config of importFieldConfigs) {
    const sourceColumn = mapping[config.key]
    if (!sourceColumn) continue
    const rawValue = rawRow[sourceColumn]
    mappedRow[config.key] = typeof rawValue === 'string' ? rawValue : undefined
  }
  return mappedRow
}

const parseImportWeight = (row: ImportCsvRow, rowNumber: number): { weightGrams?: number; skippedReason?: string } => {
  const rawWeight = parseImportNumber(row.weight_grams)
  if (rawWeight === undefined) return {}
  const convertedWeight = toImportedWeightGrams(rawWeight, row.unit)
  if (convertedWeight === undefined) return { skippedReason: `Row ${rowNumber}: ${importUnsupportedWeightUnitError}` }
  return { weightGrams: Math.round(convertedWeight * 100) / 100 }
}

const parseImportPrice = (row: ImportCsvRow, rowNumber: number): { value?: number; skippedReason?: string } => {
  if (!normalizeImportValue(row.price)) return {}
  const parsedPrice = parseImportNumber(row.price)
  if (parsedPrice === undefined) return { skippedReason: `Row ${rowNumber}: ${importInvalidPriceError}` }
  return { value: parsedPrice }
}

const parseSingleImportRow = (row: ImportCsvRow, rowNumber: number): ParsedImportRow => {
  const itemName = normalizeImportValue(row.name) || normalizeImportValue(row.product)
  const manufacturerName = normalizeImportValue(row.manufacturer)
  const categoryName = normalizeImportValue(row.item_type)

  if (!itemName) return { skippedReason: `Row ${rowNumber}: ${importMissingNameError}` }
  if (!manufacturerName) return { skippedReason: `Row ${rowNumber}: ${importMissingManufacturerError}` }
  if (!categoryName) return { skippedReason: `Row ${rowNumber}: ${importMissingCategoryError}` }

  const weightResult = parseImportWeight(row, rowNumber)
  if (weightResult.skippedReason) return { skippedReason: weightResult.skippedReason }

  const priceResult = parseImportPrice(row, rowNumber)
  if (priceResult.skippedReason) return { skippedReason: priceResult.skippedReason }

  const consumable = parseImportBool(row.consumable as string | undefined)
  const typeId = resolveImportedTypeId(categoryName, consumable)

  return {
    draft: {
      rowNumber,
      itemName,
      manufacturerName,
      categoryName,
      typeId,
      weightGrams: weightResult.weightGrams,
      value: priceResult.value,
      sourceUrl: parseImportSourceUrl(row.source_url),
      description: normalizeImportValue(row.description) || undefined,
    },
  }
}

const parseImportRows = (rows: ImportCsvRow[]): { drafts: ImportItemDraft[]; skipped: string[] } => {
  const drafts: ImportItemDraft[] = []
  const skipped: string[] = []
  for (let index = 0; index < rows.length; index += 1) {
    const row = rows[index]
    const parsed = parseSingleImportRow(row, index + 2)
    if (parsed.skippedReason) {
      skipped.push(parsed.skippedReason)
    } else if (parsed.draft) {
      drafts.push(parsed.draft)
    }
  }
  return { drafts, skipped }
}

const applyParsedImportRows = (rows: ImportCsvRow[]) => {
  const result = parseImportRows(rows)
  importDraftRows.value = result.drafts
  importSkippedRows.value = result.skipped
  importPreviewRows.value = result.drafts.slice(0, IMPORT_PREVIEW_LIMIT).map((row) => ({
    rowNumber: row.rowNumber,
    itemName: row.itemName,
    manufacturerName: row.manufacturerName,
    categoryName: row.categoryName,
    typeId: row.typeId,
  }))
}

const refreshImportFromMapping = () => {
  if (importRawRows.value.length === 0) {
    importPreviewRows.value = []
    importDraftRows.value = []
    importSkippedRows.value = []
    return
  }
  if (importMissingRequiredMappings.value.length > 0) {
    importPreviewRows.value = []
    importDraftRows.value = []
    importSkippedRows.value = []
    return
  }
  const mappedRows = importRawRows.value.map((rawRow) => mapRawRowToImportRow(rawRow, importFieldMapping.value))
  applyParsedImportRows(mappedRows)
}

// ─── column target management ────────────────────────────────────────────────

const isImportTargetSelectedInOtherColumn = (target: ImportFieldKey, column: string): boolean => {
  for (const [candidateColumn, candidateTarget] of Object.entries(importColumnTargets)) {
    if (candidateColumn !== column && candidateTarget === target) return true
  }
  return false
}

const updateImportColumnTarget = (column: string, target: ImportColumnTarget) => {
  importColumnTargets[column] = target
}

watch(
  () => ({ ...importColumnTargets }),
  () => {
    if (!props.open) return
    refreshImportFromMapping()
  },
  { deep: true },
)

// ─── file parsing ─────────────────────────────────────────────────────────────

const parseCsvFile = async (file: File): Promise<ParsedImportCsvFile> =>
  new Promise((resolve, reject) => {
    Papa.parse<Record<string, unknown>>(file, {
      header: true,
      skipEmptyLines: true,
      complete: (result) => {
        if (result.errors.length > 0) {
          reject(new Error(importInvalidCsvError))
          return
        }
        const columns = (result.meta.fields ?? []).map((f) => f.trim()).filter((f) => f.length > 0)
        const rows = result.data.map((row) => {
          const normalizedRow: Record<string, string | undefined> = {}
          for (const column of columns) {
            const value = row[column]
            normalizedRow[column] = typeof value === 'string' ? value.trim() || undefined : undefined
          }
          return normalizedRow
        })
        resolve({ rows, columns })
      },
      error: () => reject(new Error(importInvalidCsvError)),
    })
  })

// ─── dialog lifecycle ─────────────────────────────────────────────────────────

const resetDialogState = () => {
  importFileName.value = ''
  importParseError.value = ''
  importPreviewRows.value = []
  importDraftRows.value = []
  importSkippedRows.value = []
  importAvailableColumns.value = []
  importRawRows.value = []
  for (const key of Object.keys(importColumnTargets)) {
    delete importColumnTargets[key]
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) resetDialogState()
  },
)

const onFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  importFileName.value = file.name
  importParseError.value = ''
  importPreviewRows.value = []
  importDraftRows.value = []
  importSkippedRows.value = []
  importAvailableColumns.value = []
  importRawRows.value = []
  for (const key of Object.keys(importColumnTargets)) {
    delete importColumnTargets[key]
  }

  try {
    const parsedFile = await parseCsvFile(file)
    importAvailableColumns.value = parsedFile.columns
    importRawRows.value = parsedFile.rows
    for (const column of parsedFile.columns) {
      importColumnTargets[column] = ''
    }
    refreshImportFromMapping()
  } catch (error) {
    importParseError.value = error instanceof Error ? error.message : importInvalidCsvError
  } finally {
    input.value = ''
  }
}

// ─── import execution ─────────────────────────────────────────────────────────

const createImportRuntimeContext = (): ImportRuntimeContext => {
  const manufacturerIdByName = new Map<string, string>()
  for (const manufacturer of manufacturersQuery.data.value ?? []) {
    manufacturerIdByName.set(manufacturer.name.trim().toLowerCase(), manufacturer.id)
  }
  const knownTypeIds = new Set((itemTypesQuery.data.value ?? []).map((t) => t.id))
  return { manufacturerIdByName, knownTypeIds }
}

const ensureImportedManufacturerId = async (row: ImportItemDraft, ctx: ImportRuntimeContext): Promise<string> => {
  const key = row.manufacturerName.trim().toLowerCase()
  const existingId = ctx.manufacturerIdByName.get(key)
  if (existingId) return existingId
  const created = await createManufacturer({ name: normalizeTitleWords(row.manufacturerName) })
  ctx.manufacturerIdByName.set(key, created.id)
  return created.id
}

const ensureImportedType = async (row: ImportItemDraft, ctx: ImportRuntimeContext): Promise<void> => {
  if (ctx.knownTypeIds.has(row.typeId)) return
  await createItemType({ id: row.typeId, name: normalizeTitleWords(row.categoryName) })
  ctx.knownTypeIds.add(row.typeId)
}

const buildImportedItemPayload = (row: ImportItemDraft, manufacturerId: string): ItemCreate => {
  const payload: ItemCreate = {
    name: normalizeTitleWords(row.itemName),
    type: row.typeId,
    is_active: true,
    manufacturer_id: manufacturerId,
  }
  if (typeof row.weightGrams === 'number') payload.weight_grams = row.weightGrams
  if (typeof row.value === 'number') payload.value = row.value
  if (row.sourceUrl) payload.source_url = row.sourceUrl
  if (row.description) payload.description = row.description
  return payload
}

const importItemsMutation = useMutation({
  mutationFn: async (rows: ImportItemDraft[]) => {
    const ctx = createImportRuntimeContext()
    const importedErrors: string[] = []
    let importedCount = 0
    for (const row of rows) {
      try {
        const manufacturerId = await ensureImportedManufacturerId(row, ctx)
        await ensureImportedType(row, ctx)
        await createItem(buildImportedItemPayload(row, manufacturerId))
        importedCount += 1
      } catch (error) {
        const message = error instanceof Error ? error.message : importUnexpectedError
        importedErrors.push(`Row ${row.rowNumber}: ${message}`)
      }
    }
    return { importedCount, importedErrors }
  },
  onSuccess: async (result) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['items'] }),
      queryClient.invalidateQueries({ queryKey: ['manufacturers'] }),
      queryClient.invalidateQueries({ queryKey: ['item-types'] }),
    ])
    if (result.importedErrors.length === 0) {
      toast.add({
        severity: 'success',
        summary: 'Import complete',
        detail: `${result.importedCount} gear items imported.`,
        life: 3500,
      })
      emit('update:open', false)
      return
    }
    const firstError = result.importedErrors[0]
    toast.add({
      severity: 'warn',
      summary: 'Import finished with issues',
      detail: `${result.importedCount} gear items imported. ${result.importedErrors.length} rows failed. ${firstError}`,
      life: 5000,
    })
  },
  onError: (error) => {
    toast.add({
      severity: 'error',
      summary: 'Import failed',
      detail: error instanceof Error ? error.message : 'Unable to import gear.',
      life: 4000,
    })
  },
})

const onConfirmImport = async () => {
  if (importDraftRows.value.length === 0 || importItemsMutation.isPending.value) return
  await importItemsMutation.mutateAsync(importDraftRows.value)
}
</script>

<template>
  <AppTemplateDialog :model-value="open" data-element="items-import-dialog" width="min(42rem, calc(100vw - 2rem))"
    @update:model-value="(v) => emit('update:open', v as boolean)" @hide="emit('update:open', false)">
    <section
      class="border-line-subtle bg-surface-elevated flex max-h-[calc(100vh-8rem)] w-full flex-col rounded-2xl border p-4 shadow-panel backdrop-blur sm:p-5">
      <h3 class="text-ink shrink-0 text-lg font-semibold">Import Gear From CSV</h3>
      <p class="text-copy-muted mt-1 shrink-0 text-sm">
        Review parsed rows first, then confirm import. Missing categories and manufacturers are created automatically.
      </p>

      <div class="mt-4 flex-1 space-y-3 overflow-y-auto pr-1">
        <label class="grid gap-1">
          <span class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">CSV file</span>
          <input data-element="items-import-file" type="file" accept=".csv,text/csv"
            class="border-line bg-surface-base text-ink rounded-lg border px-3 py-2 text-sm file:mr-3 file:rounded-lg file:border-0 file:bg-surface-soft file:px-3 file:py-2 file:text-sm file:font-medium file:text-copy hover:file:bg-surface-muted"
            :disabled="importItemsMutation.isPending.value" @change="void onFileChange($event)" />
        </label>

        <Message v-if="importParseError" severity="error" :closable="false">
          {{ importParseError }}
        </Message>

        <div class="border-line-subtle bg-surface-muted text-copy rounded-lg border px-3 py-2 text-sm">
          {{ importSummaryText }}
        </div>

        <div v-if="importAvailableColumns.length > 0" class="space-y-2">
          <div class="grid grid-cols-2 gap-x-4 px-1">
            <p class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.06em]">CSV column</p>
            <p class="text-copy-subtle text-xs font-semibold uppercase tracking-[0.06em]">Maps to field</p>
          </div>
          <div class="border-line-subtle overflow-hidden rounded-lg border">
            <div v-for="(column, columnIndex) in importAvailableColumns" :key="`import-map-column-${column}`"
              class="grid grid-cols-2 items-center gap-x-4 px-3 py-2"
              :class="columnIndex % 2 === 0 ? 'bg-surface-base' : 'bg-surface-muted'">
              <div class="min-w-0">
                <p class="text-copy truncate text-sm" :title="column">{{ column }}</p>
                <p v-if="importRawRows[0]?.[column]" class="text-copy-subtle truncate text-xs"
                  :title="String(importRawRows[0][column])">{{ importRawRows[0][column] }}</p>
              </div>
              <AppSelect :data-element="`items-import-map-column-${column}`" :model-value="importColumnTargets[column]"
                compact :disabled="importItemsMutation.isPending.value"
                @update:model-value="(value) => updateImportColumnTarget(column, value as ImportColumnTarget)">
                <option value="">Not mapped</option>
                <option v-for="field in importFieldConfigs" :key="`${column}-${field.key}`" :value="field.key"
                  :disabled="isImportTargetSelectedInOtherColumn(field.key, column)">{{ field.label }}{{ field.required
                    ? ' *' : '' }}</option>
              </AppSelect>
            </div>
          </div>
        </div>

        <div v-if="importPreviewRows.length > 0" class="space-y-2">
          <p class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Preview</p>
          <div class="border-line-subtle bg-surface-base overflow-hidden rounded-lg border">
            <table class="text-copy w-full text-left text-xs">
              <thead class="bg-surface-muted text-copy-muted">
                <tr>
                  <th class="px-2 py-1.5">Row</th>
                  <th class="px-2 py-1.5">Name</th>
                  <th class="px-2 py-1.5">Manufacturer</th>
                  <th class="px-2 py-1.5">Category</th>
                  <th class="px-2 py-1.5">Type ID</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in importPreviewRows" :key="`import-preview-${row.rowNumber}`"
                  class="border-line-subtle border-t">
                  <td class="px-2 py-1.5">{{ row.rowNumber }}</td>
                  <td class="px-2 py-1.5">{{ row.itemName }}</td>
                  <td class="px-2 py-1.5">{{ row.manufacturerName }}</td>
                  <td class="px-2 py-1.5">{{ row.categoryName }}</td>
                  <td class="px-2 py-1.5">{{ row.typeId }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p v-if="importDraftRows.length > importPreviewRows.length" class="text-copy-subtle text-xs">
            Showing first {{ importPreviewRows.length }} of {{ importDraftRows.length }} valid rows.
          </p>
        </div>

        <div v-if="importSkippedRows.length > 0" class="space-y-2">
          <p class="text-copy text-xs font-semibold uppercase tracking-[0.06em]">Skipped rows</p>
          <ul class="text-warning-900 max-h-32 list-disc space-y-1 overflow-y-auto pl-5 text-xs">
            <li v-for="reason in importSkippedRows.slice(0, 12)" :key="reason">{{ reason }}</li>
          </ul>
          <p v-if="importSkippedRows.length > 12" class="text-copy-subtle text-xs">
            Showing first 12 of {{ importSkippedRows.length }} skipped rows.
          </p>
        </div>
      </div>

      <div class="mt-4 flex shrink-0 items-center gap-2">
        <Button data-element="items-import-confirm" label="Import" :icon="`pi ${iconRegistry.action.upload}`"
          :disabled="importDraftRows.length === 0 || importItemsMutation.isPending.value || importMissingRequiredMappings.length > 0"
          :loading="importItemsMutation.isPending.value" @click="void onConfirmImport()" />
        <Button data-element="items-import-cancel" label="Cancel" :icon="`pi ${iconRegistry.action.cancel}`"
          severity="secondary" outlined :disabled="importItemsMutation.isPending.value"
          @click="emit('update:open', false)" />
      </div>
    </section>
  </AppTemplateDialog>
</template>
