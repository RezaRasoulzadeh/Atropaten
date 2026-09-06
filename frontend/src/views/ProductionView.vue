<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CheckCircle2, Factory, PackageOpen, Plus, RotateCcw, Save, Trash2, X } from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import { productionApi, type ProductionJobRecord, type ReservationRecord } from '../api/production'
import type { OrderRecord } from '../api/orders'
import type { CurrencyUnit } from '../utils/currency'
import { formatMoney, parseMoneyInput } from '../utils/currency'
import { formatDateTime } from '../utils/date'

const props = defineProps<{ currencyUnit: CurrencyUnit; orders: OrderRecord[]; materials: any[]; machines: any[]; suppliers: any[] }>()
const emit = defineEmits<{ notify: [message:string] }>()
const jobs = ref<ProductionJobRecord[]>([])
const selectedId = ref<string|null>(null)
const statusFilter = ref('All')
const searchQuery = ref('')
const loading = ref(false)
const error = ref('')
const createMode = ref(false)
const saving = ref(false)
const form = ref({ orderId:'', orderItemId:'', quantity:'', quantityUnit:'', assignedMachineId:'', priority:'Normal', notes:'' })
const reservations = ref<ReservationRecord[]>([])
const reservationMaterial = ref('')
const reservationQuantity = ref('')
const consumedQuantity = ref('')
const wasteQuantity = ref('')
const consumptionMaterial = ref('')
const consumptionKey = ref('')
const consumptionNotes = ref('')
const outsourceSupplier = ref('')
const outsourceDescription = ref('')
const outsourceCost = ref('0')
const outsourceQuotedCost = ref('0')

const selected = computed(() => jobs.value.find(j => j.id === selectedId.value) ?? null)
const selectedOrder = computed(() => props.orders.find(o => o.id === (createMode.value ? form.value.orderId : selected.value?.orderId)) ?? null)
const selectedItem = computed(() => selectedOrder.value?.items.find(i => i.id === (createMode.value ? form.value.orderItemId : selected.value?.orderItemId)) ?? null)
const visibleJobs = computed(() => { const q=searchQuery.value.trim().toLowerCase(); return jobs.value.filter(j=>!q||[j.jobNumber,j.serviceName,jobOrder(j),props.orders.find(o=>o.id===j.orderId)?.customerName??''].some(v=>v.toLowerCase().includes(q))) })

