<script setup lang="ts">
import { onMounted, ref } from 'vue'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import type { CurrencyUnit } from '../utils/currency'
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../utils/currency'
import { formatDateTime } from '../utils/date'
import { accountingApi, type FinancialAccountRecord, type PaymentRecord } from '../api/accounting'
import { ordersApi } from '../api/orders'
import SelectField from '../components/SelectField.vue'
const props=defineProps<{order:any;currencyUnit:CurrencyUnit}>();const emit=defineEmits<{notify:[message:string];saved:[order:any]}>();const financial=ref<FinancialAccountRecord[]>([]);const payments=ref<PaymentRecord[]>([]);const amount=ref('');const account=ref('FIN-CASH');const method=ref('cash');const loading=ref(false)
const money=(n:number)=>formatMoney(n,props.currencyUnit)
async function load(){loading.value=true;try{financial.value=await accountingApi.financialAccounts();payments.value=(await accountingApi.payments()).filter(p=>p.allocations.some(a=>a.targetType==='order'&&a.targetId===props.order.id))}catch(e){emit('notify',String(e))}finally{loading.value=false}}
async function refreshOrder(){emit('saved',await ordersApi.get(props.order.id))}
async function post(){try{const n=parseMoneyInput(amount.value,props.currencyUnit);if(n===null||n<=0)throw new Error('Enter a positive amount');await accountingApi.createPayment({direction:'incoming',method:method.value,financialAccountId:account.value,customerId:props.order.customerId,amountRial:n,allocations:[{targetType:'order',targetId:props.order.id,amountRial:n}]});amount.value='';await load();await refreshOrder();emit('notify','Order payment posted')}catch(e){emit('notify',String(e))}}
async function reverse(p:PaymentRecord){try{await accountingApi.reversePayment(p.id);await load();await refreshOrder();emit('notify','Payment reversed')}catch(e){emit('notify',String(e))}}
onMounted(load)
</script>
<template><div><SectionPanel title="Payment status" subtitle="Paid and remaining amounts are derived by the backend."><div><span>Total<strong>{{ money(order.totalRial) }}</strong></span><span>Paid<strong>{{ money(order.paidRial||0) }}</strong></span><span>Remaining<strong>{{ money(order.remainingRial??order.totalRial) }}</strong></span></div><form @submit.prevent="post"><SelectField v-model="account" label="Financial account" :options="financial.map((value) => ({ label: value.name, value: value.id }))"/><SelectField v-model="method" label="Method" :options="[{ label: 'Cash', value: 'cash' }, { label: 'Bank transfer', value: 'bank_transfer' }, { label: 'Card', value: 'card' }, { label: 'Check', value: 'check' }]"/><input class="input input-bordered w-full min-w-0" v-model="amount" :placeholder="`Amount (${props.currencyUnit})`" inputmode="numeric" @blur="amount=formatMoneyInput(parseMoneyInput(amount,props.currencyUnit)||0,props.currencyUnit)"/><button class="btn btn-ghost" type="submit">Post payment</button></form></SectionPanel><SectionPanel title="Order payment history" subtitle="Reversals preserve the original payment and journal entry."><div v-if="!payments.length">No allocated payments for this order.</div><div v-for="p in payments" :key="p.id"><span><strong>{{ p.paymentNumber }}</strong><small>{{ money(p.amountRial) }} · {{ formatDateTime(p.postedAt) }}</small></span><span><StatusBadge :label="p.status" :tone="p.status==='posted'?'green':'slate'"/><button class="btn btn-ghost" v-if="p.status==='posted'" @click="reverse(p)">Reverse</button></span></div></SectionPanel></div></template>
