<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import { accountingApi, type FinancialAccountRecord, type PaymentRecord } from '../../api/accounting'
import { checksApi, type CheckRecord } from '../../api/checks'
import { purchasesApi, type PurchaseRecord } from '../../api/purchases'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency'
import { currentCanonicalDate } from '../../utils/date'
import { reportError, useWorkspaceActions } from '../../composables/useWorkspaceActions'
import { confirmAction } from '../../ui/feedback'

const props = defineProps<{ purchase: PurchaseRecord; currencyUnit: CurrencyUnit; financial: FinancialAccountRecord[] }>()
const emit = defineEmits<{ saved: [purchase: PurchaseRecord] }>()
const { busy, runAction } = useWorkspaceActions()
const account = ref(props.purchase.financialAccountId || '')
const method = ref('cash')
const amount = ref('')
const checkNumber = ref('')
const dueDate = ref(currentCanonicalDate())
const payments = ref<PaymentRecord[]>([])
const checks = ref<CheckRecord[]>([])
const loading = ref(false)
const loaded = ref(false)
const request = ref<{ signature: string; id: string } | null>(null)
const selectedAccount = computed(() => props.financial.find(a => a.id === account.value && a.active))
const methods = computed(() => selectedAccount.value?.type === 'bank'
  ? [{ label: 'Bank transfer', value: 'bank_transfer' }, { label: 'Card', value: 'card' }, { label: 'Check', value: 'check' }]
  : [{ label: 'Cash', value: 'cash' }])
const pending = computed(() => checks.value.filter(c => ['Issued', 'Delivered'].includes(c.status)).reduce((sum, c) => sum + c.amountRial, 0))
const available = computed(() => Math.max(0, (props.purchase.remainingRial ?? props.purchase.totalRial) - pending.value))
const paymentSliderValue = computed(() => {
  if (!available.value) return 0
  const value = parseMoneyInput(amount.value, props.currencyUnit) || 0
  return Math.min(100, Math.max(0, value / available.value * 100))
})
function updatePaymentSlider(event: Event) {
  const percentage = Math.min(100, Math.max(0, Number((event.target as HTMLInputElement).value)))
  amount.value = formatMoneyInput(Math.round(available.value * percentage / 100), props.currencyUnit)
}
const money = (v: number) => formatMoney(v, props.currencyUnit)
watch(methods, options => { if (!options.some(o => o.value === method.value)) method.value = options[0]!.value }, { immediate: true })

async function load() {
  const id = props.purchase.id
  loading.value = true
  loaded.value = false
  try {
    const [paymentRows, checkRows] = await Promise.all([accountingApi.payments(), checksApi.list('outgoing')])
    if (id !== props.purchase.id) return
    payments.value = paymentRows.filter(p => p.allocations.some(a => a.targetType === 'purchase' && a.targetId === id))
    checks.value = checkRows.filter(c => c.sourceType === 'purchase' && c.sourceId === id)
    loaded.value = true
  } finally { if (id === props.purchase.id) loading.value = false }
}
watch(() => props.purchase.id, async () => {
  account.value = props.purchase.financialAccountId || ''
  amount.value = ''
  payments.value = []
  checks.value = []
  request.value = null
  try { await load() } catch (e) { reportError(e) }
}, { immediate: true })

