<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, type Component } from 'vue';
import {
  BarChart3,
  BellRing,
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
} from 'lucide-vue-next';
import AppToolbar from '../components/layout/AppToolbar.vue';
import SidebarNavItem from '../components/layout/SidebarNavItem.vue';
import { customersApi } from '../api/customers';
import { ordersApi, type OrderRecord } from '../api/orders';
import { servicesApi } from '../api/services';
import { materialsApi } from '../api/materials';
import { machinesApi } from '../api/machines';
import type { CurrencyUnit } from '../utils/currency';
import OrderWorkspaceView from '../features/orders/OrderWorkspaceView.vue';
import OrdersView from '../features/orders/OrdersView.vue';
import MaterialsView from '../features/materials/MaterialsView.vue';
import ServicesView from '../features/services/ServicesView.vue';
import MachinesView from '../features/machines/MachinesView.vue';
import CustomersView from '../features/customers/CustomersView.vue';
import QuotesView from '../features/quotes/QuotesView.vue';
import QuoteWorkspaceView from '../features/quotes/QuoteWorkspaceView.vue';
import { quotesApi, type QuoteRecord } from '../api/quotes';
import SuppliersView from '../features/suppliers/SuppliersView.vue';
import PurchasesView from '../features/purchases/PurchasesView.vue';
import ProductionView from '../features/production/ProductionView.vue';
import { suppliersApi, type SupplierRecord } from '../api/suppliers';
import { purchasesApi, type PurchaseRecord } from '../api/purchases';
import AccountingView from '../features/accounting/AccountingView.vue';
import InvoicesView from '../features/invoices/InvoicesView.vue';
import ChecksView from '../features/checks/ChecksView.vue';
import LoansView from '../features/loans/LoansView.vue';
import OwnersView from '../features/owners/OwnersView.vue';
import DashboardView from '../features/dashboard/DashboardView.vue';
import ReportsView from '../features/reports/ReportsView.vue';
import SettingsView from '../features/settings/SettingsView.vue';
import ToastHost from '../components/ui/ToastHost.vue';
import ConfirmDialog from '../components/ui/ConfirmDialog.vue';
import { normalizeError, useToast } from '../ui/feedback';

interface NavigationItem {
  label: string;
  icon: Component;
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
];

const activeView = ref('Dashboard');
const isSidebarCollapsed = ref(false);
const isDrawerOpen = ref(false);
const isDesktop = ref(false);
const drawerPanel = ref<HTMLElement | null>(null);
let desktopMediaQuery: MediaQueryList | undefined;
const sidebarCollapsed = computed(() => isDesktop.value && isSidebarCollapsed.value);
const toolbarCollapsed = computed(() =>
  isDesktop.value ? isSidebarCollapsed.value : !isDrawerOpen.value,
);
const searchQuery = ref('');
const currencyUnit = ref<CurrencyUnit>('Toman');
const hasNotifications = ref(false);
const selectedOrderId = ref<string | null>(null);
const selectedQuoteId = ref<string | null>(null);
const orders = ref<OrderRecord[]>([]);
const quotes = ref<QuoteRecord[]>([]);
const suppliers = ref<SupplierRecord[]>([]);
const purchases = ref<PurchaseRecord[]>([]);
const ordersLoading = ref(false);
const ordersError = ref('');
const quotesLoading = ref(false);
const quotesError = ref('');
const customers = ref<any[]>([]);
const catalogServices = ref<any[]>([]);
const catalogMaterials = ref<any[]>([]);
const catalogMachines = ref<any[]>([]);
const toast = useToast();

const currentView = computed(() => ({
  eyebrow: 'Atropaten workspace',
  title: activeView.value,
  description: 'This workspace is ready for the next product milestone.',
}));

function selectView(label: string) {
  activeView.value = label;
  isDrawerOpen.value = false;
  selectedOrderId.value = null;
  selectedQuoteId.value = null;
  if (label === 'Orders') {
    loadOrders();
    loadOrderCatalog();
  }
  if (label === 'Quotes') {
    loadQuotes();
    loadOrderCatalog();
  }
  if (label === 'Suppliers') loadSuppliers();
  if (label === 'Purchases') {
    loadSuppliers();
    loadPurchases();
    loadOrderCatalog();
  }
}

const selectedOrder = computed(
  () => orders.value.find((order) => order.id === selectedOrderId.value) ?? null,
);
const selectedQuote = computed(
  () => quotes.value.find((quote) => quote.id === selectedQuoteId.value) ?? null,
);

