<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import {
  ArrowDownLeft,
  ArrowRight,
  BarChart3,
  BriefcaseBusiness,
  Calculator,
  ChartNoAxesCombined,
  ChevronDown,
  ChevronRight,
  CircleAlert,
  CircleDollarSign,
  ClipboardList,
  HandCoins,
  Factory,
  Info,
  Landmark,
  LayoutDashboard,
  MoreHorizontal,
  Package,
  PackageOpen,
  Plus,
  Printer,
  ReceiptText,
  FileText,
  Settings,
  ShoppingCart,
  Sparkles,
  TriangleAlert,
  Truck,
  TrendingUp,
  UserRound,
  Users,
} from 'lucide-vue-next'
import AppToolbar from './components/AppToolbar.vue'
import KpiCard from './components/KpiCard.vue'
import SectionPanel from './components/SectionPanel.vue'
import SidebarNavItem from './components/SidebarNavItem.vue'
import StatusBadge from './components/StatusBadge.vue'
import WorkspaceBottomActions from './components/WorkspaceBottomActions.vue'
import WorkspaceStickyStack from './components/WorkspaceStickyStack.vue'
import { customersApi } from './api/customers'
import { ordersApi, type OrderRecord } from './api/orders'
import { servicesApi } from './api/services'
import { materialsApi } from './api/materials'
import { machinesApi } from './api/machines'
import { formatMoney, formatSignedMoney, type CurrencyUnit } from './utils/currency'
import { formatDate, formatDateTime } from './utils/date'
import OrderWorkspaceView from './views/OrderWorkspaceView.vue'
import OrdersView from './views/OrdersView.vue'
import MaterialsView from './views/MaterialsView.vue'
import ServicesView from './views/ServicesView.vue'
import MachinesView from './views/MachinesView.vue'
import CustomersView from './views/CustomersView.vue'
import QuotesView from './views/QuotesView.vue'
import QuoteWorkspaceView from './views/QuoteWorkspaceView.vue'
import { quotesApi, type QuoteRecord } from './api/quotes'
import SuppliersView from './views/SuppliersView.vue'
import PurchasesView from './views/PurchasesView.vue'
import ProductionView from './views/ProductionView.vue'
import { suppliersApi, type SupplierRecord } from './api/suppliers'
import { purchasesApi, type PurchaseRecord } from './api/purchases'
import AccountingView from './views/AccountingView.vue'
import InvoicesView from './views/InvoicesView.vue'
import ChecksView from './views/ChecksView.vue'
import LoansView from './views/LoansView.vue'

type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate'

interface NavigationItem {
  label: string
  icon: Component
}

const navigationSections: { label: string; items: NavigationItem[] }[] = [
  {
    label: 'Workspace',
    items: [
      { label: 'Dashboard', icon: LayoutDashboard },
      { label: 'Orders', icon: ClipboardList },
      { label: 'Quotes', icon: FileText },
      { label: 'Production', icon: Printer },
      { label: 'Customers', icon: Users },
    ],
  },
  {
    label: 'Catalog & purchasing',
    items: [
      { label: 'Services', icon: BriefcaseBusiness },
      { label: 'Materials', icon: Package },
      { label: 'Machines', icon: Factory },
      { label: 'Purchases', icon: ShoppingCart },
      { label: 'Suppliers', icon: Truck },
    ],
  },
  {
    label: 'Finance',
    items: [
      { label: 'Accounting', icon: Calculator },
      { label: 'Invoices', icon: ReceiptText },
      { label: 'Checks', icon: Landmark },
      { label: 'Loans', icon: CircleDollarSign },
      { label: 'Owners', icon: UserRound },
    ],
  },
  {
    label: 'Insights & setup',
    items: [
      { label: 'Reports', icon: BarChart3 },
      { label: 'Settings', icon: Settings },
    ],
  },
]

const viewDescriptions: Record<string, { eyebrow: string; title: string; description: string }> = {
  Dashboard: {
    eyebrow: '2024-08-12',
    title: 'Good morning, Reza',
    description: 'Here is what needs your attention today.',
  },
}

