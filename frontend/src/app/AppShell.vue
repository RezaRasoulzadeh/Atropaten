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
import { invoicesApi, type InvoiceRecord } from '../api/invoices';
import { checksApi, type CheckRecord } from '../api/checks';
import { loansApi, type LoanRecord } from '../api/loans';
import { ownersApi, type OwnerRecord } from '../api/owners';
import { productionApi, type ProductionJobRecord } from '../api/production';
import { accountingApi, type AccountRecord, type ExpenseRecord, type FinancialAccountRecord, type PaymentRecord, type TransferRecord } from '../api/accounting';
import AccountingView from '../features/accounting/AccountingView.vue';
import InvoicesView from '../features/invoices/InvoicesView.vue';
import InvoiceWorkspaceView from '../features/invoices/InvoiceWorkspaceView.vue';
import ChecksView from '../features/checks/ChecksView.vue';
import LoansView from '../features/loans/LoansView.vue';
import LoanWorkspaceView from '../features/loans/LoanWorkspaceView.vue';
import OwnersView from '../features/owners/OwnersView.vue';
import DashboardView from '../features/dashboard/DashboardView.vue';
import ReportsView from '../features/reports/ReportsView.vue';
import SettingsView from '../features/settings/SettingsView.vue';
import ToastHost from '../components/ui/ToastHost.vue';
import ConfirmDialog from '../components/ui/ConfirmDialog.vue';
import EmptyState from '../components/ui/EmptyState.vue';
import { normalizeError, useToast } from '../ui/feedback';

interface NavigationItem {
  label: string;
  icon: Component;
}

interface GlobalSearchResult {
  id: string;
  view: string;
  title: string;
  subtitle: string;
  detail: string;
}

