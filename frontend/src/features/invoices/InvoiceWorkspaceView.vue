<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
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
import { printDocument } from '../../utils/print'

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
const printing = ref(false)
const printMode = computed(() => invoice.value?.status === 'Draft' ? 'pre' : 'final')

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

async function printInvoice() {
  if (!invoice.value || loading.value || printing.value) return
  printing.value = true
  try {
    await printDocument()
  } catch (error) {
    reportError(error)
  } finally {
    printing.value = false
  }
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
        :eyebrow='$t("Finance / receivables / invoice")'
        :title="$ui(invoice?.invoiceNumber || 'Invoice workspace')"
        :description="$ui(invoice ? `${$ui(invoice.customerName || 'Walk-in customer')} · ${$ui(invoice.orderId || 'No order link')}` : 'Loading invoice details…')"
      >
        <template #title-suffix>
          <StatusBadge v-if="invoice" :label="invoice.status" :tone="tone(invoice.status)" />
        </template>
        <div class="flex flex-wrap items-center justify-end gap-2">
          <button class="btn btn-primary btn-sm gap-1.5" type="button" :disabled="loading || printing || !invoice" :aria-busy="printing" @click="printInvoice">
            <Printer :size="15" aria-hidden="true" />
            <span>{{ $ui(invoice?.status === 'Draft' ? 'Print pre-invoice' : 'Print final invoice') }}</span>
          </button>
          <button class="btn btn-ghost btn-sm gap-1.5" type="button" @click="emit('back')">
            <ArrowLeft class="rtl-directional-arrow" :size="16" aria-hidden="true" />
            <span>{{ $t("Invoices") }}</span>
          </button>
        </div>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <LoadingState v-if="loading" :label='$t("Loading invoice…")' />
    <EmptyState
      v-else-if="!invoice"
      :title='$t("Invoice unavailable")'
      :description='$t("This invoice could not be loaded. Return to the invoice register and try again.")'
    >
      <template #icon><FileText :size="22" aria-hidden="true" /></template>
      <template #action>
        <button class="btn btn-primary btn-sm" type="button" @click="emit('back')">{{ $t("Back to invoices") }}</button>
      </template>
    </EmptyState>

    <template v-else>
      <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <div class="min-w-0 space-y-4">
          <AppPanel :title='$t("Invoice overview")' :subtitle='$t("Immutable commercial and customer snapshot.")'>
            <div class="grid min-w-0 gap-4 sm:grid-cols-[minmax(0,1fr)_16rem] sm:items-start">
              <div class="flex min-w-0 items-start gap-3">
                <div class="grid size-11 shrink-0 place-items-center rounded-box bg-primary/15 text-primary">
                  <FileText :size="22" aria-hidden="true" />
                </div>
                <div class="min-w-0">
                  <p class="text-xs font-medium uppercase tracking-wide text-primary/80">{{ $t("Commercial snapshot") }}</p>
                  <p class="mt-1 text-2xl font-bold tabular-nums">{{ formatMoney(invoice.totalRial, props.currencyUnit) }}</p>
                  <p class="mt-1 text-xs text-base-content/60">{{ invoice.items.length }} {{ $t("line items ·") }} {{ $ui(invoice.customerName || 'Walk-in customer') }}</p>
                </div>
              </div>
              <dl class="grid min-w-0 gap-2 rounded-box border border-base-300 bg-base-200/30 p-3 text-sm">
                <div class="flex items-center justify-between gap-3">
                  <dt class="text-xs text-base-content/55">{{ $t("Invoice date") }}</dt>
                  <dd class="text-end text-xs">{{ formatDateTime(invoice.issueDate) }}</dd>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <dt class="text-xs text-base-content/55">{{ $t("Due date") }}</dt>
                  <dd class="text-end text-xs">{{ $ui(invoice.dueDate ? formatDateTime(invoice.dueDate) : 'On receipt') }}</dd>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <dt class="text-xs text-base-content/55">{{ $t("Order") }}</dt>
                  <dd class="max-w-36 truncate text-end text-xs">{{ $ui(invoice.orderId || 'No order link') }}</dd>
                </div>
              </dl>
            </div>
          </AppPanel>

          <AppPanel :title='$t("Invoice lines")' :subtitle="$ui(`${invoice.items.length} line items`)" :flush="true">
            <DataTable v-if="invoice.items.length" :label='$t("Invoice lines")'>
              <thead>
                <tr>
                  <th scope="col" class="w-[40%]">{{ $t("Description") }}</th>
                  <th scope="col" class="w-[20%]">{{ $t("Quantity") }}</th>
                  <th scope="col" class="w-[20%] text-end">{{ $t("Unit price") }}</th>
                  <th scope="col" class="w-[20%] text-end">{{ $t("Line total") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="line in invoice.items" :key="line.id">
                  <DataTableCell><span class="block min-w-44 whitespace-normal font-medium">{{ line.description }}</span></DataTableCell>
                  <DataTableCell>{{ formatQuantityUnits(line.quantity) }} {{ $ui(line.quantityUnit) }}</DataTableCell>
                  <DataTableCell numeric>{{ formatMoney(line.unitPriceRial, props.currencyUnit) }}</DataTableCell>
                  <DataTableCell numeric><strong>{{ formatMoney(line.lineTotalRial, props.currencyUnit) }}</strong></DataTableCell>
                </tr>
              </tbody>
            </DataTable>
            <EmptyState v-else compact :title='$t("No invoice lines")' :description='$t("This invoice has no immutable line snapshots.")'><template #icon><FileText :size="21" aria-hidden="true" /></template></EmptyState>
          </AppPanel>

          <AppPanel v-if="invoice.notes" :title='$t("Notes")' :subtitle='$t("Captured when this invoice was created.")'>
            <p class="whitespace-pre-wrap text-sm leading-6">{{ invoice.notes }}</p>
          </AppPanel>
        </div>

        <aside class="min-w-0 space-y-4">
          <AppPanel :title='$t("Totals")' :subtitle='$t("Receivable position")'>
            <dl class="grid min-w-0 gap-1 text-sm">
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2">
                <dt class="text-xs text-base-content/60">{{ $t("Subtotal") }}</dt>
                <dd class="text-end tabular-nums">{{ formatMoney(invoice.subtotalRial, props.currencyUnit) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2">
                <dt class="text-xs text-base-content/60">{{ $t("Discount") }}</dt>
                <dd class="text-end tabular-nums">{{ formatMoney(invoice.discountRial, props.currencyUnit) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2 py-3 font-semibold">
                <dt>{{ $t("Total") }}</dt>
                <dd class="text-end tabular-nums text-primary">{{ formatMoney(invoice.totalRial, props.currencyUnit) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2">
                <dt class="text-xs text-base-content/60">{{ $t("Paid") }}</dt>
                <dd class="text-end tabular-nums text-success">{{ formatMoney(invoice.paidRial, props.currencyUnit) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-4 py-3 font-semibold">
                <dt>{{ $t("Remaining") }}</dt>
                <dd class="text-end tabular-nums text-warning">{{ formatMoney(invoice.remainingRial, props.currencyUnit) }}</dd>
              </div>
            </dl>
          </AppPanel>

          <AppPanel :title='$t("Record details")' :subtitle='$t("Traceable invoice history.")'>
            <dl class="grid min-w-0 gap-3 text-sm">
              <div class="flex items-start gap-2">
                <CalendarDays :size="15" class="mt-0.5 shrink-0 text-base-content/55" aria-hidden="true" />
                <div class="min-w-0"><dt class="text-xs text-base-content/55">{{ $t("Created") }}</dt><dd class="mt-0.5 text-xs">{{ formatDateTime(invoice.createdAt) }}</dd></div>
              </div>
              <div v-if="invoice.accountingJournalEntryId" class="min-w-0">
                <dt class="text-xs text-base-content/55">{{ $t("Receivable journal") }}</dt>
                <dd class="mt-0.5 break-all text-xs">{{ invoice.accountingJournalEntryId }}</dd>
              </div>
              <div v-if="invoice.cogsJournalEntryId" class="min-w-0">
                <dt class="text-xs text-base-content/55">{{ $t("COGS journal") }}</dt>
                <dd class="mt-0.5 break-all text-xs">{{ invoice.cogsJournalEntryId }}</dd>
              </div>
            </dl>
          </AppPanel>
        </aside>
      </div>

      <section class="sticky bottom-0 z-20 flex min-w-0 flex-wrap items-center justify-between gap-3 border border-base-300 bg-base-100 p-3 shadow-lg">
        <p class="hidden text-xs text-base-content/55 sm:block">{{ invoice.invoiceNumber }} · {{ $ui(invoice.status) }}</p>
        <div class="flex w-full flex-wrap justify-end gap-2 sm:w-auto">
          <button v-if="invoice.status === 'Draft'" class="btn btn-outline btn-error btn-sm" type="button" :disabled="busy" @click="remove">
            <Trash2 :size="14" /> {{ $t("Delete draft") }}
          </button>
          <button v-if="['Posted', 'Partially Paid', 'Paid'].includes(invoice.status)" class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="reverse">
            <RotateCcw :size="14" /> {{ $t("Void / reverse") }}
          </button>
          <button v-if="invoice.status === 'Draft'" class="btn btn-primary btn-sm" type="button" :disabled="busy" @click="post">
            <Plus :size="14" /> {{ $t("Post invoice") }}
          </button>
        </div>
      </section>
    </template>

    <Teleport to="body">
      <div v-if="invoice" class="print-output">
        <InvoicePrintDocument :invoice="invoice" :shop="shopSettings" :currency-unit="props.currencyUnit" :document-type="printMode" />
      </div>
    </Teleport>
  </div>
</template>
