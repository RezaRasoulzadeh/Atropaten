<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ArrowLeft, CalendarDays, FileText, Plus, Printer, RotateCcw, Trash2 } from 'lucide-vue-next'
import AppPanel from '../../components/layout/AppPanel.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { invoicesApi, type InvoiceRecord } from '../../api/invoices'
import { reportsApi, type ShopSettingsRecord } from '../../api/reports'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatQuantityUnits } from '../../utils/quantity'
import { formatDateTime } from '../../utils/date'
import { confirmAction } from '../../ui/feedback'
import InvoicePrintDocument from './InvoicePrintDocument.vue'

const props = defineProps<{ invoiceId: string; currencyUnit: CurrencyUnit }>()
const emit = defineEmits<{
  back: []
  notify: [string]
  refreshOrders: []
}>()
const { busy, runAction } = useWorkspaceActions()

const invoice = ref<InvoiceRecord | null>(null)
const shopSettings = ref<ShopSettingsRecord | null>(null)
const loading = ref(false)

function tone(value: string) {
  return value === 'Paid' || value === 'Posted'
    ? 'green'
    : value === 'Partially Paid'
      ? 'blue'
      : value === 'Voided'
        ? 'slate'
        : 'amber'
}

async function load() {
  loading.value = true
  try {
    invoice.value = await invoicesApi.get(props.invoiceId)
    try {
      shopSettings.value = await reportsApi.settings()
    } catch (error) {
      reportError(error)
    }
  } catch (error) {
    reportError(error)
  } finally {
    loading.value = false
  }
}

function printInvoice() {
  if (!invoice.value || loading.value) return
  window.print()
}

onMounted(load)

async function post() {
  return runAction(async () => {
    if (!invoice.value) return
    try {
      invoice.value = await invoicesApi.post(invoice.value.id)
      emit('notify', 'Invoice posted with AR, revenue, and eligible actual COGS.')
      emit('refreshOrders')
    } catch (error) {
      reportError(error)
    }
  })
}

async function reverse() {
  return runAction(async () => {
    if (
      !invoice.value ||
      !(await confirmAction({
        title: 'Void invoice',
        message: 'Void this invoice with reversing journal entries?',
        confirmLabel: 'Void invoice',
        danger: true,
      }))
    )
      return

    try {
      invoice.value = await invoicesApi.void(invoice.value.id)
      emit('notify', 'Invoice voided with history preserved.')
      emit('refreshOrders')
    } catch (error) {
      reportError(error)
    }
  })
}