interface SearchIndexEntry extends GlobalSearchResult {
  searchText: string;
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
const selectedInvoiceId = ref<string | null>(null);
const selectedLoanId = ref<string | null>(null);
const unsavedOrder = ref<OrderRecord | null>(null);
const orders = ref<OrderRecord[]>([]);
const suppliers = ref<SupplierRecord[]>([]);
const purchases = ref<PurchaseRecord[]>([]);
const ordersLoading = ref(false);
const customers = ref<any[]>([]);
const catalogServices = ref<any[]>([]);
const catalogMaterials = ref<any[]>([]);
const catalogMachines = ref<any[]>([]);
const invoices = ref<InvoiceRecord[]>([]);
const checks = ref<CheckRecord[]>([]);
const loans = ref<LoanRecord[]>([]);
const owners = ref<OwnerRecord[]>([]);
const productionJobs = ref<ProductionJobRecord[]>([]);
const accounts = ref<AccountRecord[]>([]);
const financialAccounts = ref<FinancialAccountRecord[]>([]);
const expenses = ref<ExpenseRecord[]>([]);
const payments = ref<PaymentRecord[]>([]);
const transfers = ref<TransferRecord[]>([]);
const globalSearchLoading = ref(true);
const toast = useToast();

function searchValue(value: unknown) {
  return value == null ? '' : String(value);
}

function searchEntry(
  view: string,
  id: unknown,
  title: unknown,
  subtitle: unknown,
  detail: unknown,
  ...extra: unknown[]
): SearchIndexEntry {
  const safeId = searchValue(id);
  const safeTitle = searchValue(title) || safeId;
  const safeSubtitle = searchValue(subtitle);
  const safeDetail = searchValue(detail);
  return {
    id: safeId,
    view,
    title: safeTitle,
    subtitle: safeSubtitle,
    detail: safeDetail,
    searchText: [safeTitle, safeSubtitle, safeDetail, ...extra.map(searchValue)]
      .join(' ')
      .toLocaleLowerCase(),
  };
}

const globalSearchIndex = computed<SearchIndexEntry[]>(() => [
  ...orders.value.map((order) =>
    searchEntry(
      'Orders',
      order.id,
      order.orderNumber,
      order.customerName || 'Walk-in customer',
      `${order.commercialStatus} · ${order.fulfillmentStatus}`,
      order.customerPhone,
      order.notes,
    ),
  ),
  ...customers.value.map((customer) =>
    searchEntry('Customers', customer.id, customer.name, customer.phone || customer.email, customer.active ? 'Active' : 'Archived', customer.address, customer.notes),
  ),
  ...catalogServices.value.map((service) =>
    searchEntry('Services', service.id, service.name, service.code, service.category, service.description),
  ),
  ...catalogMaterials.value.map((material) =>
    searchEntry('Materials', material.id, material.name, material.code, material.category || material.consumptionUnit, material.notes),
  ),
  ...catalogMachines.value.map((machine) =>
    searchEntry('Machines', machine.id, machine.name, machine.code, machine.category, machine.notes),
  ),
  ...suppliers.value.map((supplier) =>
    searchEntry('Suppliers', supplier.id, supplier.name, supplier.code || supplier.phone, supplier.active ? 'Active' : 'Archived', supplier.email, supplier.address),
  ),
  ...purchases.value.map((purchase) =>
    searchEntry('Purchases', purchase.id, purchase.purchaseNumber, purchase.supplierName || 'No supplier', purchase.status, purchase.supplierInvoiceNumber, purchase.notes),
  ),
  ...invoices.value.map((invoice) =>
    searchEntry('Invoices', invoice.id, invoice.invoiceNumber, invoice.customerName || 'Walk-in customer', invoice.status, invoice.orderId, invoice.notes),
  ),
  ...checks.value.map((check) =>
    searchEntry('Checks', check.id, check.checkNumber, check.payerPayee, `${check.direction} · ${check.status}`, check.bank, check.accountDescriptor, check.notes),
  ),
  ...loans.value.map((loan) =>
    searchEntry('Loans', loan.id, loan.loanNumber, loan.counterpartyName, `${loan.direction} · ${loan.status}`, loan.notes),
  ),
  ...owners.value.map((owner) =>
    searchEntry('Owners', owner.id, owner.name, owner.phone || owner.email, owner.active ? 'Active' : 'Archived', owner.notes),
  ),
  ...productionJobs.value.map((job) =>
    searchEntry('Production', job.id, job.jobNumber, job.serviceName, `${job.status} · ${job.orderId}`, job.notes, job.outsourceDescription),
  ),
  ...accounts.value.map((account) =>
    searchEntry('Accounting', account.id, account.name, account.code, account.type, account.active ? 'Active' : 'Archived'),
  ),
  ...financialAccounts.value.map((account) =>
    searchEntry('Accounting', account.id, account.name, account.type, account.bankName, account.accountNumber, account.details),
  ),
  ...expenses.value.map((expense) =>
    searchEntry('Accounting', expense.id, expense.expenseNumber, expense.payee || expense.description, `${expense.status} · Expense`, expense.notes),
  ),
  ...payments.value.map((payment) =>
    searchEntry('Accounting', payment.id, payment.paymentNumber, payment.reference || payment.method, `${payment.direction} · ${payment.status}`, payment.notes),
  ),
  ...transfers.value.map((transfer) =>
    searchEntry('Accounting', transfer.id, transfer.transferNumber, transfer.reference, `Transfer · ${transfer.status}`, transfer.notes),
  ),
]);

const searchResults = computed(() => {
  const query = searchQuery.value.trim().toLocaleLowerCase();
  if (!query) return [];
  const terms = query.split(/\s+/).filter(Boolean);
  return globalSearchIndex.value
    .filter((result) => terms.every((term) => result.searchText.includes(term)))
    .map(({ searchText: _searchText, ...result }) => result)
    .slice(0, 12);
});

const currentView = computed(() => ({
  eyebrow: 'Atropaten workspace',
  title: activeView.value,
  description: 'This workspace is ready for the next product milestone.',
}));

function selectView(label: string) {
  activeView.value = label;
  isDrawerOpen.value = false;
  selectedOrderId.value = null;
  selectedInvoiceId.value = null;
  selectedLoanId.value = null;
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

async function loadGlobalSearchData() {
  globalSearchLoading.value = true;
  const results = await Promise.allSettled([
    invoicesApi.list(),
    checksApi.list(),
    loansApi.list(),
    ownersApi.list(true),
    productionApi.list('All'),
    accountingApi.accounts(),
    accountingApi.financialAccounts(),
    accountingApi.expenses(),
    accountingApi.payments(),
    accountingApi.transfers(),
  ]);
  const [invoiceResult, checkResult, loanResult, ownerResult, productionResult, accountResult, financialAccountResult, expenseResult, paymentResult, transferResult] = results;
  if (invoiceResult.status === 'fulfilled') invoices.value = invoiceResult.value;
  if (checkResult.status === 'fulfilled') checks.value = checkResult.value;
  if (loanResult.status === 'fulfilled') loans.value = loanResult.value;
  if (ownerResult.status === 'fulfilled') owners.value = ownerResult.value;
  if (productionResult.status === 'fulfilled') productionJobs.value = productionResult.value;
  if (accountResult.status === 'fulfilled') accounts.value = accountResult.value;
  if (financialAccountResult.status === 'fulfilled') financialAccounts.value = financialAccountResult.value;
  if (expenseResult.status === 'fulfilled') expenses.value = expenseResult.value;
  if (paymentResult.status === 'fulfilled') payments.value = paymentResult.value;
  if (transferResult.status === 'fulfilled') transfers.value = transferResult.value;
  globalSearchLoading.value = false;
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
function openInvoice(invoiceId: string) {
  activeView.value = 'Invoices';
  selectedInvoiceId.value = invoiceId;
}
function closeInvoiceWorkspace() {
  activeView.value = 'Invoices';
  selectedInvoiceId.value = null;
}
function openLoan(loanId: string) {
  activeView.value = 'Loans';
  selectedLoanId.value = loanId;
}

async function selectSearchResult(result: GlobalSearchResult) {
  searchQuery.value = '';
  if (result.view === 'Orders') {
    await loadOrders();
    if (orders.value.some((order) => order.id === result.id)) openOrder(result.id);
    else selectView('Orders');
    return;
  }
  if (result.view === 'Invoices') {
    openInvoice(result.id);
    return;
  }
  if (result.view === 'Loans') {
    openLoan(result.id);
    return;
  }
  selectView(result.view);
}
function openNewLoan() {
  activeView.value = 'Loans';
  selectedLoanId.value = 'new';
}
function closeLoanWorkspace() {
  activeView.value = 'Loans';
  selectedLoanId.value = null;
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
    archived: false,
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

function handleWorkspaceNavigation(event: Event) {
  const view = (event as CustomEvent<string>).detail;
  if (view) selectView(view);
}

onMounted(() => {
  desktopMediaQuery = window.matchMedia('(min-width: 56rem)');
  syncViewport();
  desktopMediaQuery.addEventListener('change', syncViewport);
  document.addEventListener('click', closeDrawerOnOutsideClick);
  window.addEventListener('atropaten:navigate', handleWorkspaceNavigation);
  loadOrderCatalog();
  loadOrders();
  loadSuppliers();
  loadPurchases();
  loadGlobalSearchData();
});

onBeforeUnmount(() => {
  desktopMediaQuery?.removeEventListener('change', syncViewport);
  document.removeEventListener('click', closeDrawerOnOutsideClick);
  window.removeEventListener('atropaten:navigate', handleWorkspaceNavigation);
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
        :search-results="searchResults"
        :search-loading="globalSearchLoading"
        :currency-unit="currencyUnit"
        :collapsed="toolbarCollapsed"
        @toggle-sidebar="toggleSidebar"
        @update:search-query="searchQuery = $event"
        @select-search-result="selectSearchResult"
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
        class="flex min-h-0 min-w-0 flex-col overflow-y-auto bg-base-200 p-4 pt-0 lg:p-6 lg:pt-0"
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

          <div
            v-else-if="activeView === 'Orders'"
            key="orders"
            class="flex h-full min-h-0 min-w-0 flex-col"
          >
            <Transition mode="out-in">
              <OrderWorkspaceView
                v-if="selectedOrder"
                :order="selectedOrder"
                :currency-unit="currencyUnit"
                :customers="customers"
                :services="catalogServices"
                :materials="catalogMaterials"
                :machines="catalogMachines"
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
                @edit-order="openOrder"
                @order-updated="updateOrder"
                @order-removed="removeOrder"
                @notify="showToast"
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
            :suppliers="suppliers"
            @notify="showToast"
          />

          <div v-else-if="activeView === 'Invoices'" key="invoices">
            <Transition mode="out-in">
              <InvoiceWorkspaceView
                v-if="selectedInvoiceId"
                :key="selectedInvoiceId"
                :invoice-id="selectedInvoiceId"
                :currency-unit="currencyUnit"
                @back="closeInvoiceWorkspace"
                @notify="showToast"
                @refresh-orders="loadOrders"
              />
              <InvoicesView
                v-else
                key="invoices-list"
                :currency-unit="currencyUnit"
                :orders="orders"
                @notify="showToast"
                @refresh-orders="loadOrders"
                @open-invoice="openInvoice"
              />
            </Transition>
          </div>

          <ChecksView
            v-else-if="activeView === 'Checks'"
            key="checks"
            :currency-unit="currencyUnit"
            @notify="showToast"
          />

          <div v-else-if="activeView === 'Loans'" key="loans">
            <Transition mode="out-in">
              <LoanWorkspaceView
                v-if="selectedLoanId"
                :key="selectedLoanId"
                :loan-id="selectedLoanId === 'new' ? null : selectedLoanId"
                :is-new="selectedLoanId === 'new'"
                :currency-unit="currencyUnit"
                @back="closeLoanWorkspace"
                @notify="showToast"
              />
              <LoansView
                v-else
                key="loans-list"
                :currency-unit="currencyUnit"
                @notify="showToast"
                @open-loan="openLoan"
                @new-loan="openNewLoan"
              />
            </Transition>
          </div>

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

          <EmptyState
            v-else
            key="empty"
            title="Workspace view unavailable"
            :description="`${currentView.title} is not available yet. Return to the dashboard to continue.`"
          >
            <template #action>
              <button class="btn btn-primary" type="button" @click="selectView('Dashboard')">
                Back to dashboard
              </button>
            </template>
          </EmptyState>
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