const activeView = ref('Dashboard')
const isSidebarCollapsed = ref(false)
const searchQuery = ref('')
const selectedPeriod = ref('This week')
const currencyUnit = ref<CurrencyUnit>('Toman')
const selectedOrderId = ref<string | null>(null)
const selectedQuoteId = ref<string | null>(null)
const orders = ref<OrderRecord[]>([])
const quotes = ref<QuoteRecord[]>([])
const suppliers = ref<SupplierRecord[]>([])
const purchases = ref<PurchaseRecord[]>([])
const ordersLoading = ref(false)
const ordersError = ref('')
const quotesLoading = ref(false)
const quotesError = ref('')
const customers = ref<any[]>([])
const catalogServices = ref<any[]>([])
const catalogMaterials = ref<any[]>([])
const catalogMachines = ref<any[]>([])
const toastMessage = ref('')
let toastTimer: ReturnType<typeof setTimeout> | undefined

const currentView = computed(() => viewDescriptions[activeView.value] ?? {
  eyebrow: 'Atropaten workspace',
  title: activeView.value,
  description: 'This workspace is ready for the next product milestone.',
})

const productionJobs = [
  { id: 'ORD-1048', customer: 'Mehr Studio', item: 'Business cards · 500 pcs', promisedAt: '2024-08-12T15:30:00+03:30', status: 'In progress', tone: 'blue' as Tone, progress: 72 },
  { id: 'ORD-1045', customer: 'Arman Foods', item: 'Menu · A3 laminated', promisedAt: '2024-08-12T17:00:00+03:30', status: 'Ready to print', tone: 'green' as Tone, progress: 100 },
  { id: 'ORD-1042', customer: 'Nika Events', item: 'Event banners · 3 pcs', promisedAt: '2024-08-13T10:00:00+03:30', status: 'Queued', tone: 'slate' as Tone, progress: 18 },
  { id: 'ORD-1039', customer: 'Pendar Clinic', item: 'Appointment cards · 1,000 pcs', promisedAt: '2024-08-12T13:30:00+03:30', status: 'Waiting for proof', tone: 'amber' as Tone, progress: 45 },
]

const financialAlerts = [
  { label: 'Check due today', detail: 'Pars Paper · #CH-2094', detailDate: undefined, amountRial: 128_000_000, tone: 'red' as Tone, icon: CircleAlert },
  { label: 'Installment due this week', detail: 'Digital press loan', detailDate: '2024-08-16', amountRial: 86_000_000, tone: 'amber' as Tone, icon: Landmark },
  { label: 'Invoice overdue', detail: 'Pendar Clinic · 12 days', detailDate: undefined, amountRial: 245_000_000, tone: 'red' as Tone, icon: ReceiptText },
]

const lowStockItems = [
  { name: 'A4 300gsm matte', available: '3 reams', threshold: '5 reams', tone: 'red' as Tone, icon: TriangleAlert },
  { name: 'Gloss lamination roll', available: '42 m', threshold: '50 m', tone: 'amber' as Tone, icon: TriangleAlert },
  { name: 'Black toner · Ricoh', available: '18%', threshold: '25%', tone: 'amber' as Tone, icon: TriangleAlert },
]

const recentTransactions = [
  { title: 'Payment received', party: 'Mehr Studio · ORD-1048', amountRial: 92_000_000, direction: 'in' as const, time: '10:42 AM', tone: 'green' as Tone, icon: ArrowDownLeft },
  { title: 'Purchase recorded', party: 'Pars Paper · PUR-0061', amountRial: 128_000_000, direction: 'out' as const, time: '09:18 AM', tone: 'slate' as Tone, icon: ShoppingCart },
  { title: 'Expense posted', party: 'Courier and delivery', amountRial: 8_400_000, direction: 'out' as const, time: 'Yesterday', tone: 'slate' as Tone, icon: ReceiptText },
  { title: 'Payment received', party: 'Arman Foods · ORD-1045', amountRial: 64_000_000, direction: 'in' as const, time: 'Yesterday', tone: 'green' as Tone, icon: ArrowDownLeft },
]

function selectView(label: string) {
  activeView.value = label
  selectedOrderId.value = null
  selectedQuoteId.value = null
  if (label === 'Orders') { loadOrders(); loadOrderCatalog() }
  if (label === 'Quotes') { loadQuotes(); loadOrderCatalog() }
  if (label === 'Suppliers') loadSuppliers()
  if (label === 'Purchases') { loadSuppliers(); loadPurchases(); loadOrderCatalog() }
}