async function remove() {
  return runAction(async () => {
    if (
      !invoice.value ||
      !(await confirmAction({
        title: 'Delete draft invoice',
        message: 'Delete this draft invoice permanently?',
        confirmLabel: 'Delete invoice',
        danger: true,
      }))
    )
      return

    try {
      await invoicesApi.deleteDraft(invoice.value.id)
      emit('notify', 'Draft invoice deleted.')
      emit('refreshOrders')
      emit('back')
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
        eyebrow="Finance / receivables / invoice"
        :title="invoice?.invoiceNumber || 'Invoice workspace'"
        :description="invoice ? `${invoice.customerName || 'Walk-in customer'} · ${invoice.orderId || 'No order link'}` : 'Loading invoice details…'"
      >
        <template #title-suffix>
          <StatusBadge v-if="invoice" :label="invoice.status" :tone="tone(invoice.status)" />
        </template>
        <div class="flex flex-wrap items-center justify-end gap-2">
          <button class="btn btn-primary btn-sm gap-1.5" type="button" :disabled="loading" @click="printInvoice">
            <Printer :size="15" aria-hidden="true" />
            <span>Print / save PDF</span>
          </button>
          <button class="btn btn-ghost btn-sm gap-1.5" type="button" @click="emit('back')">
            <ArrowLeft :size="16" aria-hidden="true" />
            <span>Invoices</span>
          </button>
        </div>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <LoadingState v-if="loading" label="Loading invoice…" />
    <EmptyState
      v-else-if="!invoice"
      title="Invoice unavailable"
      description="This invoice could not be loaded. Return to the invoice register and try again."
    >
      <template #icon><FileText :size="22" aria-hidden="true" /></template>
      <template #action>
        <button class="btn btn-primary btn-sm" type="button" @click="emit('back')">Back to invoices</button>
      </template>
    </EmptyState>

    <template v-else>
      <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <div class="min-w-0 space-y-4">
          <AppPanel title="Invoice overview" subtitle="Immutable commercial and customer snapshot.">
            <div class="grid min-w-0 gap-4 sm:grid-cols-[minmax(0,1fr)_16rem] sm:items-start">
              <div class="flex min-w-0 items-start gap-3">
                <div class="grid size-11 shrink-0 place-items-center rounded-box bg-primary/15 text-primary">
                  <FileText :size="22" aria-hidden="true" />
                </div>
                <div class="min-w-0">
                  <p class="text-xs font-medium uppercase tracking-wide text-primary/80">Commercial snapshot</p>
                  <p class="mt-1 text-2xl font-bold tabular-nums">{{ formatMoney(invoice.totalRial, props.currencyUnit) }}</p>
                  <p class="mt-1 text-xs text-base-content/60">{{ invoice.items.length }} line items · {{ invoice.customerName || 'Walk-in customer' }}</p>
                </div>
              </div>
              <dl class="grid min-w-0 gap-2 rounded-box border border-base-300 bg-base-200/30 p-3 text-sm">
                <div class="flex items-center justify-between gap-3">
                  <dt class="text-xs text-base-content/55">Invoice date</dt>
                  <dd class="text-end text-xs">{{ formatDateTime(invoice.issueDate) }}</dd>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <dt class="text-xs text-base-content/55">Due date</dt>
                  <dd class="text-end text-xs">{{ invoice.dueDate ? formatDateTime(invoice.dueDate) : 'On receipt' }}</dd>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <dt class="text-xs text-base-content/55">Order</dt>
                  <dd class="max-w-36 truncate text-end text-xs">{{ invoice.orderId || 'No order link' }}</dd>
                </div>
              </dl>
            </div>
          </AppPanel>

          <AppPanel title="Invoice lines" :subtitle="`${invoice.items.length} immutable snapshot lines`" :flush="true">
            <DataTable v-if="invoice.items.length" label="Invoice lines">
              <thead>
                <tr>
                  <th scope="col" class="w-[40%]">Description</th>
                  <th scope="col" class="w-[20%]">Quantity</th>
                  <th scope="col" class="w-[20%] text-end">Unit price</th>
                  <th scope="col" class="w-[20%] text-end">Line total</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="line in invoice.items" :key="line.id">
                  <DataTableCell><span class="block min-w-44 whitespace-normal font-medium">{{ line.description }}</span></DataTableCell>
                  <DataTableCell>{{ formatQuantityUnits(line.quantity) }} {{ line.quantityUnit }}</DataTableCell>
                  <DataTableCell numeric>{{ formatMoney(line.unitPriceRial, props.currencyUnit) }}</DataTableCell>
                  <DataTableCell numeric><strong>{{ formatMoney(line.lineTotalRial, props.currencyUnit) }}</strong></DataTableCell>
                </tr>
              </tbody>
            </DataTable>
            <EmptyState v-else compact title="No invoice lines" description="This invoice has no immutable line snapshots."><template #icon><FileText :size="21" aria-hidden="true" /></template></EmptyState>
          </AppPanel>

          <AppPanel v-if="invoice.notes" title="Notes" subtitle="Captured when this invoice was created.">
            <p class="whitespace-pre-wrap text-sm leading-6">{{ invoice.notes }}</p>
          </AppPanel>
        </div>

        <aside class="min-w-0 space-y-4">
          <AppPanel title="Totals" subtitle="Receivable position">
            <dl class="grid min-w-0 gap-1 text-sm">
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2">
                <dt class="text-xs text-base-content/60">Subtotal</dt>
                <dd class="text-end tabular-nums">{{ formatMoney(invoice.subtotalRial, props.currencyUnit) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2">
                <dt class="text-xs text-base-content/60">Discount</dt>
                <dd class="text-end tabular-nums">{{ formatMoney(invoice.discountRial, props.currencyUnit) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2 py-3 font-semibold">
                <dt>Total</dt>
                <dd class="text-end tabular-nums text-primary">{{ formatMoney(invoice.totalRial, props.currencyUnit) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2">
                <dt class="text-xs text-base-content/60">Paid</dt>
                <dd class="text-end tabular-nums text-success">{{ formatMoney(invoice.paidRial, props.currencyUnit) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-4 py-3 font-semibold">
                <dt>Remaining</dt>
                <dd class="text-end tabular-nums text-warning">{{ formatMoney(invoice.remainingRial, props.currencyUnit) }}</dd>
              </div>
            </dl>
          </AppPanel>

          <AppPanel title="Record details" subtitle="Traceable invoice history.">
            <dl class="grid min-w-0 gap-3 text-sm">
              <div class="flex items-start gap-2">
                <CalendarDays :size="15" class="mt-0.5 shrink-0 text-base-content/55" aria-hidden="true" />
                <div class="min-w-0"><dt class="text-xs text-base-content/55">Created</dt><dd class="mt-0.5 text-xs">{{ formatDateTime(invoice.createdAt) }}</dd></div>
              </div>
              <div v-if="invoice.accountingJournalEntryId" class="min-w-0">
                <dt class="text-xs text-base-content/55">Receivable journal</dt>
                <dd class="mt-0.5 break-all text-xs">{{ invoice.accountingJournalEntryId }}</dd>
              </div>
              <div v-if="invoice.cogsJournalEntryId" class="min-w-0">
                <dt class="text-xs text-base-content/55">COGS journal</dt>
                <dd class="mt-0.5 break-all text-xs">{{ invoice.cogsJournalEntryId }}</dd>
              </div>
            </dl>
          </AppPanel>
        </aside>
      </div>

      <section class="sticky bottom-0 z-20 flex min-w-0 flex-wrap items-center justify-between gap-3 border border-base-300 bg-base-100 p-3 shadow-lg">
        <p class="hidden text-xs text-base-content/55 sm:block">{{ invoice.invoiceNumber }} · {{ invoice.status }}</p>
        <div class="flex w-full flex-wrap justify-end gap-2 sm:w-auto">
          <button v-if="invoice.status === 'Draft'" class="btn btn-outline btn-error btn-sm" type="button" :disabled="busy" @click="remove">
            <Trash2 :size="14" /> Delete draft
          </button>
          <button v-if="['Posted', 'Partially Paid', 'Paid'].includes(invoice.status)" class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="reverse">
            <RotateCcw :size="14" /> Void / reverse
          </button>
          <button v-if="invoice.status === 'Draft'" class="btn btn-primary btn-sm" type="button" :disabled="busy" @click="post">
            <Plus :size="14" /> Post invoice
          </button>
        </div>
      </section>
    </template>

    <Teleport to="body">
      <div v-if="invoice" class="print-output">
        <InvoicePrintDocument :invoice="invoice" :shop="shopSettings" :currency-unit="props.currencyUnit" />
      </div>
    </Teleport>
  </div>
</template>