onMounted(load)
async function load(){ loading.value=true; error.value=''; try { jobs.value=await productionApi.list(statusFilter.value); if(selectedId.value&&!jobs.value.some(j=>j.id===selectedId.value))selectedId.value=null; if(!selectedId.value&&jobs.value.length)select(jobs.value[0].id) } catch(e){error.value=message(e,'Production queue could not be loaded.')} finally{loading.value=false} }
async function select(id:string){ selectedId.value=id; createMode.value=false; try{reservations.value=await productionApi.reservations('',id,'')}catch(e){error.value=message(e,'Reservations could not be loaded.')} }
function beginCreate(){createMode.value=true;selectedId.value=null;reservations.value=[];form.value={orderId:'',orderItemId:'',quantity:'',quantityUnit:'',assignedMachineId:'',priority:'Normal',notes:''}}
function chooseOrder(){const o=props.orders.find(x=>x.id===form.value.orderId);const i=o?.items[0];form.value.orderItemId=i?.id??'';form.value.quantity=i?.quantity??'';form.value.quantityUnit=i?.quantityUnit??''}
async function create(){if(!form.value.orderId||!form.value.orderItemId||!form.value.quantity)return; saving.value=true;try{const j=await productionApi.create({...form.value,plannedAt:null});jobs.value.unshift(j);createMode.value=false;await select(j.id);emit('notify','Production job created from the saved order item.')}catch(e){error.value=message(e,'Production job could not be created.')}finally{saving.value=false}}
async function changeStatus(status:string){if(!selected.value)return;try{const j=await productionApi.status(selected.value.id,status);replace(j);await select(j.id);emit('notify',`Job moved to ${status}.`)}catch(e){error.value=message(e,'Production status could not be changed.')}}
async function remove(){if(!selected.value)return;try{await productionApi.delete(selected.value.id);jobs.value=jobs.value.filter(j=>j.id!==selected.value!.id);selectedId.value=null;reservations.value=[];emit('notify','Unposted production draft deleted.')}catch(e){error.value=message(e,'Production history cannot be deleted.')}}
async function reserve(){if(!selected.value||!reservationMaterial.value||!reservationQuantity.value)return;try{await productionApi.reserve({materialId:reservationMaterial.value,orderId:selected.value.orderId,orderItemId:selected.value.orderItemId,productionJobId:selected.value.id,quantity:reservationQuantity.value});reservationQuantity.value='';reservations.value=await productionApi.reservations('',selected.value.id,'');await load();emit('notify','Inventory reserved without creating a movement.')}catch(e){error.value=message(e,'Reservation exceeds available stock or is invalid.')}}
async function release(r:ReservationRecord){try{await productionApi.releaseReservation(r.id);if(selected.value)reservations.value=await productionApi.reservations('',selected.value.id,'')}catch(e){error.value=message(e,'Reservation could not be released.')}}
async function consume(){if(!selected.value||!consumptionMaterial.value||(!consumedQuantity.value&&!wasteQuantity.value))return;try{await productionApi.consume(selected.value.id,{materialId:consumptionMaterial.value,consumedQuantity:consumedQuantity.value||'0',wasteQuantity:wasteQuantity.value||'0',idempotencyKey:consumptionKey.value||`ui-${Date.now()}`,notes:consumptionNotes.value});consumedQuantity.value='';wasteQuantity.value='';consumptionKey.value='';consumptionNotes.value='';await load();if(selected.value)await select(selected.value.id);emit('notify','Consumption posted as immutable inventory movement(s).')}catch(e){error.value=message(e,'Consumption could not be posted.')}}
async function saveOutsource(){if(!selected.value)return;const cost=parseMoneyInput(outsourceCost.value,props.currencyUnit);const quoted=parseMoneyInput(outsourceQuotedCost.value,props.currencyUnit);if(cost===null||quoted===null){error.value='Enter whole outsourced costs.';return}try{const j=await productionApi.outsource(selected.value.id,{supplierId:outsourceSupplier.value,description:outsourceDescription.value,sentAt:'',expectedReturnAt:'',receivedAt:'',notes:'',quotedCostRial:quoted,actualCostRial:cost});replace(j);emit('notify','Outsourced production metadata saved.')}catch(e){error.value=message(e,'Outsourcing metadata could not be saved.')}}
function replace(j:ProductionJobRecord){const i=jobs.value.findIndex(x=>x.id===j.id);if(i>=0)jobs.value.splice(i,1,j)}
function jobOrder(j:ProductionJobRecord){return props.orders.find(o=>o.id===j.orderId)?.orderNumber??j.orderId}
function jobContext(j:ProductionJobRecord){const o=props.orders.find(x=>x.id===j.orderId);return `${o?.promisedAt?`Due ${date(o.promisedAt)}`:'No promised date'} · ${props.machines.find(m=>m.id===j.assignedMachineId)?.name??'Unassigned'}`}
function statusTone(s:string){return s==='Completed'?'green':s==='Cancelled'||s==='Failed'?'red':s==='In Progress'||s==='Ready'?'blue':s==='Paused'?'amber':'slate'}
function message(e:unknown,fallback:string){return e instanceof Error&&e.message?e.message:typeof e==='string'?e:fallback}
function date(v:string){try{return v?formatDateTime(v):'—'}catch{return '—'}}
</script>

