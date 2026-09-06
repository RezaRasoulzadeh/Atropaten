<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { FileText, Plus, RotateCcw } from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { invoicesApi, type InvoiceRecord } from '../api/invoices'
import { ordersApi, type OrderRecord } from '../api/orders'
import type { CurrencyUnit } from '../utils/currency'
import { formatMoney } from '../utils/currency'
const props=defineProps<{order:OrderRecord;currencyUnit:CurrencyUnit}>();const emit=defineEmits<{notify:[string];saved:[order:OrderRecord]}>();const invoice=ref<InvoiceRecord|null>(null);const error=ref('')
async function load(){try{if(props.order.invoiceId)invoice.value=await invoicesApi.get(props.order.invoiceId)}catch(e){error.value=String(e)}}onMounted(load)
async function create(){try{invoice.value=await invoicesApi.createFromOrder(props.order.id);emit('notify','Draft invoice created from this order.');emit('saved',await ordersApi.get(props.order.id))}catch(e){error.value=String(e)}}
async function post(){if(!invoice.value)return;try{invoice.value=await invoicesApi.post(invoice.value.id);emit('notify','Invoice posted.');emit('saved',await ordersApi.get(props.order.id))}catch(e){error.value=String(e)}}
async function reverse(){if(!invoice.value||!window.confirm('Void this invoice with a reversal?'))return;try{invoice.value=await invoicesApi.void(invoice.value.id);emit('notify','Invoice voided with history preserved.');emit('saved',await ordersApi.get(props.order.id))}catch(e){error.value=String(e)}}
</script>
<template><SectionPanel title="Invoice" subtitle="Commercial snapshot and receivable status"><div v-if="error" class="form-error">{{error}}</div><div v-if="invoice" class="invoice-order-panel"><div class="inspector-heading"><div class="material-inspector-icon"><FileText :size="19"/></div><div><h3>{{invoice.invoiceNumber}}</h3><p>{{invoice.items.length}} lines · {{formatMoney(invoice.totalRial,props.currencyUnit)}}</p></div><StatusBadge :label="invoice.status" :tone="invoice.status==='Paid'||invoice.status==='Posted'?'green':invoice.status==='Voided'?'slate':'amber'"/></div><div class="accounting-row"><span>Paid</span><strong>{{formatMoney(invoice.paidRial,props.currencyUnit)}}</strong></div><div class="accounting-row"><span>Remaining</span><strong>{{formatMoney(invoice.remainingRial,props.currencyUnit)}}</strong></div><div class="inspector-actions"><button v-if="invoice.status==='Draft'" class="button button-primary" @click="post"><Plus :size="15"/> Post invoice</button><button v-if="invoice.status==='Posted'||invoice.status==='Partially Paid'||invoice.status==='Paid'" class="button button-secondary" @click="reverse"><RotateCcw :size="15"/> Void</button></div></div><div v-else class="workspace-state"><p>No invoice linked to this order.</p><button class="button button-primary" @click="create"><Plus :size="15"/> Create invoice</button></div></SectionPanel></template>