const selectedOrder = computed(() => orders.value.find((order) => order.id === selectedOrderId.value) ?? null)
const selectedQuote = computed(() => quotes.value.find((quote) => quote.id === selectedQuoteId.value) ?? null)

async function loadOrders() {
  ordersLoading.value = true
  ordersError.value = ''
  try { orders.value = await ordersApi.list() } catch (error) { ordersError.value = String(error) } finally { ordersLoading.value = false }
}

async function loadOrderCatalog() {
  try {
    customers.value = await customersApi.list(true)
    catalogServices.value = await servicesApi.list(true)
    catalogMaterials.value = await materialsApi.list(true)
    catalogMachines.value = await machinesApi.list(true)
  } catch (error) { showToast(String(error)) }
}
async function loadQuotes() { quotesLoading.value=true; quotesError.value=''; try { quotes.value=await quotesApi.list() } catch(error) { quotesError.value=String(error) } finally { quotesLoading.value=false } }
async function loadSuppliers() { try { suppliers.value = await suppliersApi.list(true) } catch (error) { showToast(String(error)) } }
async function loadPurchases() { try { purchases.value = await purchasesApi.list() } catch (error) { showToast(String(error)) } }

function openOrder(orderId: string) {
  activeView.value = 'Orders'
  selectedOrderId.value = orderId
}
function openQuote(quoteId: string) { activeView.value='Quotes'; selectedQuoteId.value=quoteId }

async function openNewOrder() {
  activeView.value = 'Orders'
  try {
    const order = await ordersApi.create({ customerId: '', promisedAt: null, priority: 'Normal', notes: '', discountRial: 0 })
    orders.value = [order, ...orders.value]
    selectedOrderId.value = order.id
  } catch (error) { showToast(String(error)) }
}

function closeOrderWorkspace() {
  activeView.value = 'Orders'
  selectedOrderId.value = null
}
function closeQuoteWorkspace() { activeView.value='Quotes'; selectedQuoteId.value=null }

function updateOrder(order: OrderRecord) {
  orders.value = orders.value.some((item) => item.id === order.id) ? orders.value.map((item) => item.id === order.id ? order : item) : [order, ...orders.value]
}
function updateQuote(quote: QuoteRecord) { quotes.value=quotes.value.some(item=>item.id===quote.id)?quotes.value.map(item=>item.id===quote.id?quote:item):[quote,...quotes.value] }
async function openNewQuote() { activeView.value='Quotes'; try { const quote=await quotesApi.create({customerId:'',expiryDate:null,notes:'',discountRial:0}); quotes.value=[quote,...quotes.value]; selectedQuoteId.value=quote.id } catch(error) { showToast(String(error)) } }
function openConvertedOrder(orderId: string) { selectedQuoteId.value=null; activeView.value='Orders'; loadOrders(); selectedOrderId.value=orderId }

onMounted(() => { loadOrderCatalog(); loadOrders(); loadQuotes(); loadSuppliers(); loadPurchases() })

function showToast(message: string) {
  toastMessage.value = message
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMessage.value = ''
  }, 2800)
}

</script>

