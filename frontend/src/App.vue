<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import {
  BarChart3,
  BriefcaseBusiness,
  Calculator,
  CircleDollarSign,
  ClipboardList,
  Factory,
  Landmark,
  LayoutDashboard,
  Package,
  Printer,
  ReceiptText,
  FileText,
  Settings,
  ShoppingCart,
  Sparkles,
  Truck,
  UserRound,
  Users,
} from 'lucide-vue-next'
import AppToolbar from './components/AppToolbar.vue'
import SidebarNavItem from './components/SidebarNavItem.vue'
import { customersApi } from './api/customers'
import { ordersApi, type OrderRecord } from './api/orders'
import { servicesApi } from './api/services'
import { materialsApi } from './api/materials'
import { machinesApi } from './api/machines'
import type { CurrencyUnit } from './utils/currency'
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
import OwnersView from './views/OwnersView.vue'
import DashboardView from './views/DashboardView.vue'
import ReportsView from './views/ReportsView.vue'
import SettingsView from './views/SettingsView.vue'

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

const activeView = ref('Dashboard')
const isSidebarCollapsed = ref(false)
const searchQuery = ref('')
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

const currentView = computed(() => ({
  eyebrow: 'Atropaten workspace',
  title: activeView.value,
  description: 'This workspace is ready for the next product milestone.',
}))

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
        <DashboardView v-if="activeView === 'Dashboard'" key="dashboard" :currency-unit="currencyUnit" @navigate="selectView" @new-order="openNewOrder" @notify="showToast" />

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

        <OwnersView v-else-if="activeView === 'Owners'" key="owners" :currency-unit="currencyUnit" @notify="showToast" />

        <ReportsView v-else-if="activeView === 'Reports'" key="reports" :currency-unit="currencyUnit" @notify="showToast" />

        <SettingsView v-else-if="activeView === 'Settings'" key="settings" @notify="showToast" />

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
        <span>Authoritative data</span>
      </footer>
    </div>

    <Transition name="toast">
      <div v-if="toastMessage" class="toast" role="status">{{ toastMessage }}</div>
    </Transition>
  </div>
</template>
