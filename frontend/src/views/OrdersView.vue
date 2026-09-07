<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus, SearchX } from 'lucide-vue-next'
import StatusBadge from '../components/StatusBadge.vue'
import SearchFilterBar from '../components/SearchFilterBar.vue'
import SearchField from '../components/SearchField.vue'
import SelectField from '../components/SelectField.vue'
import SectionPanel from '../components/SectionPanel.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import WorkspaceHeader from '../components/WorkspaceHeader.vue'
import type { OrderRecord } from '../api/orders'
import { formatMoney, type CurrencyUnit } from '../utils/currency'
import { formatDateTime } from '../utils/date'
type Tone='blue'|'green'|'amber'|'red'|'slate'
const props=defineProps<{orders:OrderRecord[];currencyUnit:CurrencyUnit;loading?:boolean;error?:string}>()
const emit=defineEmits<{ 'open-order':[id:string];'new-order':[] }>()
const query=ref('');const commercial=ref('All');const fulfillment=ref('All');const payment=ref('All');const priority=ref('All');const selected=ref<string|null>(null)
const commercialOptions=[{label:'All',value:'All'},{label:'Draft',value:'Draft'},{label:'Confirmed',value:'Confirmed'},{label:'Closed',value:'Closed'},{label:'Cancelled',value:'Cancelled'}]
const fulfillmentOptions=[{label:'All',value:'All'},{label:'Pending',value:'Pending'},{label:'In Production',value:'In Production'},{label:'Ready',value:'Ready'},{label:'Delivered',value:'Delivered'}]
const paymentOptions=[{label:'All',value:'All'},{label:'Unpaid',value:'Unpaid'},{label:'Partially Paid',value:'Partially Paid'},{label:'Paid',value:'Paid'}]
const priorityOptions=[{label:'All',value:'All'},{label:'Urgent',value:'Urgent'},{label:'High',value:'High'},{label:'Normal',value:'Normal'},{label:'Low',value:'Low'}]
const orderItems=(order:OrderRecord)=>Array.isArray(order.items)?order.items:[]
const filtered=computed(()=>props.orders.filter(o=>{const q=query.value.trim().toLowerCase();return(!q||[o.orderNumber,o.customerName,...orderItems(o).map(i=>i.serviceName)].some(v=>String(v??'').toLowerCase().includes(q)))&&(commercial.value==='All'||o.commercialStatus===commercial.value)&&(fulfillment.value==='All'||o.fulfillmentStatus===fulfillment.value)&&(payment.value==='All'||o.paymentStatus===payment.value)&&(priority.value==='All'||o.priority===priority.value)}))
function money(v:number){return formatMoney(v,props.currencyUnit)}
function tone(v:string):Tone{return v==='Confirmed'||v==='In Production'?'blue':v==='Closed'||v==='Delivered'||v==='Paid'||v==='Ready'?'green':v==='Cancelled'?'red':v==='Partially Paid'?'amber':'slate'}
function itemSummary(o:OrderRecord){return orderItems(o).map(i=>i.serviceName).join(' · ')||'No configured items'}
function clear(){query.value='';commercial.value=fulfillment.value=payment.value=priority.value='All'}
</script>
<template>
  <div class="space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader eyebrow="Sales / operational queue" title="Orders" description="Track persisted customer orders and their independent state axes.">
        <button class="btn btn-primary gap-2" type="button" @click="emit('new-order')">
          <Plus :size="16" :stroke-width="1.8" aria-hidden="true" />
          <span>New order</span>
        </button>
      </WorkspaceHeader>

      <SearchFilterBar>
        <template #search>
          <SearchField v-model="query" label="Search orders" placeholder="Order, customer, or service" />
        </template>
        <template #filters>
          <SelectField class="w-36" v-model="commercial" label="Commercial" :options="commercialOptions" />
          <SelectField class="w-36" v-model="fulfillment" label="Fulfillment" :options="fulfillmentOptions" />
          <SelectField class="w-32" v-model="payment" label="Payment" :options="paymentOptions" />
          <SelectField class="w-32" v-model="priority" label="Priority" :options="priorityOptions" />
        </template>
        <template #count><span>{{ filtered.length }} of {{ props.orders.length }} orders</span></template>
        <template #actions>
          <button v-if="query || commercial !== 'All' || fulfillment !== 'All' || payment !== 'All' || priority !== 'All'" class="btn btn-ghost btn-sm" type="button" @click="clear">Clear</button>
        </template>
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <SectionPanel title="All orders" subtitle="Open an order to inspect its accepted pricing snapshots." :flush="true">
      <template #action>
        <span class="text-xs text-base-content/60">Independent state axes</span>
      </template>

      <div v-if="loading" class="p-6 text-sm text-base-content/60">Loading orders…</div>
      <div v-else-if="error" class="p-6 text-sm text-error">{{ error }}</div>
      <div v-else-if="!filtered.length" class="flex min-h-40 flex-col items-center justify-center gap-3 p-6 text-center">
        <SearchX class="text-base-content/50" :size="24" aria-hidden="true" />
        <strong>{{ props.orders.length ? 'No orders match these filters' : 'No persisted orders yet' }}</strong>
        <button v-if="props.orders.length" class="btn btn-ghost btn-sm" type="button" @click="clear">Clear filters</button>
      </div>
      <div v-else>
        <div class="divide-y divide-base-300">
          <article
            v-for="order in filtered"
            :key="order.id"
            class="min-w-0 cursor-pointer p-3 transition-colors hover:bg-base-200"
            :class="{ 'bg-base-300': selected === order.id }"
            tabindex="0"
            @click="selected = order.id"
            @dblclick="emit('open-order', order.id)"
            @keydown.enter="emit('open-order', order.id)"
          >
            <div class="flex min-w-0 items-center justify-between gap-3">
              <div class="flex min-w-0 items-center gap-2">
                <button class="btn btn-ghost btn-sm -ms-2 h-8 min-h-8 shrink-0 px-2 font-semibold" type="button" @click.stop="emit('open-order', order.id)">
                  {{ order.orderNumber }}
                </button>
                <p class="truncate text-sm font-semibold">{{ order.customerName || 'Walk-in customer' }}</p>
              </div>
              <strong class="shrink-0 text-sm text-primary">{{ money(order.totalRial) }}</strong>
            </div>

            <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2 lg:grid-cols-4">
              <div class="min-w-0">
                <span class="block text-base-content/50">Items</span>
                <span class="block break-words text-base-content/80">{{ itemSummary(order) }}</span>
                <span class="block text-base-content/50">{{ orderItems(order).length }} line items</span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Contact</span>
                <span class="block truncate text-base-content/80">{{ order.customerPhone || 'No contact details' }}</span>
              </div>
              <div>
                <span class="block text-base-content/50">Created</span>
                <span class="block text-base-content/80">{{ formatDateTime(order.createdAt) }}</span>
              </div>
              <div>
                <span class="block text-base-content/50">Promised</span>
                <span class="block text-base-content/80">{{ order.promisedAt ? formatDateTime(order.promisedAt) : '—' }}</span>
              </div>
            </div>

            <div class="mt-2 flex flex-wrap gap-1.5">
              <StatusBadge :label="order.commercialStatus" :tone="tone(order.commercialStatus)" />
              <StatusBadge :label="order.fulfillmentStatus" :tone="tone(order.fulfillmentStatus)" />
              <StatusBadge :label="order.paymentStatus" :tone="tone(order.paymentStatus)" />
              <StatusBadge :label="order.priority" :tone="tone(order.priority)" />
            </div>
          </article>
        </div>
      </div>
    </SectionPanel>
  </div>
</template>