async function refresh() {
  const id = props.purchase.id
  const updated = await purchasesApi.get(id)
  if (props.purchase.id !== id) return
  emit('saved', updated)
  await load()
}
function record() {
  return runAction(async () => {
    if (!loaded.value || props.purchase.status !== 'Posted') return
    const value = parseMoneyInput(amount.value, props.currencyUnit)
    if (!selectedAccount.value) throw new Error('Select an active payment account.')
    if (!value || value <= 0 || value > available.value) throw new Error('Enter an amount within the unpaid balance not covered by pending checks.')
    const signature = JSON.stringify([props.purchase.id, account.value, method.value, value, checkNumber.value, dueDate.value])
    if (request.value?.signature !== signature) request.value = { signature, id: crypto.randomUUID() }
    if (method.value === 'check') {
      if (!checkNumber.value.trim() || !dueDate.value) throw new Error('Enter the check number and due date.')
      await checksApi.create({ id: request.value.id, direction: 'outgoing', checkNumber: checkNumber.value.trim(), bank: selectedAccount.value.bankName || selectedAccount.value.name,
        accountDescriptor: selectedAccount.value.accountNumber, financialAccountId: account.value, supplierId: props.purchase.supplierId,
        payerPayee: props.purchase.supplierName, sourceType: 'purchase', sourceId: props.purchase.id, amountRial: value,
        issueDate: currentCanonicalDate(), dueDate: dueDate.value, status: 'Draft' })
    } else {
      await accountingApi.createPayment({ id: request.value.id, direction: 'outgoing', method: method.value, financialAccountId: account.value,
        supplierId: props.purchase.supplierId, amountRial: value, allocations: [{ targetType: 'purchase', targetId: props.purchase.id, amountRial: value }] })
    }
    // Clear only after the mutation succeeds; retries use the same ID.
    request.value = null
    amount.value = ''
    await refresh()
  })
}
function transition(check: CheckRecord, to: string) {
  return runAction(async () => {
    if (!(await confirmAction({ title: `${to} check`, message: to === 'Cleared' ? 'Confirm that the bank has cleared this check. This records the bank withdrawal.' : `Move check ${check.checkNumber} to ${to}?`, confirmLabel: 'Confirm' }))) return
    await checksApi.transition(check.id, to, '', `purchase-check:${check.id}:${to}`)
    await refresh()
  })
}
function reverse(payment: PaymentRecord) {
  return runAction(async () => {
    if (!(await confirmAction({ title: 'Reverse payment', message: 'Reverse this entire payment, including any allocations to other purchases?', confirmLabel: 'Reverse', danger: true }))) return
    await accountingApi.reversePayment(payment.id)
    await refresh()
  })
}
function allocated(payment: PaymentRecord) {
  return payment.allocations.filter(a => a.targetType === 'purchase' && a.targetId === props.purchase.id).reduce((sum, a) => sum + a.amountRial, 0)
}
</script>

