<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowDown, ArrowUp, CheckCheck, Edit3, FileText, Plus, Receipt, Save, ShoppingCart, Trash2 } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import SelectField from '../../components/ui/SelectField.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'
import type { usePurchasesWorkspace } from './usePurchasesWorkspace'

const props = defineProps<{
  workspace: ReturnType<typeof usePurchasesWorkspace>
  currencyUnit: CurrencyUnit
  suppliers: any[]
  materials: any[]
}>()
const {
  busy,
  current,
  createMode,
  editingItemId,
  form,
  item,
  financial,
  cancelEditor,
  save,
  addItem,
  startItemEdit,
  cancelItemEdit,
  removeItem,
  reorder,
} = props.workspace

const activeStep = ref(1)
const steps = [
  { number: 1, title: 'Purchase details', description: 'Supplier, date, account' },
  { number: 2, title: 'Items', description: 'Materials and quantities' },
  { number: 3, title: 'Costs & notes', description: 'Adjustments and context' },
  { number: 4, title: 'Review', description: 'Confirm before saving' },
]

const sortedItems = computed(() => [...(current.value?.items || [])].sort((left, right) => left.position - right.position))
const title = computed(() => createMode.value ? 'Add purchase' : 'Edit purchase')
const purchaseNumber = computed(() => current.value?.purchaseNumber || 'New purchase')
const supplierName = computed(() => current.value?.supplierName || 'No supplier selected')
const itemCount = computed(() => current.value?.items.length || 0)
const total = computed(() => current.value?.totalRial || 0)
const historyLocked = computed(() => Boolean(current.value && current.value.status !== 'Draft'))

function stepClass(number: number) {
  if (number === activeStep.value) return 'wizard-step-active'
  if (number < activeStep.value) return 'wizard-step-complete'
  return 'wizard-step-idle'
}

function next() {
  if (activeStep.value < steps.length) activeStep.value += 1
}

function updatePurchaseMoney(field: 'discountRial' | 'shippingRial' | 'taxRial' | 'additionalCostsRial', value: string) {
  const parsed = parseMoneyInput(value, props.currencyUnit)
  if (parsed !== null) props.workspace.form.value[field] = parsed
}

function financialAccountLabel(id: string) {
  const account = financial.value.find((value) => value.id === id)
  if (!account) return 'No account selected'
  const identity = account.type === 'bank'
    ? [account.bankName, account.accountNumber && `Account ${account.accountNumber}`, account.cardNumber && `Card ${account.cardNumber}`].filter(Boolean).join(' · ')
    : 'Cash account'
  return identity ? `${account.name} · ${identity}` : account.name
}
</script>

