<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Factory, PackageOpen } from 'lucide-vue-next'
import StatusBadge from '../components/StatusBadge.vue'
import { productionApi, type ConsumptionRecord, type ProductionJobRecord, type ReservationRecord } from '../api/production'
import type { OrderRecord } from '../api/orders'
import type { CurrencyUnit } from '../utils/currency'
import { formatMoney } from '../utils/currency'
import { formatDateTime } from '../utils/date'
const props=defineProps<{order:OrderRecord;currencyUnit:CurrencyUnit}>()
const jobs=ref<ProductionJobRecord[]>([]);const error=ref('');const reserved=ref<Record<string,ReservationRecord[]>>({});const usage=ref<Record<string,ConsumptionRecord[]>>({})
const progress=computed(()=>jobs.value.length?Math.round(jobs.value.filter(j=>j.status==='Completed').length/jobs.value.length*100):0)
onMounted(async()=>{try{jobs.value=(await productionApi.list('All')).filter(j=>j.orderId===props.order.id);await Promise.all(jobs.value.map(async j=>{reserved.value[j.id]=await productionApi.reservations('',j.id,'');usage.value[j.id]=await productionApi.consumptions(j.id)}))}catch(e){error.value=String(e)}})
function reservedCount(id:string){return (reserved.value[id]??[]).filter(r=>r.status==='active').length}
function usageCount(id:string){const rows=usage.value[id]??[];return rows.reduce((n,r)=>n+Number(r.consumedQuantity)+Number(r.wasteQuantity),0)}
function tone(s:string){return s==='Completed'?'green':s==='Cancelled'||s==='Failed'?'red':s==='In Progress'||s==='Ready'?'blue':s==='Paused'?'amber':'slate'}
</script>
<template><section><header><div><h2>Production jobs</h2><p>{{ jobs.length }} jobs · {{ progress }}% completed · independent from payment</p></div><Factory :size="19" aria-hidden="true"/></header><p v-if="error">{{ error }}</p><div v-else-if="jobs.length"><div v-for="job in jobs" :key="job.id"><span><Factory :size="16"/></span><div><strong>{{ job.jobNumber }} · {{ job.serviceName }}</strong><small>{{ job.quantity }} {{ job.quantityUnit }} · {{ job.createdAt?formatDateTime(job.createdAt):'—' }} · {{ reservedCount(job.id) }} active reservations · {{ usageCount(job.id) }} used / waste</small></div><StatusBadge :label="job.status" :tone="tone(job.status)"/><span>{{ formatMoney(job.actualTotalCostRial,props.currencyUnit) }}</span></div></div><div v-else><PackageOpen :size="21"/><strong>No production jobs linked yet</strong><span>Confirm the order, then create a job from the Production workspace.</span></div></section>
</template>