<template>
  <section data-enter-scope class="space-y-3 rounded-box border border-base-300 p-4" aria-label="Purchase payments">
    <h3 class="text-sm font-semibold">Payment</h3>
    <dl class="grid grid-cols-2 gap-3 text-xs">
      <div><dt class="text-base-content/60">Paid (cash / cleared checks)</dt><dd class="mt-1 font-semibold text-success">{{ money(purchase.paidRial) }}</dd></div>
      <div><dt class="text-base-content/60">Remaining</dt><dd class="mt-1 font-semibold">{{ money(purchase.remainingRial) }}</dd></div>
      <div v-if="pending"><dt class="text-base-content/60">Pending checks</dt><dd class="mt-1 text-warning">{{ money(pending) }}</dd></div>
    </dl>
    <p v-if="purchase.status === 'Draft'" class="text-xs text-base-content/60">Not paid yet. Post the purchase, then record payment here now or later.</p>
    <template v-else-if="purchase.status === 'Posted' && available > 0">
      <div class="grid gap-3 sm:grid-cols-2">
        <FormField><span>Pay from</span><SelectField v-model="account" :disabled="busy" :options="[{ label: 'Select cash or bank account', value: '' }, ...financial.filter(a => a.active).map(a => ({ label: a.name, value: a.id }))]" /></FormField>
        <FormField><span>Method</span><SelectField v-model="method" :disabled="busy || !account" :options="methods" /></FormField>
        <FormField class="sm:col-span-2" :label="`Amount (${currencyUnit})`">
          <div class="relative min-w-0">
            <AppInput v-model="amount" :money="currencyUnit" class="pe-14" :disabled="busy" :placeholder="`Amount (${currencyUnit})`" inputmode="numeric" @blur="amount = formatMoneyInput(parseMoneyInput(amount, currencyUnit) || 0, currencyUnit)" />
            <button class="absolute inset-y-1 end-1 rounded px-2 text-xs font-medium text-primary transition-colors hover:bg-primary/10 disabled:cursor-not-allowed disabled:opacity-40" type="button" :disabled="busy || !loaded || !available" aria-label="Use maximum payment amount" @click="amount = formatMoneyInput(available, currencyUnit)">Max</button>
          </div>
          <div class="payment-slider mt-2 w-full overflow-visible">
            <input class="range range-primary range-sm w-full" type="range" min="0" max="100" step="5" :value="paymentSliderValue" :disabled="busy || !loaded || !available" aria-label="Payment amount percentage slider" :aria-valuetext="`${Math.round(paymentSliderValue)}% of available unpaid balance`" @input="updatePaymentSlider" />
            <div class="mt-1 grid w-full grid-cols-5 px-2.5 text-center text-[10px] leading-3 text-base-content/40" aria-hidden="true"><span v-for="tick in [0, 25, 50, 75, 100]" :key="tick">|</span></div>
            <div class="mt-1 grid w-full grid-cols-5 px-2.5 text-center text-xs leading-4 text-base-content/50" aria-hidden="true"><span v-for="tick in [0, 25, 50, 75, 100]" :key="tick">{{ tick }}</span></div>
          </div>
        </FormField>
        <template v-if="account && method === 'check'">
          <FormField><span>Check number</span><AppInput v-model="checkNumber" class="input w-full" :disabled="busy" /></FormField>
          <FormField><span>Due date</span><JalaliDatePicker v-model="dueDate" :disabled="busy" /></FormField>
        </template>
      </div>
      <button type="button" class="btn btn-primary btn-sm" :disabled="busy || loading || !loaded || !account" data-enter-submit @click="record">{{ method === 'check' ? 'Create check draft' : 'Record payment' }}</button>
      <p class="text-xs leading-5 text-base-content/60">Selecting an account does not pay the purchase. Cash/bank payments reduce the supplier payable when recorded. Delivered checks move it to checks payable; clearing records cash paid.</p>
    </template>
    <p v-if="loading" class="text-xs text-base-content/60">Loading payment history…</p>
    <div v-for="payment in payments" :key="payment.id" class="flex flex-wrap items-center justify-between gap-2 border-t border-base-300 pt-2 text-xs">
      <span>{{ payment.paymentNumber }} · {{ money(allocated(payment)) }} · {{ payment.status }}</span>
      <button v-if="payment.status === 'posted'" type="button" class="btn btn-ghost btn-xs" :disabled="busy" @click="reverse(payment)">Reverse</button>
    </div>
    <div v-for="check in checks" :key="check.id" class="space-y-2 border-t border-base-300 pt-2 text-xs">
      <p>Check {{ check.checkNumber }} · {{ money(check.amountRial) }} · {{ check.status }}</p>
      <div class="flex flex-wrap gap-2">
        <button v-if="check.status === 'Draft'" type="button" class="btn btn-outline btn-xs" :disabled="busy" @click="transition(check, 'Issued')">Issue</button>
        <button v-if="check.status === 'Issued'" type="button" class="btn btn-outline btn-xs" :disabled="busy" @click="transition(check, 'Delivered')">Mark delivered</button>
        <button v-if="check.status === 'Delivered'" type="button" class="btn btn-outline btn-xs" :disabled="busy" @click="transition(check, 'Cleared')">Mark cleared</button>
        <button v-if="['Draft', 'Issued'].includes(check.status)" type="button" class="btn btn-ghost btn-xs" :disabled="busy" @click="transition(check, 'Cancelled')">Cancel check</button>
        <button v-if="['Delivered', 'Cleared'].includes(check.status)" type="button" class="btn btn-ghost btn-xs" :disabled="busy" @click="transition(check, 'Returned')">Return check</button>
      </div>
    </div>
  </section>
</template>