<template>
  <div class="production-view">
    <WorkspaceStickyStack>
      <header class="workspace-heading production-heading"><div><p class="eyebrow">Operations / production ledger</p><h1>Production</h1><p class="heading-description">Reserve material, record actual usage, and keep job progress independent from commercial status.</p></div><button class="button button-primary" type="button" @click="beginCreate"><Plus class="button-icon" :size="16"/>New production job</button></header>
      <section class="production-filter-bar panel"><label class="production-search"><span class="sr-only">Search production jobs</span><input v-model="searchQuery" type="search" placeholder="Search job, order, customer, or service"/></label><span class="filter-label">Queue</span><span class="select-control"><select v-model="statusFilter" aria-label="Production status" @change="load"><option>All</option><option>Pending</option><option>Ready</option><option>In Progress</option><option>Paused</option><option>Completed</option><option>Cancelled</option></select></span><span class="filter-result">{{ visibleJobs.length }} jobs</span></section>
    </WorkspaceStickyStack>
    <div v-if="error" class="production-error" role="alert"><span>{{ error }}</span><button class="icon-button" type="button" @click="error=''" aria-label="Dismiss"><X :size="15"/></button></div>
    <section class="production-layout">
      <SectionPanel title="Production queue" subtitle="Priority ordered persisted jobs." class="production-queue-panel">
        <div v-if="loading" class="production-empty">Loading production queue…</div>
        <div v-else-if="visibleJobs.length" class="production-job-list"><button v-for="job in visibleJobs" :key="job.id" type="button" class="production-job-row" :class="{'is-selected':selectedId===job.id}" @click="select(job.id)"><span class="production-job-icon"><Factory :size="17"/></span><span class="production-job-copy"><strong>{{ job.jobNumber }} · {{ job.serviceName }}</strong><small>{{ jobOrder(job) }} · {{ job.quantity }} {{ job.quantityUnit }} · {{ jobContext(job) }}</small></span><StatusBadge :label="job.status" :tone="statusTone(job.status)"/></button></div>
        <div v-else class="production-empty"><PackageOpen :size="22"/><p>No production jobs yet.</p><button class="button button-secondary" type="button" @click="beginCreate"><Plus class="button-icon" :size="15"/>Create job</button></div>
      </SectionPanel>

      <SectionPanel v-if="createMode" title="New production job" subtitle="The service and cost are copied from the immutable order item snapshot." class="production-inspector">
        <form class="production-form" @submit.prevent="create"><label class="form-field form-field-wide"><span>Confirmed order</span><select v-model="form.orderId" @change="chooseOrder"><option value="">Select an order</option><option v-for="order in props.orders.filter(o=>o.commercialStatus==='Confirmed')" :key="order.id" :value="order.id">{{ order.orderNumber }} · {{ order.customerName||'Walk-in' }}</option></select></label><label class="form-field form-field-wide"><span>Order item</span><select v-model="form.orderItemId"><option value="">Select an item</option><option v-for="item in selectedOrder?.items??[]" :key="item.id" :value="item.id">{{ item.serviceName }} · {{ item.quantity }} {{ item.quantityUnit }}</option></select></label><div class="production-form-grid"><label class="form-field"><span>Quantity</span><input v-model="form.quantity" inputmode="decimal"/></label><label class="form-field"><span>Unit</span><input v-model="form.quantityUnit"/></label></div><div class="production-form-grid"><label class="form-field"><span>Priority</span><select v-model="form.priority"><option>Urgent</option><option>High</option><option>Normal</option><option>Low</option></select></label><label class="form-field"><span>Machine</span><select v-model="form.assignedMachineId"><option value="">Unassigned</option><option v-for="machine in props.machines" :key="machine.id" :value="machine.id">{{ machine.name }}</option></select></label></div><label class="form-field form-field-wide"><span>Notes</span><textarea v-model="form.notes" rows="3"></textarea></label><div class="production-actions"><button class="button button-secondary" type="button" @click="createMode=false">Cancel</button><button class="button button-primary" type="submit" :disabled="saving"><Save class="button-icon" :size="15"/>Create job</button></div></form>
      </SectionPanel>

      <SectionPanel v-else-if="selected" title="Job inspector" :subtitle="`${selected.serviceName} · ${jobOrder(selected)}`" class="production-inspector">
        <div class="inspector-status"><StatusBadge :label="selected.status" :tone="statusTone(selected.status)"/><span class="production-date">{{ date(selected.createdAt) }}</span></div><div class="production-inspector-title"><Factory :size="19"/><div><h3>{{ selected.jobNumber }}</h3><p>{{ selected.quantity }} {{ selected.quantityUnit }} · {{ selected.priority }} priority</p></div></div>
        <dl class="inspector-details"><div><dt>Estimated cost</dt><dd>{{ formatMoney(selected.estimatedCostRial, props.currencyUnit) }}</dd></div><div><dt>Actual cost</dt><dd>{{ formatMoney(selected.actualTotalCostRial, props.currencyUnit) }}</dd></div><div><dt>Started</dt><dd>{{ date(selected.startedAt) }}</dd></div><div><dt>Completed</dt><dd>{{ date(selected.completedAt) }}</dd></div></dl>
        <div class="production-status-actions"><button v-if="selected.status==='Pending'" class="button button-secondary" @click="changeStatus('Ready')">Mark ready</button><button v-if="selected.status==='Pending'||selected.status==='Ready'||selected.status==='Paused'" class="button button-primary" @click="changeStatus('In Progress')">Start production</button><button v-if="selected.status==='In Progress'" class="button button-secondary" @click="changeStatus('Paused')">Pause</button><button v-if="selected.status==='In Progress'" class="button button-primary" @click="changeStatus('Completed')"><CheckCircle2 class="button-icon" :size="15"/>Complete</button><button v-if="selected.status!=='Completed'&&selected.status!=='Cancelled'" class="button button-danger" @click="changeStatus('Cancelled')">Cancel</button></div>
        <div class="production-subsection"><h3>Reservations</h3><div v-for="r in reservations" :key="r.id" class="production-line"><span>{{ props.materials.find(m=>m.id===r.materialId)?.name??r.materialId }} · {{ r.quantity }}</span><button v-if="r.status==='active'" class="text-button" @click="release(r)">Release</button><small v-else>{{ r.status }}</small></div><div class="production-form-grid"><select v-model="reservationMaterial"><option value="">Material</option><option v-for="m in props.materials" :key="m.id" :value="m.id">{{ m.name }} · available {{ m.availableStock }}</option></select><input v-model="reservationQuantity" placeholder="Qty" inputmode="decimal"/></div><button class="button button-secondary" @click="reserve"><RotateCcw class="button-icon" :size="15"/>Reserve stock</button></div>
        <div class="production-subsection"><h3>Actual consumption & waste</h3><div class="production-form-grid"><select v-model="consumptionMaterial"><option value="">Material</option><option v-for="m in props.materials" :key="m.id" :value="m.id">{{ m.name }}</option></select><input v-model="consumedQuantity" placeholder="Consumed" inputmode="decimal"/><input v-model="wasteQuantity" placeholder="Waste" inputmode="decimal"/></div><input v-model="consumptionKey" class="production-wide-input" placeholder="Idempotency key (optional)"/><button class="button button-secondary" @click="consume">Post immutable movement</button></div>
        <div class="production-subsection"><h3>Outsourced production</h3><div class="production-form-grid"><select v-model="outsourceSupplier"><option value="">Supplier</option><option v-for="supplier in props.suppliers" :key="supplier.id" :value="supplier.id">{{ supplier.name }}</option></select><label class="form-field"><span>Quoted ({{ props.currencyUnit }})</span><input v-model="outsourceQuotedCost" placeholder="0" inputmode="numeric"/></label><label class="form-field"><span>Actual ({{ props.currencyUnit }})</span><input v-model="outsourceCost" placeholder="0" inputmode="numeric"/></label></div><input v-model="outsourceDescription" class="production-wide-input" placeholder="Outsourced scope / return details"/><button class="button button-secondary" @click="saveOutsource">Save outsourcing</button></div>
        <button v-if="selected.status==='Pending'||selected.status==='Ready'" class="text-button production-delete" @click="remove"><Trash2 :size="14"/>Delete unposted draft</button>
      </SectionPanel>
      <SectionPanel v-else title="Job inspector" subtitle="Select a production job to inspect it." class="production-inspector production-inspector-empty"><PackageOpen :size="22"/><p>Production details will appear here.</p></SectionPanel>
    </section>
  </div>