async function loadOrders() {
  ordersLoading.value = true;
  ordersError.value = '';
  try {
    orders.value = await ordersApi.list();
  } catch (error) {
    ordersError.value = normalizeError(error).message;
    toast.error(error, 'Orders');
  } finally {
    ordersLoading.value = false;
  }
}

async function loadOrderCatalog() {
  try {
    customers.value = await customersApi.list(true);
    catalogServices.value = await servicesApi.list(true);
    catalogMaterials.value = await materialsApi.list(true);
    catalogMachines.value = await machinesApi.list(true);
  } catch (error) {
    toast.error(error, 'Catalog');
  }
}
async function loadQuotes() {
  quotesLoading.value = true;
  quotesError.value = '';
  try {
    quotes.value = await quotesApi.list();
  } catch (error) {
    quotesError.value = normalizeError(error).message;
    toast.error(error, 'Quotes');
  } finally {
    quotesLoading.value = false;
  }
}
async function loadSuppliers() {
  try {
    suppliers.value = await suppliersApi.list(true);
  } catch (error) {
    toast.error(error, 'Suppliers');
  }
}
async function loadPurchases() {
  try {
    purchases.value = await purchasesApi.list();
  } catch (error) {
    toast.error(error, 'Purchases');
  }
}

function openOrder(orderId: string) {
  activeView.value = 'Orders';
  selectedOrderId.value = orderId;
}
function openQuote(quoteId: string) {
  activeView.value = 'Quotes';
  selectedQuoteId.value = quoteId;
}

async function openNewOrder() {
  activeView.value = 'Orders';
  try {
    const order = await ordersApi.create({
      customerId: '',
      promisedAt: null,
      priority: 'Normal',
      notes: '',
      discountRial: 0,
    });
    orders.value = [order, ...orders.value];
    selectedOrderId.value = order.id;
  } catch (error) {
    toast.error(error, 'New order');
  }
}

function closeOrderWorkspace() {
  activeView.value = 'Orders';
  selectedOrderId.value = null;
}
function closeQuoteWorkspace() {
  activeView.value = 'Quotes';
  selectedQuoteId.value = null;
}

function updateOrder(order: OrderRecord) {
  orders.value = orders.value.some((item) => item.id === order.id)
    ? orders.value.map((item) => (item.id === order.id ? order : item))
    : [order, ...orders.value];
}
function updateQuote(quote: QuoteRecord) {
  quotes.value = quotes.value.some((item) => item.id === quote.id)
    ? quotes.value.map((item) => (item.id === quote.id ? quote : item))
    : [quote, ...quotes.value];
}
async function openNewQuote() {
  activeView.value = 'Quotes';
  try {
    const quote = await quotesApi.create({
      customerId: '',
      expiryDate: null,
      notes: '',
      discountRial: 0,
    });
    quotes.value = [quote, ...quotes.value];
    selectedQuoteId.value = quote.id;
  } catch (error) {
    toast.error(error, 'New quote');
  }
}
function openConvertedOrder(orderId: string) {
  selectedQuoteId.value = null;
  activeView.value = 'Orders';
  loadOrders();
  selectedOrderId.value = orderId;
}

function syncViewport() {
  isDesktop.value = desktopMediaQuery?.matches ?? window.innerWidth >= 896;
}

function toggleSidebar() {
  if (isDesktop.value) isSidebarCollapsed.value = !isSidebarCollapsed.value;
  else isDrawerOpen.value = !isDrawerOpen.value;
}

function closeDrawerOnOutsideClick(event: MouseEvent) {
  if (isDesktop.value || !isDrawerOpen.value) return;
  const target = event.target;
  if (!(target instanceof Element)) return;
  if (drawerPanel.value?.contains(target) || target.closest('[data-drawer-toggle]')) return;
  isDrawerOpen.value = false;
}

onMounted(() => {
  desktopMediaQuery = window.matchMedia('(min-width: 56rem)');
  syncViewport();
  desktopMediaQuery.addEventListener('change', syncViewport);
  document.addEventListener('click', closeDrawerOnOutsideClick);
  loadOrderCatalog();
  loadOrders();
  loadQuotes();
  loadSuppliers();
  loadPurchases();
});

onBeforeUnmount(() => {
  desktopMediaQuery?.removeEventListener('change', syncViewport);
  document.removeEventListener('click', closeDrawerOnOutsideClick);
});

