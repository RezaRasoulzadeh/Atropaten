<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { FileText, Plus } from 'lucide-vue-next'
import AppPanel from '../../components/layout/AppPanel.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import DataTableRow from '../../components/ui/DataTableRow.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SearchField from '../../components/ui/SearchField.vue'
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { invoicesApi, type InvoiceRecord } from '../../api/invoices'
import type { OrderRecord } from '../../api/orders'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'

const { busy, runAction } = useWorkspaceActions()
const props = defineProps<{ currencyUnit: CurrencyUnit; orders: OrderRecord[] }>()
const emit = defineEmits<{
  notify: [string]
  refreshOrders: []
  'open-invoice': [string]
}>()

const rows = ref<InvoiceRecord[]>([])
const query = ref('')
const status = ref('All')
const loading = ref(false)

const filtered = computed(() => {
  const search = query.value.trim().toLowerCase()
  return rows.value.filter(
    (value) =>
      (status.value === 'All' || value.status === status.value) &&
      (!search ||
        [value.invoiceNumber, value.customerName, value.orderId]
          .join(' ')
          .toLowerCase()
          .includes(search)),
  )
})
const readyOrders = computed(() =>
  props.orders.filter((value) => !value.invoiceId && value.totalRial > 0),
)

function tone(value: string) {
  return value === 'Paid' || value === 'Posted'
    ? 'green'
    : value === 'Partially Paid'
      ? 'blue'
      : value === 'Voided'
        ? 'slate'
        : 'amber'
}

function clearFilters() {
  query.value = ''
  status.value = 'All'
}

async function load() {
  loading.value = true
  try {
    rows.value = await invoicesApi.list()
  } catch (error) {
    reportError(error)
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function create(orderId: string) {
  return runAction(async () => {
    try {
      const value = await invoicesApi.createFromOrder(orderId)
      emit('notify', 'Draft invoice created from the saved order snapshot.')
      emit('refreshOrders')
      emit('open-invoice', value.id)
    } catch (error) {
      reportError(error)
    }
  })
}
</script>

<template>
  <div class="min-w-0 space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        eyebrow="Finance / receivables"
        title="Invoices"
        description="Review saved commercial snapshots, payment progress, and posting history."
      >
      </WorkspaceHeader>

      <SearchFilterBar>
        <template #search>
          <SearchField
            v-model="query"
            label="Search invoices"
            placeholder="Invoice, customer, or order"
          />
        </template>
        <template #filters>
          <SelectField
            v-model="status"
            label="Status"
            :options="
              ['All', 'Draft', 'Posted', 'Partially Paid', 'Paid', 'Voided'].map((value) => ({
                label: value,
                value,
              }))
            "
          />
        </template>
        <template #count>
          <span>{{ filtered.length }} of {{ rows.length }} shown</span>
        </template>
        <template #actions>
          <button
            v-if="query || status !== 'All'"
            class="btn btn-ghost btn-sm"
            type="button"
            @click="clearFilters"
          >
            Clear
          </button>
        </template>
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <AppPanel
      title="Invoice register"
      subtitle="Select a row to open the full invoice workspace."
      :flush="true"
    >
      <template #action><span class="text-xs text-base-content/60">{{ filtered.length }} shown</span></template>

      <LoadingState v-if="loading" label="Loading invoices…" />
      <EmptyState
        v-else-if="!filtered.length"
        title="No invoices in this view"
        :description="rows.length ? 'Adjust the search or status filter.' : 'Create an invoice from a priced order below.'"
      >
        <template #icon><FileText :size="22" aria-hidden="true" /></template>
        <template #action>
          <button v-if="rows.length" class="btn btn-primary btn-sm" type="button" @click="clearFilters">
            Clear filters
          </button>
        </template>
      </EmptyState>
      <DataTable v-else label="Invoice register">
        <thead>
          <tr>
            <th scope="col" class="w-[17%]">Invoice</th>
            <th scope="col" class="w-[25%]">Customer</th>
            <th scope="col" class="w-[17%]">Order</th>
            <th scope="col" class="w-[16%]">Issued / due</th>
            <th scope="col" class="w-[8%] text-center">Lines</th>
            <th scope="col" class="w-[17%] text-end">Total</th>
          </tr>
        </thead>
        <tbody>
          <DataTableRow
            v-for="value in filtered"
            :key="value.id"
            interactive
            @activate="emit('open-invoice', value.id)"
          >
            <DataTableCell>
              <strong class="block whitespace-nowrap text-sm">{{ value.invoiceNumber }}</strong>
              <span class="mt-1 block text-xs text-base-content/55">{{ value.items.length }} line items</span>
            </DataTableCell>
            <DataTableCell>
              <strong class="block max-w-64 truncate text-sm font-medium">{{ value.customerName || 'Walk-in customer' }}</strong>
              <StatusBadge class="mt-1" :label="value.status" :tone="tone(value.status)" />
            </DataTableCell>
            <DataTableCell>
              <span class="block max-w-44 truncate text-sm">{{ value.orderId || 'No order link' }}</span>
              <span class="mt-1 block text-xs text-base-content/55">Saved snapshot</span>
            </DataTableCell>
            <DataTableCell>
              <span class="block whitespace-nowrap text-sm">{{ formatDateTime(value.issueDate) }}</span>
              <span class="mt-1 block whitespace-nowrap text-xs text-base-content/55">Due {{ value.dueDate ? formatDateTime(value.dueDate) : 'on receipt' }}</span>
            </DataTableCell>
            <DataTableCell class="text-center">
              <span class="badge badge-ghost min-w-8 justify-center tabular-nums">{{ value.items.length }}</span>
            </DataTableCell>
            <DataTableCell numeric>
              <strong class="text-sm text-primary">{{ formatMoney(value.totalRial, props.currencyUnit) }}</strong>
              <span class="mt-1 block text-xs text-base-content/55">{{ formatMoney(value.remainingRial, props.currencyUnit) }} due</span>
            </DataTableCell>
          </DataTableRow>
        </tbody>
      </DataTable>
    </AppPanel>

    <AppPanel
      title="Orders ready to invoice"
      subtitle="Creating an invoice copies the order's stored pricing snapshots exactly."
      :flush="true"
    >
      <template #action><span class="text-xs text-base-content/60">{{ readyOrders.length }} ready</span></template>
      <div v-if="readyOrders.length" class="divide-y divide-base-300">
        <div
          v-for="order in readyOrders"
          :key="order.id"
          class="flex min-w-0 flex-wrap items-center justify-between gap-3 px-4 py-3 transition-colors hover:bg-base-200/60"
        >
          <div class="min-w-0">
            <strong class="block truncate text-sm">{{ order.orderNumber }} · {{ order.customerName || 'Walk-in customer' }}</strong>
            <span class="mt-1 block truncate text-xs text-base-content/60">
              {{ order.items.length }} items · {{ formatMoney(order.totalRial, props.currencyUnit) }}
            </span>
          </div>
          <button class="btn btn-primary btn-sm shrink-0" type="button" :disabled="busy" @click="create(order.id)">
            <Plus :size="14" /> Create invoice
          </button>
        </div>
      </div>
      <EmptyState v-else title="All priced orders are invoiced" description="New eligible orders will appear here." />
    </AppPanel>
  </div>
</template>
