<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { FileText, Plus, RotateCcw, Trash2, X } from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import { invoicesApi, type InvoiceRecord } from '../api/invoices'
import type { OrderRecord } from '../api/orders'
import { formatMoney, type CurrencyUnit } from '../utils/currency'
import { formatDateTime } from '../utils/date'
import { confirmAction, normalizeError } from '../ui/feedback'

const props = defineProps<{ currencyUnit: CurrencyUnit; orders: OrderRecord[] }>()
const emit = defineEmits<{ notify: [string]; refreshOrders: [] }>()
const rows = ref<InvoiceRecord[]>([])
const selected = ref<string | null>(null)
const query = ref('')
const status = ref('All')
const error = ref('')
const loading = ref(false)
const current = computed(() => rows.value.find((v) => v.id === selected.value) ?? null)
const filtered = computed(() => rows.value.filter((v) => (status.value === 'All' || v.status === status.value) && (!query.value.trim() || [v.invoiceNumber, v.customerName, v.orderId].join(' ').toLowerCase().includes(query.value.trim().toLowerCase()))))
function tone(value: string) { return value === 'Paid' || value === 'Posted' ? 'green' : value === 'Partially Paid' ? 'blue' : value === 'Voided' ? 'slate' : 'amber' }
async function load() { loading.value = true; try { rows.value = await invoicesApi.list(); if (!selected.value && rows.value[0]) selected.value = rows.value[0].id } catch (e) { error.value = normalizeError(e).message } finally { loading.value = false } }
onMounted(load)
async function create(orderId: string) { try { const value = await invoicesApi.createFromOrder(orderId); rows.value = [value, ...rows.value]; selected.value = value.id; emit('notify', 'Draft invoice created from the saved order snapshot.'); emit('refreshOrders') } catch (e) { error.value = normalizeError(e).message } }
async function post() { if (!current.value) return; try { const value = await invoicesApi.post(current.value.id); rows.value = rows.value.map((v) => v.id === value.id ? value : v); emit('notify', 'Invoice posted with AR, revenue, and eligible actual COGS.'); emit('refreshOrders') } catch (e) { error.value = normalizeError(e).message } }
async function reverse() { if (!current.value || !(await confirmAction({ title: 'Void invoice', message: 'Void this invoice with reversing journal entries?', confirmLabel: 'Void invoice', danger: true }))) return; try { const value = await invoicesApi.void(current.value.id); rows.value = rows.value.map((v) => v.id === value.id ? value : v); emit('notify', 'Invoice voided with history preserved.'); emit('refreshOrders') } catch (e) { error.value = normalizeError(e).message } }
async function remove() { if (!current.value || !(await confirmAction({ title: 'Delete draft invoice', message: 'Delete this draft invoice permanently?', confirmLabel: 'Delete invoice', danger: true }))) return; try { await invoicesApi.deleteDraft(current.value.id); rows.value = rows.value.filter((v) => v.id !== current.value!.id); selected.value = null; emit('notify', 'Draft invoice deleted.') } catch (e) { error.value = normalizeError(e).message } }
</script>
<template>
  <div>
    <WorkspaceStickyStack><header><div><p>Finance / receivables</p><h1>Invoices</h1><p>Invoice saved order snapshots once; derive receivables from allocations.</p></div><span>{{ rows.length }} invoices</span></header><section><label class="form-control gap-1"><span>Search invoices</span><input class="input input-bordered w-full min-w-0" v-model="query" type="search" placeholder="Search invoice, customer, or order" /></label><label class="form-control gap-1"><span>Status</span><select class="select select-bordered w-full min-w-0" v-model="status"><option>All</option><option>Draft</option><option>Posted</option><option>Partially Paid</option><option>Paid</option><option>Voided</option></select></label><span>{{ filtered.length }} shown</span></section></WorkspaceStickyStack>
    <div v-if="error" role="alert">{{ error }} <button class="btn btn-ghost" @click="error = ''" aria-label="Dismiss"><X :size="14" /></button></div>
    <section><SectionPanel title="Invoice register" subtitle="Invoice status and amounts are derived in Go from posted allocations."><div v-if="loading">Loading invoices…</div><div v-else-if="filtered.length"><table class="table table-zebra w-full"><thead><tr><th>Invoice</th><th>Customer</th><th>Order</th><th>Date</th><th>Total</th><th>Remaining</th><th>Status</th></tr></thead><tbody><tr v-for="value in filtered" :key="value.id" :class="{ 'bg-base-300': selected === value.id }" @click="selected = value.id"><td><span>{{ value.invoiceNumber }}</span></td><td>{{ value.customerName || 'Walk-in customer' }}</td><td>{{ value.orderId || '—' }}</td><td>{{ formatDateTime(value.issueDate) }}</td><td>{{ formatMoney(value.totalRial, props.currencyUnit) }}</td><td>{{ formatMoney(value.remainingRial, props.currencyUnit) }}</td><td><StatusBadge :label="value.status" :tone="tone(value.status)" /></td></tr></tbody></table></div><div v-else><FileText :size="22" /><strong>{{ rows.length ? 'No invoices match this filter' : 'No invoices yet' }}</strong></div></SectionPanel>
      <SectionPanel v-if="current" title="Invoice inspector" subtitle="Exact immutable line and accounting snapshots."><div><div><div><FileText :size="19" /></div><div><h3>{{ current.invoiceNumber }}</h3><p>{{ current.customerName || 'Walk-in customer' }} · {{ current.orderId || 'No order link' }}</p></div></div><StatusBadge :label="current.status" :tone="tone(current.status)" /><div><div v-for="line in current.items" :key="line.id"><span><strong>{{ line.description }}</strong><small>{{ line.quantity }} {{ line.quantityUnit }} · snapshot {{ formatMoney(line.unitPriceRial, props.currencyUnit) }}</small></span><strong>{{ formatMoney(line.lineTotalRial, props.currencyUnit) }}</strong></div></div><dl><div><dt>Subtotal</dt><dd>{{ formatMoney(current.subtotalRial, props.currencyUnit) }}</dd></div><div><dt>Paid</dt><dd>{{ formatMoney(current.paidRial, props.currencyUnit) }}</dd></div><div><dt>Remaining</dt><dd>{{ formatMoney(current.remainingRial, props.currencyUnit) }}</dd></div></dl><div><button class="btn btn-ghost" v-if="current.status === 'Draft'" @click="post"><Plus :size="15" /> Post invoice</button><button class="btn btn-ghost" v-if="current.status === 'Draft'" @click="remove"><Trash2 :size="15" /> Delete draft</button><button class="btn btn-ghost" v-if="['Posted','Partially Paid','Paid'].includes(current.status)" @click="reverse"><RotateCcw :size="15" /> Void / reverse</button></div></div></SectionPanel><SectionPanel v-else title="Invoice inspector" subtitle="Select an invoice or create one from a saved order."><div>No invoice selected.</div></SectionPanel></section>
    <SectionPanel title="Orders ready to invoice" subtitle="Creating an invoice copies stored order pricing snapshots exactly."><div><div v-for="order in props.orders.filter((value) => !value.invoiceId && value.totalRial > 0)" :key="order.id"><span><strong>{{ order.orderNumber }} · {{ order.customerName || 'Walk-in customer' }}</strong><small>{{ formatMoney(order.totalRial, props.currencyUnit) }} · {{ order.items.length }} items</small></span><button class="btn btn-ghost" @click="create(order.id)"><Plus :size="14" /> Create invoice</button></div><div v-if="!props.orders.some((value) => !value.invoiceId && value.totalRial > 0)">All priced orders already have invoices.</div></div></SectionPanel>
  </div>
</template>