function showToast(message: string) {
  const isFailure =
    /^(error:|failed|unable|could not|cannot|invalid|not found|conflict|required)/i.test(
      message.trim(),
    );
  if (isFailure) toast.error(normalizeError(message), 'Operation failed');
  else toast.success(message);
}

function openNotifications() {
  hasNotifications.value = false;
  toast.info('You are all caught up.');
}
</script>

<template>
  <div
    class="drawer lg:drawer-open h-screen min-h-0 overflow-hidden bg-base-200 font-sans text-base-content"
    :class="{ 'drawer-open': isDesktop || isDrawerOpen }"
  >
    <input id="atropaten-drawer" v-model="isDrawerOpen" type="checkbox" class="drawer-toggle" />

    <div class="drawer-content grid h-full min-h-0 min-w-0 grid-rows-[auto_minmax(0,1fr)]">
      <AppToolbar
        class="sticky top-0 z-20"
        :search-query="searchQuery"
        :currency-unit="currencyUnit"
        :collapsed="toolbarCollapsed"
        @toggle-sidebar="toggleSidebar"
        @update:search-query="searchQuery = $event"
        @update:currency-unit="currencyUnit = $event"
        @new-order="openNewOrder"
        @navigate="selectView"
      />

      <main
        class="min-h-0 min-w-0 overflow-y-auto bg-base-200 p-4 pt-0 lg:p-6 lg:pt-0"
        tabindex="-1"
      >
        <Transition mode="out-in">
          <DashboardView
            v-if="activeView === 'Dashboard'"
            key="dashboard"
            :currency-unit="currencyUnit"
            @navigate="selectView"
            @new-order="openNewOrder"
            @notify="showToast"
          />

          <div v-else-if="activeView === 'Orders'" key="orders">
            <Transition mode="out-in">
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
              <OrdersView
                v-else
                key="orders-list"
                :orders="orders"
                :loading="ordersLoading"
                :error="ordersError"
                :currency-unit="currencyUnit"
                @open-order="openOrder"
                @new-order="openNewOrder"
              />
            </Transition>
          </div>

          <div v-else-if="activeView === 'Quotes'" key="quotes">
            <Transition mode="out-in">
              <QuoteWorkspaceView
                v-if="selectedQuote"
                :key="selectedQuote.id"
                :quote="selectedQuote"
                :currency-unit="currencyUnit"
                :customers="customers"
                :services="catalogServices"
                :materials="catalogMaterials"
                @back="closeQuoteWorkspace"
                @notify="showToast"
                @saved="updateQuote"
                @converted="openConvertedOrder"
              />
              <QuotesView
                v-else
                key="quotes-list"
                :quotes="quotes"
                :loading="quotesLoading"
                :error="quotesError"
                :currency-unit="currencyUnit"
                @open-quote="openQuote"
                @new-quote="openNewQuote"
              />
            </Transition>
          </div>

          <CustomersView
            v-else-if="activeView === 'Customers'"
            key="customers"
            @notify="showToast"
          />

          <ServicesView
            v-else-if="activeView === 'Services'"
            key="services"
            :currency-unit="currencyUnit"
            @notify="showToast"
          />

          <MaterialsView
            v-else-if="activeView === 'Materials'"
            key="materials"
            :currency-unit="currencyUnit"
            @notify="showToast"
          />

          <MachinesView
            v-else-if="activeView === 'Machines'"
            key="machines"
            :currency-unit="currencyUnit"
            @notify="showToast"
          />

          <SuppliersView
            v-else-if="activeView === 'Suppliers'"
            key="suppliers"
            @notify="showToast"
          />

          <PurchasesView
            v-else-if="activeView === 'Purchases'"
            key="purchases"
            :currency-unit="currencyUnit"
            :suppliers="suppliers"
            :materials="catalogMaterials"
            @notify="showToast"
          />

          <ProductionView
            v-else-if="activeView === 'Production'"
            key="production"
            :currency-unit="currencyUnit"
            :orders="orders"
            :materials="catalogMaterials"
            :machines="catalogMachines"
            :suppliers="suppliers"
            @notify="showToast"
          />

          <AccountingView
            v-else-if="activeView === 'Accounting'"
            key="accounting"
            :currency-unit="currencyUnit"
            :orders="orders"
            :suppliers="suppliers"
            :purchases="purchases"
            :customers="customers"
            @notify="showToast"
            @refresh-orders="loadOrders"
          />

          <InvoicesView
            v-else-if="activeView === 'Invoices'"
            key="invoices"
            :currency-unit="currencyUnit"
            :orders="orders"
            @notify="showToast"
            @refresh-orders="loadOrders"
          />

          <ChecksView
            v-else-if="activeView === 'Checks'"
            key="checks"
            :currency-unit="currencyUnit"
            @notify="showToast"
          />

          <LoansView
            v-else-if="activeView === 'Loans'"
            key="loans"
            :currency-unit="currencyUnit"
            @notify="showToast"
          />

          <OwnersView
            v-else-if="activeView === 'Owners'"
            key="owners"
            :currency-unit="currencyUnit"
            @notify="showToast"
          />

          <ReportsView
            v-else-if="activeView === 'Reports'"
            key="reports"
            :currency-unit="currencyUnit"
            @notify="showToast"
          />

          <SettingsView v-else-if="activeView === 'Settings'" key="settings" @notify="showToast" />

          <section
            v-else
            key="empty"
            class="flex min-h-full flex-col items-center justify-center gap-3 text-center"
          >
            <div class="text-primary" aria-hidden="true">
              <Sparkles :size="22" :stroke-width="1.8" />
            </div>
            <p class="text-xs uppercase tracking-wide text-base-content/60">Atropaten workspace</p>
            <h1>{{ currentView.title }}</h1>
            <p>{{ currentView.description }}</p>
            <button class="btn btn-outline" type="button" @click="selectView('Dashboard')">
              Back to dashboard
            </button>
          </section>
        </Transition>
      </main>
    </div>

    <div class="drawer-side z-30">
      <label
        for="atropaten-drawer"
        aria-label="Close navigation"
        class="drawer-overlay"
        @click.prevent="isDrawerOpen = false"
      ></label>
      <aside
        ref="drawerPanel"
        class="flex h-full min-h-full max-h-full flex-col overflow-hidden border-e border-base-300 bg-base-100"
        :class="sidebarCollapsed ? 'w-[4.5rem]' : 'w-[14.5rem]'"
        aria-label="Primary navigation"
      >
        <div
          class="navbar min-h-16 shrink-0 border-b border-base-300 bg-base-100"
          :class="sidebarCollapsed ? 'justify-center px-0' : 'gap-3 px-4'"
        >
          <div
            class="grid size-8 shrink-0 place-items-center rounded bg-primary font-bold text-primary-content"
            aria-hidden="true"
          >
            A
          </div>
          <div v-if="!sidebarCollapsed" class="min-w-0">
            <span class="block truncate text-sm font-bold">Atropaten</span>
            <span class="block text-xs leading-4 text-base-content/60">Print shop control</span>
          </div>
          <button
            v-if="!sidebarCollapsed"
            class="btn btn-ghost btn-square btn-sm relative ms-auto"
            type="button"
            aria-label="Notifications"
            title="Notifications"
            @click="openNotifications"
          >
            <BellRing :size="17" :stroke-width="1.8" aria-hidden="true" />
            <span
              v-if="hasNotifications"
              class="absolute end-1 top-1 size-2 rounded-full bg-error ring-2 ring-base-100"
              aria-hidden="true"
            ></span>
          </button>
        </div>

        <nav
          class="menu w-full flex-nowrap items-stretch min-h-0 min-w-0 flex-1 overflow-x-hidden overflow-y-scroll overscroll-contain p-2"
        >
          <div v-for="section in navigationSections" :key="section.label" class="mb-3 last:mb-0">
            <p
              v-if="!sidebarCollapsed"
              class="mb-1 px-2 text-xs font-semibold leading-4 uppercase tracking-wide text-base-content/50"
            >
              {{ section.label }}
            </p>
            <SidebarNavItem
              v-for="item in section.items"
              :key="item.label"
              :label="item.label"
              :icon="item.icon"
              :active="activeView === item.label"
              :collapsed="sidebarCollapsed"
              @select="selectView(item.label)"
            />
          </div>
        </nav>

        <div class="shrink-0 border-t border-base-300 p-3 text-xs text-base-content/60">
          <div class="flex items-center gap-2">
            <span class="size-2 shrink-0 rounded-full bg-success" aria-hidden="true"></span>
            <span v-if="!sidebarCollapsed">Local workspace</span>
          </div>
          <div v-if="!sidebarCollapsed" class="mt-1 text-xs leading-4">v0.1</div>
        </div>
      </aside>
    </div>

    <ToastHost />
    <ConfirmDialog />
  </div>
</template>