</template>

<style scoped>
.production-layout{display:grid;grid-template-columns:minmax(0,1fr) minmax(360px,460px);gap:18px;align-items:start}.production-filter-bar{display:flex;align-items:center;gap:10px;margin-top:16px}.production-search{min-width:180px;flex:1}.production-search input{width:100%;box-sizing:border-box}.production-heading{align-items:center}.production-job-list{display:grid;gap:4px}.production-job-row{min-width:0;display:flex;align-items:center;gap:12px;width:100%;padding:13px 12px;border:1px solid transparent;border-radius:10px;background:transparent;text-align:left;color:inherit;cursor:pointer}.production-job-row:hover,.production-job-row.is-selected{background:var(--surface-muted);border-color:var(--border)}.production-job-icon{display:grid;place-items:center;width:34px;height:34px;border-radius:9px;background:var(--accent-soft);color:var(--accent)}.production-job-copy{min-width:0;flex:1;display:grid;gap:3px}.production-job-copy strong,.production-job-copy small{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.production-job-copy small,.production-date{color:var(--text-muted);font-size:12px}.production-empty{min-height:260px;display:grid;place-items:center;align-content:center;gap:10px;color:var(--text-muted)}.production-form,.production-subsection{display:grid;gap:12px}.production-form-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px}.production-form-grid select,.production-form-grid input,.production-wide-input{min-width:0;width:100%;box-sizing:border-box}.production-actions,.production-status-actions{display:flex;flex-wrap:wrap;gap:8px}.production-inspector-title{display:flex;gap:10px;align-items:center;margin:16px 0;color:var(--accent)}.production-inspector-title h3{margin:0;color:var(--text)}.production-inspector-title p{margin:3px 0 0;color:var(--text-muted);font-size:12px}.production-subsection{border-top:1px solid var(--border);padding-top:16px;margin-top:18px}.production-subsection h3{margin:0;font-size:13px}.production-line{display:flex;justify-content:space-between;gap:8px;align-items:center;font-size:13px}.production-line small{color:var(--text-muted)}.production-delete{margin-top:18px;color:var(--danger)}.production-error{display:flex;justify-content:space-between;gap:10px;padding:10px 12px;margin:14px 0;background:var(--danger-soft);color:var(--danger);border-radius:8px}@media(max-width:1050px){.production-layout{grid-template-columns:minmax(0,1fr)}}
</style>
