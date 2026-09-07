<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CheckCircle2, Factory, PackageOpen, Plus, RotateCcw, Save, Trash2, X } from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import SearchField from '../components/SearchField.vue'
import SelectField from '../components/SelectField.vue'
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
  <div>
    <WorkspaceStickyStack>
      <header><div><p>Operations / production ledger</p><h1>Production</h1><p>Reserve material, record actual usage, and keep job progress independent from commercial status.</p></div><button class="btn btn-ghost" type="button" @click="beginCreate"><Plus :size="16"/>New production job</button></header>
      <section class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_14rem_auto] sm:items-end" aria-label="Production filters"><SearchField v-model="searchQuery" label="Search production jobs" placeholder="Search job, order, customer, or service"/><SelectField v-model="statusFilter" label="Queue" aria-label="Production status" :options="['All', 'Pending', 'Ready', 'In Progress', 'Paused', 'Completed', 'Cancelled'].map((value) => ({ label: value, value }))"/><span class="self-end pb-2">{{ visibleJobs.length }} jobs</span></section>
    </WorkspaceStickyStack>
    <div v-if="error" role="alert"><span>{{ error }}</span><button class="btn btn-ghost" type="button" @click="error=''" aria-label="Dismiss"><X :size="15"/></button></div>
    <section>
      <SectionPanel title="Production queue" subtitle="Priority ordered persisted jobs.">
        <div v-if="loading">Loading production queue…</div>
        <div v-else-if="visibleJobs.length"><button v-for="job in visibleJobs" :key="job.id" type="button" :class="{'bg-base-300':selectedId===job.id}" @click="select(job.id)"><span><Factory :size="17"/></span><span><strong>{{ job.jobNumber }} · {{ job.serviceName }}</strong><small>{{ jobOrder(job) }} · {{ job.quantity }} {{ job.quantityUnit }} · {{ jobContext(job) }}</small></span><StatusBadge :label="job.status" :tone="statusTone(job.status)"/></button></div>
        <div v-else><PackageOpen :size="22"/><p>No production jobs yet.</p><button class="btn btn-ghost" type="button" @click="beginCreate"><Plus :size="15"/>Create job</button></div>
      </SectionPanel>

      <SectionPanel v-if="createMode" title="New production job" subtitle="The service and cost are copied from the immutable order item snapshot.">
        <form @submit.prevent="create"><SelectField v-model="form.orderId" label="Confirmed order" :options="[{ label: 'Select an order', value: '' }, ...props.orders.filter((order) => order.commercialStatus === 'Confirmed').map((order) => ({ label: `${order.orderNumber} · ${order.customerName || 'Walk-in'}`, value: order.id }))]" @update:model-value="chooseOrder"/><SelectField v-model="form.orderItemId" label="Order item" :options="[{ label: 'Select an item', value: '' }, ...(selectedOrder?.items ?? []).map((item) => ({ label: `${item.serviceName} · ${item.quantity} ${item.quantityUnit}`, value: item.id }))]"/><div><label class="form-control gap-1"><span>Quantity</span><input class="input input-bordered w-full min-w-0" v-model="form.quantity" inputmode="decimal"/></label><label class="form-control gap-1"><span>Unit</span><input class="input input-bordered w-full min-w-0" v-model="form.quantityUnit"/></label></div><div><SelectField v-model="form.priority" label="Priority" :options="['Urgent', 'High', 'Normal', 'Low'].map((value) => ({ label: value, value }))"/><SelectField v-model="form.assignedMachineId" label="Machine" :options="[{ label: 'Unassigned', value: '' }, ...props.machines.map((machine) => ({ label: machine.name, value: machine.id }))]"/></div><label class="form-control gap-1"><span>Notes</span><textarea class="textarea textarea-bordered w-full min-w-0" v-model="form.notes" rows="3"></textarea></label><div><button class="btn btn-ghost" type="button" @click="createMode=false">Cancel</button><button class="btn btn-ghost" type="submit" :disabled="saving"><Save :size="15"/>Create job</button></div></form>
      </SectionPanel>

      <SectionPanel v-else-if="selected" title="Job inspector" :subtitle="`${selected.serviceName} · ${jobOrder(selected)}`">
        <div><StatusBadge :label="selected.status" :tone="statusTone(selected.status)"/><span>{{ date(selected.createdAt) }}</span></div><div><Factory :size="19"/><div><h3>{{ selected.jobNumber }}</h3><p>{{ selected.quantity }} {{ selected.quantityUnit }} · {{ selected.priority }} priority</p></div></div>
        <dl><div><dt>Estimated cost</dt><dd>{{ formatMoney(selected.estimatedCostRial, props.currencyUnit) }}</dd></div><div><dt>Actual cost</dt><dd>{{ formatMoney(selected.actualTotalCostRial, props.currencyUnit) }}</dd></div><div><dt>Started</dt><dd>{{ date(selected.startedAt) }}</dd></div><div><dt>Completed</dt><dd>{{ date(selected.completedAt) }}</dd></div></dl>
        <div><button class="btn btn-ghost" v-if="selected.status==='Pending'" @click="changeStatus('Ready')">Mark ready</button><button class="btn btn-ghost" v-if="selected.status==='Pending'||selected.status==='Ready'||selected.status==='Paused'" @click="changeStatus('In Progress')">Start production</button><button class="btn btn-ghost" v-if="selected.status==='In Progress'" @click="changeStatus('Paused')">Pause</button><button class="btn btn-ghost" v-if="selected.status==='In Progress'" @click="changeStatus('Completed')"><CheckCircle2 :size="15"/>Complete</button><button class="btn btn-ghost" v-if="selected.status!=='Completed'&&selected.status!=='Cancelled'" @click="changeStatus('Cancelled')">Cancel</button></div>
        <div><h3>Reservations</h3><div v-for="r in reservations" :key="r.id"><span>{{ props.materials.find(m=>m.id===r.materialId)?.name??r.materialId }} · {{ r.quantity }}</span><button class="btn btn-ghost" v-if="r.status==='active'" @click="release(r)">Release</button><small v-else>{{ r.status }}</small></div><div><SelectField v-model="reservationMaterial" label="Material" :options="[{ label: 'Material', value: '' }, ...props.materials.map((material) => ({ label: `${material.name} · available ${material.availableStock}`, value: material.id }))]"/><input class="input input-bordered w-full min-w-0" v-model="reservationQuantity" placeholder="Qty" inputmode="decimal"/></div><button class="btn btn-ghost" @click="reserve"><RotateCcw :size="15"/>Reserve stock</button></div>
        <div><h3>Actual consumption & waste</h3><div><SelectField v-model="consumptionMaterial" label="Material" :options="[{ label: 'Material', value: '' }, ...props.materials.map((material) => ({ label: material.name, value: material.id }))]"/><input class="input input-bordered w-full min-w-0" v-model="consumedQuantity" placeholder="Consumed" inputmode="decimal"/><input class="input input-bordered w-full min-w-0" v-model="wasteQuantity" placeholder="Waste" inputmode="decimal"/></div><input class="input input-bordered w-full min-w-0" v-model="consumptionKey" placeholder="Idempotency key (optional)"/><button class="btn btn-ghost" @click="consume">Post immutable movement</button></div>
        <div><h3>Outsourced production</h3><div><SelectField v-model="outsourceSupplier" label="Supplier" :options="[{ label: 'Supplier', value: '' }, ...props.suppliers.map((supplier) => ({ label: supplier.name, value: supplier.id }))]"/><label class="form-control gap-1"><span>Quoted ({{ props.currencyUnit }})</span><input class="input input-bordered w-full min-w-0" v-model="outsourceQuotedCost" placeholder="0" inputmode="numeric"/></label><label class="form-control gap-1"><span>Actual ({{ props.currencyUnit }})</span><input class="input input-bordered w-full min-w-0" v-model="outsourceCost" placeholder="0" inputmode="numeric"/></label></div><input class="input input-bordered w-full min-w-0" v-model="outsourceDescription" placeholder="Outsourced scope / return details"/><button class="btn btn-ghost" @click="saveOutsource">Save outsourcing</button></div>
        <button class="btn btn-ghost" v-if="selected.status==='Pending'||selected.status==='Ready'" @click="remove"><Trash2 :size="14"/>Delete unposted draft</button>
      </SectionPanel>
      <SectionPanel v-else title="Job inspector" subtitle="Select a production job to inspect it."><PackageOpen :size="22"/><p>Production details will appear here.</p></SectionPanel>
    </section>
  </div>
</template>
