<script setup lang="ts">
import { computed, onMounted, ref, watch, type Component } from 'vue'
import { ArrowRight, BarChart3, CircleAlert, HandCoins, Package, Printer, ReceiptText, RefreshCw, ShoppingCart, TriangleAlert, TrendingUp } from 'lucide-vue-next'
import KpiCard from '../components/KpiCard.vue'
import SectionPanel from '../components/SectionPanel.vue'
import WorkspaceBottomActions from '../components/WorkspaceBottomActions.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { reportsApi, type DashboardRecord } from '../api/reports'
import { formatMoney } from '../utils/currency'
import { formatDate, formatDateTime, currentCanonicalDate } from '../utils/date'
import { normalizeError } from '../ui/feedback'
const props = defineProps<{ currencyUnit: 'Rial' | 'Toman' }>(); const emit = defineEmits<{ navigate: [view: string]; newOrder: []; notify: [message: string] }>(); const data = ref<DashboardRecord | null>(null); const loading = ref(false); const error = ref(''); const end = currentCanonicalDate(); const start = new Date(Date.now() - 30 * 86400000).toISOString().slice(0, 10); const money = (v: number) => formatMoney(v, props.currencyUnit)
async function load() { loading.value = true; try { data.value = await reportsApi.dashboard(start, end) } catch (e) { error.value = normalizeError(e).message } finally { loading.value = false } }; onMounted(load); watch(() => props.currencyUnit, load); const attention = computed(() => data.value?.attention ?? [])
</script>
<template>
    <div>
        <WorkspaceStickyStack>
            <header>
                <div>
                    <p>{{ data?.startDate || start }} → {{ data?.endDate || end }}</p>
                    <h1>Good morning</h1>
                    <p>Here is what needs your attention today.</p>
                </div>
                <div><button class="btn btn-outline btn-primary" type="button" @click="load">
                        <RefreshCw :size="15" /><span>Refresh</span>
                    </button><button class="btn btn-primary" type="button" @click="emit('newOrder')"><span>New order</span></button>
                </div>
            </header>
        </WorkspaceStickyStack>
        <p v-if="error">{{ error }}</p>
        <div>
            <section>
                <KpiCard :value="money(data?.revenueRial || 0)" detail="Posted invoices in period" title="Sales"
                    :trend="loading ? 'Loading' : 'Persisted'" :icon="TrendingUp" accent="blue" />
                <KpiCard :value="money(data?.grossProfitRial || 0)" detail="Journal-derived P&L" title="Gross profit"
                    trend="Journal" :icon="BarChart3" accent="green" />
                <KpiCard :value="money(data?.receivableRial || 0)" :detail="`${data?.openInvoiceCount || 0} open invoices`"
                    title="Receivables" trend="Authoritative" :icon="HandCoins" accent="amber" />
                <KpiCard :value="money(data?.payableRial || 0)" detail="Journal-derived" title="Payables" trend="Supplier"
                    :icon="ReceiptText" accent="red" />
            </section>
            <section>
                <SectionPanel title="Needs attention" subtitle="Due obligations and operational exceptions">
                    <div><button class="btn btn-ghost" v-for="item in attention" :key="`${item.kind}-${item.detail}`"
                            type="button" @click="emit('notify', item.detail)">
                            <CircleAlert :size="15" /><span><span>{{ item.label }}</span><span>{{ item.detail }}<template
                                        v-if="item.date"> ·
                                        {{ formatDate(item.date) }}</template></span></span><span>{{ item.amountRial ? money(item.amountRial) : '—' }}</span>
                        </button>
                        <p v-if="!attention.length">No attention items from persisted data.</p>
                    </div>
                </SectionPanel>
                <SectionPanel title="Production queue" subtitle="Persisted active jobs">
                    <div>
                        <table class="table table-zebra w-full">
                            <thead>
                                <tr>
                                    <th>Order</th>
                                    <th>Customer</th>
                                    <th>Service</th>
                                    <th>Status</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="job in data?.production" :key="job.id">
                                    <td>{{ job.orderNumber || job.id }}</td>
                                    <td>{{ job.customer || '—' }}</td>
                                    <td>{{ job.service }}</td>
                                    <td>
                                        <StatusBadge :label="job.status" tone="blue" />
                                    </td>
                                </tr>
                                <tr v-if="!data?.production?.length">
                                    <td colspan="4">No active production jobs.</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </SectionPanel>
            </section>
            <section>
                <SectionPanel title="Low stock" subtitle="Movement-derived availability">
                    <div>
                        <div v-for="item in data?.lowStock" :key="item.id">
                            <TriangleAlert :size="15" /><span>{{ item.name }}</span><span>{{ item.availableUnits / 1000000 }}
                                {{ item.unit }}</span><span>Reorder at {{ item.reorderLevelUnits / 1000000 }}</span>
                        </div>
                        <p v-if="!data?.lowStock?.length">No low-stock materials.</p>
                    </div>
                </SectionPanel>
                <SectionPanel title="Recent payments" subtitle="Latest persisted financial activity">
                    <div>
                        <div v-for="item in data?.recentActivity" :key="item.id"><span>
                                <HandCoins :size="15" />
                            </span><span><span>{{ item.label }}</span><span>{{ item.detail }} ·
                                    {{ formatDateTime(item.date) }}</span></span><span
                                :class="{ 'value-positive': item.direction === 'incoming' }">{{ money(item.amountRial) }}</span>
                        </div>
                        <p v-if="!data?.recentActivity?.length">No payments recorded.</p>
                    </div>
                </SectionPanel>
            </section>
        </div>
        <WorkspaceBottomActions>
            <div>
                <p>Quick actions</p>
                <p>Jump into real workflows.</p>
            </div>
            <div><button class="btn btn-primary" type="button" @click="emit('newOrder')"><span>New order</span></button><button
                    class="btn btn-ghost" type="button" @click="emit('navigate', 'Production')">
                    <Printer :size="16" /><span>Production</span>
                </button><button class="btn btn-ghost" type="button" @click="emit('navigate', 'Purchases')">
                    <ShoppingCart :size="16" /><span>Purchase</span>
                </button><button class="btn btn-ghost" type="button" @click="emit('navigate', 'Reports')">
                    <BarChart3 :size="16" /><span>Reports</span>
                </button></div>
        </WorkspaceBottomActions>
    </div>
</template>