<template>
  <div class="app-shell" :class="{ 'sidebar-collapsed': isSidebarCollapsed }">
    <aside class="sidebar" aria-label="Primary navigation">
      <div class="brand-block">
        <div class="brand-mark" aria-hidden="true">A</div>
        <div class="brand-copy" :class="{ 'is-hidden': isSidebarCollapsed }">
          <span class="brand-name">Atropaten</span>
          <span class="brand-caption">Print shop control</span>
        </div>
      </div>

      <nav class="sidebar-nav">
        <div v-for="section in navigationSections" :key="section.label" class="nav-section">
          <p v-if="!isSidebarCollapsed" class="nav-section-label">{{ section.label }}</p>
          <SidebarNavItem
            v-for="item in section.items"
            :key="item.label"
            :label="item.label"
            :icon="item.icon"
            :active="activeView === item.label"
            :collapsed="isSidebarCollapsed"
            @select="selectView(item.label)"
          />
        </div>
      </nav>

      <div class="sidebar-footer">
        <div class="sync-state">
          <span class="sync-dot" aria-hidden="true"></span>
          <span class="sync-copy" :class="{ 'is-hidden': isSidebarCollapsed }">Local workspace · synced</span>
        </div>
        <div class="sidebar-version" :class="{ 'is-hidden': isSidebarCollapsed }">v0.1 foundation</div>
      </div>
    </aside>

    <div class="app-frame">
      <AppToolbar
        :collapsed="isSidebarCollapsed"
        :search-query="searchQuery"
        :currency-unit="currencyUnit"
        @toggle-sidebar="isSidebarCollapsed = !isSidebarCollapsed"
        @update:search-query="searchQuery = $event"
        @update:currency-unit="currencyUnit = $event"
        @notifications="showToast('You are all caught up.')"
      />

      <main class="workspace" :class="{ 'workspace-dashboard': activeView === 'Dashboard' }" tabindex="-1">
        <Transition name="workspace-view" mode="out-in">
        <div v-if="activeView === 'Dashboard'" key="dashboard" class="dashboard-view">
          <WorkspaceStickyStack>
            <header class="workspace-heading">
            <div>
              <p class="eyebrow">{{ formatDate(currentView.eyebrow) }}</p>
              <h1>{{ currentView.title }}</h1>
              <p class="heading-description">{{ currentView.description }}</p>
            </div>
            <div class="heading-actions">
              <label class="select-control period-select">
                <span class="sr-only">Dashboard period</span>
                <select v-model="selectedPeriod" aria-label="Dashboard period">
                  <option>This week</option>
                  <option>This month</option>
                  <option>This quarter</option>
                </select>
                <ChevronDown class="select-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" />
              </label>
              <button class="button button-primary" type="button" @click="openNewOrder">
                <Plus class="button-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />
                New order
              </button>
            </div>
            </header>
          </WorkspaceStickyStack>

          <div class="dashboard-content">
            <section class="kpi-grid" aria-label="Business summary">
              <KpiCard :value="formatMoney(2_486_000_000, currencyUnit)" :detail="`vs. ${formatMoney(2_204_000_000, currencyUnit)} last week`" title="Sales" trend="+12.8%" :icon="TrendingUp" accent="blue" />
              <KpiCard :value="formatMoney(842_000_000, currencyUnit)" detail="33.9% margin this week" title="Gross profit" trend="+9.4%" :icon="ChartNoAxesCombined" accent="green" />
              <KpiCard :value="formatMoney(1_268_000_000, currencyUnit)" detail="7 open invoices" title="Receivables" trend="3 due soon" :icon="HandCoins" accent="amber" />
              <KpiCard :value="formatMoney(624_000_000, currencyUnit)" detail="4 supplier balances" title="Payables" trend="2 due soon" :icon="ReceiptText" accent="red" />
            </section>

            <section class="dashboard-grid dashboard-grid-primary">
            <SectionPanel title="Production queue" subtitle="4 active jobs · sorted by promised delivery" class="production-panel">
              <template #action>
                <button class="text-button" type="button" @click="selectView('Production')">View production <ArrowRight :size="14" :stroke-width="1.8" aria-hidden="true" /></button>
              </template>
              <div class="table-wrap">
                <table class="data-table production-table">
                  <thead>
                    <tr>
                      <th scope="col">Job</th>
                      <th scope="col">Item</th>
                      <th scope="col">Promised</th>
                      <th scope="col">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="job in productionJobs" :key="job.id">
                      <td>
                        <span class="table-primary">{{ job.id }}</span>
                        <span class="table-secondary">{{ job.customer }}</span>
                      </td>
                      <td>
                        <span class="table-primary">{{ job.item }}</span>
                        <div class="progress-track" :aria-label="`${job.progress}% complete`">
                          <span :style="{ width: `${job.progress}%` }"></span>
                        </div>
                      </td>
                      <td :class="{ 'text-danger': job.status === 'Waiting for proof' }">{{ formatDateTime(job.promisedAt) }}</td>
                      <td><StatusBadge :label="job.status" :tone="job.tone" /></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </SectionPanel>

            <SectionPanel title="Needs attention" subtitle="Financial obligations and exceptions" class="alerts-panel">
              <template #action>
                <span class="attention-count">3</span>
              </template>
              <div class="alert-list">
                <button v-for="alert in financialAlerts" :key="alert.label" class="alert-row" type="button" @click="showToast(`${alert.label}: ${formatMoney(alert.amountRial, currencyUnit)}`)">
                  <span class="alert-icon" :class="`tone-${alert.tone}`" aria-hidden="true"><component :is="alert.icon" :size="15" :stroke-width="1.8" /></span>
                  <span class="alert-content">
                    <span class="alert-label">{{ alert.label }}</span>
                    <span class="alert-detail">{{ alert.detail }}<template v-if="alert.detailDate"> · {{ formatDate(alert.detailDate) }}</template></span>
                  </span>
                  <span class="alert-amount">{{ formatMoney(alert.amountRial, currencyUnit) }}</span>
                  <ChevronRight class="alert-chevron" :size="16" :stroke-width="1.8" aria-hidden="true" />
                </button>
              </div>
              <div class="panel-callout">
                <span class="callout-icon" aria-hidden="true"><Info :size="15" :stroke-width="1.8" /></span>
                <span>Resolve urgent items before the 4:00 PM dispatch run.</span>
              </div>
            </SectionPanel>
          </section>

            <section class="dashboard-grid dashboard-grid-secondary">
            <SectionPanel title="Low stock" subtitle="Materials approaching reorder point" class="low-stock-panel">
              <template #action>
                <button class="text-button" type="button" @click="selectView('Materials')">Open materials <ArrowRight :size="14" :stroke-width="1.8" aria-hidden="true" /></button>
              </template>
              <div class="stock-list">
                <div v-for="item in lowStockItems" :key="item.name" class="stock-row">
                  <component :is="item.icon" class="stock-indicator" :class="`tone-${item.tone}`" :size="15" :stroke-width="1.8" aria-hidden="true" />
                  <span class="stock-name">{{ item.name }}</span>
                  <span class="stock-available">{{ item.available }}</span>
                  <span class="stock-threshold">Reorder at {{ item.threshold }}</span>
                </div>
              </div>
            </SectionPanel>

            <SectionPanel title="Recent transactions" subtitle="Latest activity across the shop" class="transactions-panel">
              <template #action>
                <button class="icon-button" type="button" aria-label="More transaction actions" @click="showToast('Transaction filters are coming soon.')"><MoreHorizontal :size="17" :stroke-width="1.8" aria-hidden="true" /></button>
              </template>
              <div class="transaction-list">
                <div v-for="transaction in recentTransactions" :key="`${transaction.title}-${transaction.party}`" class="transaction-row">
                  <span class="transaction-icon" :class="`tone-${transaction.tone}`" aria-hidden="true"><component :is="transaction.icon" :size="15" :stroke-width="1.8" /></span>
                  <span class="transaction-content">
                    <span class="transaction-title">{{ transaction.title }}</span>
                    <span class="transaction-party">{{ transaction.party }}</span>
                  </span>
                  <span class="transaction-value" :class="{ 'value-positive': transaction.direction === 'in' }">{{ formatSignedMoney(transaction.amountRial, currencyUnit, transaction.direction === 'in' ? '+' : '−') }}</span>
                  <span class="transaction-time">{{ transaction.time }}</span>
                </div>
              </div>
            </SectionPanel>
            </section>
          </div>
          <WorkspaceBottomActions class="quick-actions" aria-labelledby="quick-actions-title">
            <div>
              <p id="quick-actions-title" class="section-kicker">Quick actions</p>
              <p class="quick-actions-help">Jump into the workflows you use most.</p>
            </div>
            <div class="quick-action-buttons">
              <button class="quick-action" type="button" @click="openNewOrder"><Plus class="quick-action-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />New order</button>
              <button class="quick-action" type="button" @click="selectView('Production')"><PackageOpen class="quick-action-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />Start production</button>
              <button class="quick-action" type="button" @click="selectView('Purchases')"><ShoppingCart class="quick-action-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />Record purchase</button>
              <button class="quick-action" type="button" @click="selectView('Accounting')"><Calculator class="quick-action-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />Post expense</button>
            </div>
          </WorkspaceBottomActions>
        </div>

        <div v-else-if="activeView === 'Orders'" key="orders" class="orders-view-transition">
          <Transition name="workspace-view" mode="out-in">
            <OrderWorkspaceView
              v-if="selectedOrder"
              :key="selectedOrder.id"
              :order="selectedOrder"
              :currency-unit="currencyUnit"
              :customers="customers"
              :services="catalogServices"
              :materials="catalogMaterials"
              @back="closeOrderWorkspace"
              @notify="showToast"
              @saved="updateOrder"
            />
            <OrdersView v-else key="orders-list" :orders="orders" :loading="ordersLoading" :error="ordersError" :currency-unit="currencyUnit" @open-order="openOrder" @new-order="openNewOrder" />
          </Transition>
        </div>

        <div v-else-if="activeView === 'Quotes'" key="quotes" class="orders-view-transition">
          <Transition name="workspace-view" mode="out-in">
            <QuoteWorkspaceView v-if="selectedQuote" :key="selectedQuote.id" :quote="selectedQuote" :currency-unit="currencyUnit" :customers="customers" :services="catalogServices" :materials="catalogMaterials" @back="closeQuoteWorkspace" @notify="showToast" @saved="updateQuote" @converted="openConvertedOrder" />
            <QuotesView v-else key="quotes-list" :quotes="quotes" :loading="quotesLoading" :error="quotesError" :currency-unit="currencyUnit" @open-quote="openQuote" @new-quote="openNewQuote" />
          </Transition>
        </div>

        <CustomersView v-else-if="activeView === 'Customers'" key="customers" @notify="showToast" />

        <ServicesView v-else-if="activeView === 'Services'" key="services" :currency-unit="currencyUnit" @notify="showToast" />

        <MaterialsView v-else-if="activeView === 'Materials'" key="materials" :currency-unit="currencyUnit" @notify="showToast" />

        <MachinesView v-else-if="activeView === 'Machines'" key="machines" :currency-unit="currencyUnit" @notify="showToast" />

        <SuppliersView v-else-if="activeView === 'Suppliers'" key="suppliers" @notify="showToast" />

        <PurchasesView v-else-if="activeView === 'Purchases'" key="purchases" :currency-unit="currencyUnit" :suppliers="suppliers" :materials="catalogMaterials" @notify="showToast" />

        <ProductionView v-else-if="activeView === 'Production'" key="production" :currency-unit="currencyUnit" :orders="orders" :materials="catalogMaterials" :machines="catalogMachines" :suppliers="suppliers" @notify="showToast" />

        <AccountingView v-else-if="activeView === 'Accounting'" key="accounting" :currency-unit="currencyUnit" :orders="orders" :suppliers="suppliers" :purchases="purchases" :customers="customers" @notify="showToast" @refresh-orders="loadOrders" />

        <InvoicesView v-else-if="activeView === 'Invoices'" key="invoices" :currency-unit="currencyUnit" :orders="orders" @notify="showToast" @refresh-orders="loadOrders" />

        <ChecksView v-else-if="activeView === 'Checks'" key="checks" :currency-unit="currencyUnit" @notify="showToast" />

        <LoansView v-else-if="activeView === 'Loans'" key="loans" :currency-unit="currencyUnit" @notify="showToast" />

        <section v-else key="empty" class="empty-workspace">
          <div class="empty-workspace-icon" aria-hidden="true"><Sparkles :size="22" :stroke-width="1.8" /></div>
          <p class="eyebrow">Atropaten workspace</p>
          <h1>{{ currentView.title }}</h1>
          <p>{{ currentView.description }}</p>
          <button class="button button-secondary" type="button" @click="selectView('Dashboard')">Back to dashboard</button>
        </section>
        </Transition>
      </main>

      <footer class="status-bar" aria-label="Workspace status">
        <span class="status-live"><span class="sync-dot" aria-hidden="true"></span> Local mode</span>
        <span class="status-divider" aria-hidden="true"></span>
        <span>Last updated just now</span>
        <span class="status-spacer"></span>
        <span>Cash <strong>{{ formatMoney(482_000_000, currencyUnit) }}</strong></span>
        <span>Bank <strong>{{ formatMoney(1_864_000_000, currencyUnit) }}</strong></span>
        <span>Receivable <strong>{{ formatMoney(1_268_000_000, currencyUnit) }}</strong></span>
        <span>Payable <strong>{{ formatMoney(624_000_000, currencyUnit) }}</strong></span>
      </footer>
    </div>

    <Transition name="toast">
      <div v-if="toastMessage" class="toast" role="status">{{ toastMessage }}</div>
    </Transition>
  </div>
</template>