<template>
  <div v-if="current" class="service-wizard flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Purchase editor">
    <header class="service-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-semibold tracking-tight text-primary">{{ title }}</h1>
        <p class="mt-1 text-sm text-base-content/65">{{ historyLocked ? 'Safely update purchase notes without changing its inventory history.' : 'Record a supplier purchase and prepare its inventory receipt.' }}</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Purchases' }, { label: title, current: true }]" @navigate="cancelEditor" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="busy" @click="cancelEditor">Cancel</button>
        <button class="btn btn-success gap-2" type="button" :disabled="busy" @click="save"><Save :size="16" aria-hidden="true" />Save purchase</button>
      </div>
    </header>

    <div class="service-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="service-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200">
        <nav aria-label="Purchase setup steps" class="service-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible">
          <button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" @click="activeStep = step.number">
            <span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span>
            <span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ step.title }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ step.description }}</small></span>
          </button>
        </nav>
      </aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="service-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
          <form id="purchase-editor" class="service-editor min-w-0" @submit.prevent="next">
            <section v-if="activeStep === 1" class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Receipt :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Purchase details</h2><p class="text-sm text-base-content/60">{{ historyLocked ? 'Invoice reference and notes remain editable after posting.' : 'Set the supplier, date, payment account, and invoice reference.' }}</p></div></div>
              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <FormField class="gap-1"><span>Supplier <em class="text-error">*</em></span><SelectField v-model="form.supplierId" :disabled="historyLocked" :options="[{ label: 'Select supplier', value: '' }, ...props.suppliers.map((supplier) => ({ label: supplier.name, value: supplier.id }))]" aria-label="Purchase supplier" /><small class="text-xs leading-5 text-base-content/60">The supplier connected to this purchase.</small></FormField>
                <FormField class="gap-1"><span>Purchase date <em class="text-error">*</em></span><JalaliDatePicker v-model="form.purchaseDate" :disabled="historyLocked" placeholder="Select Jalali date" /><small class="text-xs leading-5 text-base-content/60">The date used for purchase and inventory history.</small></FormField>
                <FormField class="gap-1 sm:col-span-2"><span>Paid from account <em class="text-error">*</em></span><SelectField v-model="form.financialAccountId" :disabled="historyLocked" :options="[{ label: 'Select cash or bank account', value: '' }, ...financial.filter((account) => account.active).map((account) => ({ label: financialAccountLabel(account.id), value: account.id }))]" aria-label="Purchase payment account" /><small class="text-xs leading-5 text-base-content/60">The active cash or bank account used for this purchase.</small></FormField>
                <FormField class="gap-1 sm:col-span-2"><span>Supplier invoice number</span><AppInput v-model="form.supplierInvoiceNumber" class="input w-full min-w-0" placeholder="Optional invoice reference" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Keep the supplier’s reference for reconciliation.</small></FormField>
              </div>
            </section>

            <section v-else-if="activeStep === 2" class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><ShoppingCart :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Purchase items</h2><p class="text-sm text-base-content/60">Add the materials and quantities received from the supplier.</p></div></div>
              <div class="space-y-2">
                <div v-for="(line, index) in sortedItems" :key="line.id" class="flex min-w-0 flex-wrap items-center gap-3 rounded-box border border-base-300 bg-base-100/35 px-3 py-3">
                  <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><ShoppingCart :size="17" aria-hidden="true" /></span>
                  <div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ line.materialName }}</strong><span class="block truncate text-xs text-base-content/55">{{ line.purchaseQuantity }} {{ line.purchaseUnit }} → {{ line.consumptionQuantity }} {{ line.consumptionUnit }}<span v-if="line.notes"> · {{ line.notes }}</span></span></div>
                  <span class="shrink-0 text-sm tabular-nums">{{ formatMoney(line.lineTotalRial, props.currencyUnit) }}</span>
                  <div v-if="!historyLocked" class="flex shrink-0 items-center gap-1"><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="busy" aria-label="Edit item" @click="startItemEdit(line)"><Edit3 :size="13" aria-hidden="true" /></button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="busy || index === 0" aria-label="Move item up" @click="reorder(index, -1)"><ArrowUp :size="13" aria-hidden="true" /></button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="busy || index === sortedItems.length - 1" aria-label="Move item down" @click="reorder(index, 1)"><ArrowDown :size="13" aria-hidden="true" /></button><button class="btn btn-ghost btn-xs btn-square text-error" type="button" :disabled="busy" aria-label="Remove item" @click="removeItem(line.id)"><Trash2 :size="13" aria-hidden="true" /></button></div>
                </div>
                <EmptyState v-if="!sortedItems.length" compact title="No items added" description="Add the first material below to build this purchase."><template #icon><ShoppingCart :size="21" aria-hidden="true" /></template></EmptyState>
              </div>
              <div v-if="!historyLocked" class="rounded-box border border-dashed border-base-300 bg-base-100/25 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Plus :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Add material</h3><p class="mt-1 text-xs leading-5 text-base-content/60">The purchase quantity is converted using the selected material’s unit setup.</p></div></div><div class="mt-4 grid min-w-0 gap-3 sm:grid-cols-2"><FormField class="gap-1"><span>Material</span><SelectField v-model="item.materialId" :options="[{ label: 'Select material', value: '' }, ...props.materials.map((material) => ({ label: material.name, value: material.id }))]" aria-label="Purchase material" /></FormField><FormField class="gap-1"><span>Purchase quantity</span><AppInput v-model="item.purchaseQuantity" class="input w-full min-w-0" inputmode="decimal" placeholder="1" /></FormField><FormField class="gap-1"><span>Unit cost ({{ props.currencyUnit }})</span><AppInput v-model="item.unitAcquisitionCostRial" :money="props.currencyUnit" class="input w-full min-w-0" inputmode="numeric" placeholder="0" /></FormField><FormField class="gap-1"><span>Line note</span><AppInput v-model="item.notes" class="input w-full min-w-0" placeholder="Optional note" /></FormField></div><div class="mt-3 flex flex-wrap justify-end gap-2 border-t border-base-300 pt-3"><button v-if="editingItemId" class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="cancelItemEdit">Cancel edit</button><button class="btn btn-outline btn-sm gap-2" type="button" :disabled="busy" @click="addItem"><Plus :size="14" aria-hidden="true" />{{ editingItemId ? 'Update item' : 'Add item' }}</button></div></div>
              <div v-else class="rounded-box border border-info/30 bg-info/10 p-4 text-sm text-base-content/75">This purchase already has inventory history. Its material lines, quantities, date, account, and totals are locked. You can safely update the invoice reference or notes.</div>
            </section>

            <section v-else-if="activeStep === 3" class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Receipt :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Costs &amp; notes</h2><p class="text-sm text-base-content/60">Add purchase-level adjustments and context for your team.</p></div></div>
              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <FormField class="gap-1"><span>Discount ({{ props.currencyUnit }})</span><AppInput :model-value="formatMoneyInput(form.discountRial, props.currencyUnit)" :disabled="historyLocked" :money="props.currencyUnit" class="input w-full min-w-0" inputmode="numeric" @update:model-value="updatePurchaseMoney('discountRial', $event)" /></FormField>
                <FormField class="gap-1"><span>Shipping ({{ props.currencyUnit }})</span><AppInput :model-value="formatMoneyInput(form.shippingRial, props.currencyUnit)" :disabled="historyLocked" :money="props.currencyUnit" class="input w-full min-w-0" inputmode="numeric" @update:model-value="updatePurchaseMoney('shippingRial', $event)" /></FormField>
                <FormField class="gap-1"><span>Tax ({{ props.currencyUnit }})</span><AppInput :model-value="formatMoneyInput(form.taxRial, props.currencyUnit)" :disabled="historyLocked" :money="props.currencyUnit" class="input w-full min-w-0" inputmode="numeric" @update:model-value="updatePurchaseMoney('taxRial', $event)" /></FormField>
                <FormField class="gap-1"><span>Additional costs ({{ props.currencyUnit }})</span><AppInput :model-value="formatMoneyInput(form.additionalCostsRial, props.currencyUnit)" :disabled="historyLocked" :money="props.currencyUnit" class="input w-full min-w-0" inputmode="numeric" @update:model-value="updatePurchaseMoney('additionalCostsRial', $event)" /></FormField>
                <FormField class="gap-1 sm:col-span-2"><span>Notes</span><AppTextarea v-model="form.notes" class="textarea w-full min-w-0" rows="7" placeholder="Delivery details, quality notes, or payment context" /><small class="text-xs leading-5 text-base-content/60">These notes stay with the purchase record.</small></FormField>
              </div>
            </section>

            <section v-else class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><FileText :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Review purchase</h2><p class="text-sm text-base-content/60">Confirm the purchase before saving it.</p></div></div>
              <div class="rounded-box border border-base-300 bg-base-100/35 p-4"><dl class="divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-3"><dt class="text-base-content/60">Supplier</dt><dd class="max-w-[65%] truncate text-end">{{ supplierName }}</dd></div><div class="flex justify-between gap-3 py-3"><dt class="text-base-content/60">Purchase date</dt><dd class="text-end">{{ form.purchaseDate }}</dd></div><div class="flex justify-between gap-3 py-3"><dt class="text-base-content/60">Items</dt><dd class="text-end">{{ itemCount }} line item{{ itemCount === 1 ? '' : 's' }}</dd></div><div class="flex justify-between gap-3 py-3"><dt class="text-base-content/60">Account</dt><dd class="max-w-[65%] truncate text-end">{{ financialAccountLabel(form.financialAccountId) }}</dd></div><div class="flex justify-between gap-3 py-3 last:pb-0"><dt class="text-base-content/60">Total</dt><dd class="font-semibold tabular-nums text-primary">{{ formatMoney(total, props.currencyUnit) }}</dd></div></dl></div><div class="rounded-box border border-info/30 bg-info/10 p-4 text-sm text-base-content/75">{{ historyLocked ? 'This is a history-safe edit. The purchase number and inventory-affecting values remain unchanged.' : 'The purchase remains a draft until you post it. Posting creates the related inventory movements.' }}</div>
            </section>

          </form>
        </section>

        <aside class="service-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto"><div class="space-y-4"><div class="relative min-h-48 overflow-hidden rounded-box bg-base-300 p-5"><div class="absolute inset-0 bg-gradient-to-t from-black/95 via-black/65 to-black/10" aria-hidden="true"></div><div class="absolute inset-0 grid place-items-center text-white/10"><ShoppingCart :size="84" :stroke-width="1" aria-hidden="true" /></div><div class="relative z-10 flex min-h-36 flex-col justify-end text-white"><div class="mt-auto flex min-w-0 items-end justify-start gap-3"><div class="min-w-0"><h2 class="truncate text-lg font-semibold">{{ purchaseNumber }}</h2><p class="mt-1 truncate text-xs text-white/70">{{ supplierName }} · {{ itemCount }} items</p></div><StatusBadge :label="current.status" :tone="current.status === 'Posted' ? 'green' : current.status === 'Draft' ? 'amber' : 'slate'" /></div></div></div><div class="divide-y divide-base-300 rounded-box border border-base-300 bg-base-100/35"><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Supplier</span><strong class="max-w-[10rem] truncate text-end">{{ supplierName }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Items</span><strong>{{ itemCount }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Total</span><strong class="text-primary tabular-nums">{{ formatMoney(total, props.currencyUnit) }}</strong></div></div><div class="flex gap-2 border-t border-base-300 pt-3 text-sm"><Receipt class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">{{ historyLocked ? 'Inventory-affecting values are protected. Update the invoice reference or notes and save.' : activeStep === 1 ? 'Start with the supplier and payment details.' : activeStep === 2 ? 'Add every material line received in this purchase.' : activeStep === 3 ? 'Adjust the financial totals and add context.' : 'Review the draft before saving or posting it.' }}</p></div></div></aside>
      </div>
    </div>
  </div>
</template>
