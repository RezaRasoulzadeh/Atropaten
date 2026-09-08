<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, type Component } from 'vue';
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
  Settings,
  ShoppingCart,
  Sparkles,
  Truck,
  UserRound,
  Users,
} from 'lucide-vue-next';
import NotificationCenter from '../components/ui/NotificationCenter.vue';
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
const selectedOrderId = ref<string | null>(null);
const unsavedOrder = ref<OrderRecord | null>(null);
const orders = ref<OrderRecord[]>([]);
const suppliers = ref<SupplierRecord[]>([]);
const purchases = ref<PurchaseRecord[]>([]);
const ordersLoading = ref(false);
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
  unsavedOrder.value = null;
  if (label === 'Orders') {
    loadOrders();
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
  () =>
    (unsavedOrder.value?.id === selectedOrderId.value ? unsavedOrder.value : null) ??
    orders.value.find((order) => order.id === selectedOrderId.value) ??
    null,
);

async function loadOrders() {
  ordersLoading.value = true;
  try {
    orders.value = await ordersApi.list();
  } catch (error) {
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

async function openDashboardOrder(orderId: string) {
  await Promise.all([loadOrders(), loadOrderCatalog()]);
  if (orders.value.some(order => order.id === orderId)) openOrder(orderId);
  else toast.error('Unable to open this order. Refresh the dashboard and try again.', 'Orders');
}
function openOrder(orderId: string) {
  activeView.value = 'Orders';
  unsavedOrder.value = null;
  selectedOrderId.value = orderId;
}
function openNewOrder() {
  activeView.value = 'Orders';
  const now = new Date().toISOString();
  const id = `new-order-${Date.now()}`;
  unsavedOrder.value = {
    id,
    orderNumber: 'New order',
    customerId: '',
    customerName: '',
    customerPhone: '',
    notes: '',
    createdAt: now,
    updatedAt: now,
    promisedAt: null,
    priority: 'Normal',
    commercialStatus: 'Draft',
    fulfillmentStatus: 'Pending',
    paymentStatus: 'Unpaid',
    subtotalRial: 0,
    discountRial: 0,
    totalRial: 0,
    estimatedCostRial: 0,
    productionJobCount: 0,
    completedProductionJobs: 0,
    inProgressProductionJobs: 0,
    items: [],
  };
  selectedOrderId.value = id;
}

function closeOrderWorkspace() {
  activeView.value = 'Orders';
  selectedOrderId.value = null;
  unsavedOrder.value = null;
}
function removeOrder(orderId: string) {
  orders.value = orders.value.filter((order) => order.id !== orderId);
  closeOrderWorkspace();
}
function updateOrder(order: OrderRecord) {
	if (unsavedOrder.value && unsavedOrder.value.id !== order.id) {
		const draftId = unsavedOrder.value.id;
		unsavedOrder.value = null;
		orders.value = [order, ...orders.value.filter((item) => item.id !== order.id)];
		if (selectedOrderId.value === draftId) selectedOrderId.value = order.id;
		return;
	}
  if (unsavedOrder.value?.id === order.id) {
    unsavedOrder.value = order;
    return;
  }
  unsavedOrder.value = null;
  orders.value = orders.value.some((item) => item.id === order.id)
    ? orders.value.map((item) => (item.id === order.id ? order : item))
    : [order, ...orders.value];
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
      >
        <template #notifications>
          <NotificationCenter
            :currency-unit="currencyUnit"
            :refresh-key="activeView"
            @navigate="selectView"
          />
        </template>
      </AppToolbar>

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
            @open-order="openDashboardOrder"
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
                :is-new="selectedOrderId?.startsWith('new-order-') ?? false"
                @back="closeOrderWorkspace"
                @removed="removeOrder"
                @notify="showToast"
                @saved="updateOrder"
              />
              <OrdersView
                v-else
                key="orders-list"
                :orders="orders"
                :loading="ordersLoading"
                :currency-unit="currencyUnit"
                @open-order="openOrder"
                @new-order="openNewOrder"
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
