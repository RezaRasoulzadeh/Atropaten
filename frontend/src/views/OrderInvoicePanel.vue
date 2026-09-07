<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { FileText, Plus, RotateCcw } from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { invoicesApi, type InvoiceRecord } from '../api/invoices'
import { ordersApi, type OrderRecord } from '../api/orders'
import type { CurrencyUnit } from '../utils/currency'
import { formatMoney } from '../utils/currency'
import { confirmAction, normalizeError } from '../ui/feedback'
const props=defineProps<{order:OrderRecord;currencyUnit:CurrencyUnit}>();const emit=defineEmits<{notify:[string];saved:[order:OrderRecord]}>();const invoice=ref<InvoiceRecord|null>(null);const error=ref('')
async function load(){try{if(props.order.invoiceId)invoice.value=await invoicesApi.get(props.order.invoiceId)}catch(e){error.value=String(e)}}onMounted(load)
async function create(){try{invoice.value=await invoicesApi.createFromOrder(props.order.id);emit('notify','Draft invoice created from this order.');emit('saved',await ordersApi.get(props.order.id))}catch(e){error.value=String(e)}}
async function post(){if(!invoice.value)return;try{invoice.value=await invoicesApi.post(invoice.value.id);emit('notify','Invoice posted.');emit('saved',await ordersApi.get(props.order.id))}catch(e){error.value=String(e)}}
async function reverse(){if(!invoice.value||!(await confirmAction({title:'Void invoice',message:'Void this invoice with a reversal?',confirmLabel:'Void invoice',danger:true})))return;try{invoice.value=await invoicesApi.void(invoice.value.id);emit('notify','Invoice voided with history preserved.');emit('saved',await ordersApi.get(props.order.id))}catch(e){error.value=normalizeError(e).message}}
</script>
<template><SectionPanel title="Invoice" subtitle="Commercial snapshot and receivable status"><div v-if="error">{{error}}</div><div v-if="invoice"><div><div><FileText :size="19"/></div><div><h3>{{invoice.invoiceNumber}}</h3><p>{{invoice.items.length}} lines · {{formatMoney(invoice.totalRial,props.currencyUnit)}}</p></div><StatusBadge :label="invoice.status" :tone="invoice.status==='Paid'||invoice.status==='Posted'?'green':invoice.status==='Voided'?'slate':'amber'"/></div><div><span>Paid</span><strong>{{formatMoney(invoice.paidRial,props.currencyUnit)}}</strong></div><div><span>Remaining</span><strong>{{formatMoney(invoice.remainingRial,props.currencyUnit)}}</strong></div><div><button class="btn btn-ghost" v-if="invoice.status==='Draft'" @click="post"><Plus :size="15"/> Post invoice</button><button class="btn btn-ghost" v-if="invoice.status==='Posted'||invoice.status==='Partially Paid'||invoice.status==='Paid'" @click="reverse"><RotateCcw :size="15"/> Void</button></div></div><div v-else><p>No invoice linked to this order.</p><button class="btn btn-ghost" @click="create"><Plus :size="15"/> Create invoice</button></div></SectionPanel></template>
