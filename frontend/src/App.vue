<script setup lang="ts">
import { computed, ref } from 'vue'
import AppToolbar from './components/AppToolbar.vue'
import KpiCard from './components/KpiCard.vue'
import SectionPanel from './components/SectionPanel.vue'
import SidebarNavItem from './components/SidebarNavItem.vue'
import StatusBadge from './components/StatusBadge.vue'

type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate'

interface NavigationItem {
  label: string
  icon: string
}

const navigationSections: { label: string; items: NavigationItem[] }[] = [
  {
    label: 'Workspace',
    items: [
      { label: 'Dashboard', icon: 'dashboard' },
      { label: 'Orders', icon: 'orders' },
      { label: 'Production', icon: 'production' },
      { label: 'Customers', icon: 'customers' },
    ],
  },
  {
    label: 'Catalog & purchasing',
    items: [
      { label: 'Services', icon: 'services' },
      { label: 'Materials', icon: 'materials' },
      { label: 'Purchases', icon: 'purchases' },
      { label: 'Suppliers', icon: 'suppliers' },
    ],
  },
  {
    label: 'Finance',
    items: [
      { label: 'Accounting', icon: 'accounting' },
      { label: 'Checks', icon: 'checks' },
      { label: 'Loans', icon: 'loans' },
      { label: 'Owners', icon: 'owners' },
    ],
  },
  {
    label: 'Insights & setup',
    items: [
      { label: 'Reports', icon: 'reports' },
      { label: 'Settings', icon: 'settings' },
    ],
  },
]

const viewDescriptions: Record<string, { eyebrow: string; title: string; description: string }> = {
  Dashboard: {
    eyebrow: 'Monday, 12 August 2024',
    title: 'Good morning, Reza',
    description: 'Here is what needs your attention today.',
  },
}

const activeView = ref('Dashboard')
const isSidebarCollapsed = ref(false)
const searchQuery = ref('')
const selectedPeriod = ref('This week')
const toastMessage = ref('')
let toastTimer: ReturnType<typeof setTimeout> | undefined

const currentView = computed(() => viewDescriptions[activeView.value] ?? {
  eyebrow: 'Atropaten workspace',
  title: activeView.value,
  description: 'This workspace is ready for the next product milestone.',
})

const productionJobs = [
  { id: 'ORD-1048', customer: 'Mehr Studio', item: 'Business cards · 500 pcs', due: 'Today, 15:30', status: 'In progress', tone: 'blue' as Tone, progress: 72 },
  { id: 'ORD-1045', customer: 'Arman Foods', item: 'Menu · A3 laminated', due: 'Today, 17:00', status: 'Ready to print', tone: 'green' as Tone, progress: 100 },
  { id: 'ORD-1042', customer: 'Nika Events', item: 'Event banners · 3 pcs', due: 'Tomorrow, 10:00', status: 'Queued', tone: 'slate' as Tone, progress: 18 },
  { id: 'ORD-1039', customer: 'Pendar Clinic', item: 'Appointment cards · 1,000 pcs', due: 'Overdue · 2h', status: 'Waiting for proof', tone: 'amber' as Tone, progress: 45 },
]

const financialAlerts = [
  { label: 'Check due today', detail: 'Pars Paper · #CH-2094', amount: '$1,280', tone: 'red' as Tone, icon: 'checks' },
  { label: 'Installment due this week', detail: 'Digital press loan · 16 Aug', amount: '$860', tone: 'amber' as Tone, icon: 'loans' },
  { label: 'Invoice overdue', detail: 'Pendar Clinic · 12 days', amount: '$2,450', tone: 'red' as Tone, icon: 'receipt' },
]

const lowStockItems = [
  { name: 'A4 300gsm matte', available: '3 reams', threshold: '5 reams', tone: 'red' as Tone },
  { name: 'Gloss lamination roll', available: '42 m', threshold: '50 m', tone: 'amber' as Tone },
  { name: 'Black toner · Ricoh', available: '18%', threshold: '25%', tone: 'amber' as Tone },
]

const recentTransactions = [
  { title: 'Payment received', party: 'Mehr Studio · ORD-1048', amount: '+$920', time: '10:42 AM', tone: 'green' as Tone, icon: 'arrow-down' },
  { title: 'Purchase recorded', party: 'Pars Paper · PUR-0061', amount: '−$1,280', time: '09:18 AM', tone: 'slate' as Tone, icon: 'purchases' },
  { title: 'Expense posted', party: 'Courier and delivery', amount: '−$84', time: 'Yesterday', tone: 'slate' as Tone, icon: 'receipt' },
  { title: 'Payment received', party: 'Arman Foods · ORD-1045', amount: '+$640', time: 'Yesterday', tone: 'green' as Tone, icon: 'arrow-down' },
]

function selectView(label: string) {
  activeView.value = label
}

function showToast(message: string) {
  toastMessage.value = message
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMessage.value = ''
  }, 2800)
}

function createNewOrder() {
  showToast('New order workspace will be available in the next milestone.')
}
</script>

<template>
  <div class="app-shell" :class="{ 'sidebar-collapsed': isSidebarCollapsed }">
    <aside class="sidebar" aria-label="Primary navigation">
      <div class="brand-block">
        <div class="brand-mark" aria-hidden="true">A</div>
        <div v-if="!isSidebarCollapsed" class="brand-copy">
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
          <span v-if="!isSidebarCollapsed">Local workspace · synced</span>
        </div>
        <div v-if="!isSidebarCollapsed" class="sidebar-version">v0.1 foundation</div>
      </div>
    </aside>

    <div class="app-frame">
      <AppToolbar
        :collapsed="isSidebarCollapsed"
        :search-query="searchQuery"
        @toggle-sidebar="isSidebarCollapsed = !isSidebarCollapsed"
        @update:search-query="searchQuery = $event"
        @notifications="showToast('You are all caught up.')"
      />

      <main class="workspace" tabindex="-1">
        <template v-if="activeView === 'Dashboard'">
          <header class="workspace-heading">
            <div>
              <p class="eyebrow">{{ currentView.eyebrow }}</p>
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
              </label>
              <button class="button button-primary" type="button" @click="createNewOrder">
                <span class="button-icon" aria-hidden="true">+</span>
                New order
              </button>
            </div>
          </header>

          <section class="kpi-grid" aria-label="Business summary">
            <KpiCard title="Sales" value="$24,860" detail="vs. $22,040 last week" trend="+12.8%" icon="sales" accent="blue" />
            <KpiCard title="Gross profit" value="$8,420" detail="33.9% margin this week" trend="+9.4%" icon="profit" accent="green" />
            <KpiCard title="Receivables" value="$12,680" detail="7 open invoices" trend="3 due soon" icon="receivables" accent="amber" />
            <KpiCard title="Payables" value="$6,240" detail="4 supplier balances" trend="2 due soon" icon="payables" accent="red" />
          </section>

          <section class="dashboard-grid dashboard-grid-primary">
            <SectionPanel title="Production queue" subtitle="4 active jobs · sorted by promised delivery" class="production-panel">
              <template #action>
                <button class="text-button" type="button" @click="selectView('Production')">View production <span aria-hidden="true">→</span></button>
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
                      <td :class="{ 'text-danger': job.status === 'Waiting for proof' }">{{ job.due }}</td>
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
                <button v-for="alert in financialAlerts" :key="alert.label" class="alert-row" type="button" @click="showToast(`${alert.label}: ${alert.amount}`)">
                  <span class="alert-icon" :class="`tone-${alert.tone}`" aria-hidden="true">
                    <span v-if="alert.icon === 'checks'">✓</span>
                    <span v-else-if="alert.icon === 'loans'">↗</span>
                    <span v-else>!</span>
                  </span>
                  <span class="alert-content">
                    <span class="alert-label">{{ alert.label }}</span>
                    <span class="alert-detail">{{ alert.detail }}</span>
                  </span>
                  <span class="alert-amount">{{ alert.amount }}</span>
                  <span class="alert-chevron" aria-hidden="true">›</span>
                </button>
              </div>
              <div class="panel-callout">
                <span class="callout-icon" aria-hidden="true">i</span>
                <span>Resolve urgent items before the 4:00 PM dispatch run.</span>
              </div>
            </SectionPanel>
          </section>

          <section class="dashboard-grid dashboard-grid-secondary">
            <SectionPanel title="Low stock" subtitle="Materials approaching reorder point" class="low-stock-panel">
              <template #action>
                <button class="text-button" type="button" @click="selectView('Materials')">Open materials <span aria-hidden="true">→</span></button>
              </template>
              <div class="stock-list">
                <div v-for="item in lowStockItems" :key="item.name" class="stock-row">
                  <span class="stock-indicator" :class="`tone-${item.tone}`" aria-hidden="true"></span>
                  <span class="stock-name">{{ item.name }}</span>
                  <span class="stock-available">{{ item.available }}</span>
                  <span class="stock-threshold">Reorder at {{ item.threshold }}</span>
                </div>
              </div>
            </SectionPanel>

            <SectionPanel title="Recent transactions" subtitle="Latest activity across the shop" class="transactions-panel">
              <template #action>
                <button class="icon-button" type="button" aria-label="More transaction actions" @click="showToast('Transaction filters are coming soon.')">•••</button>
              </template>
              <div class="transaction-list">
                <div v-for="transaction in recentTransactions" :key="`${transaction.title}-${transaction.party}`" class="transaction-row">
                  <span class="transaction-icon" :class="`tone-${transaction.tone}`" aria-hidden="true">
                    <span v-if="transaction.icon === 'arrow-down'">↓</span>
                    <span v-else>↗</span>
                  </span>
                  <span class="transaction-content">
                    <span class="transaction-title">{{ transaction.title }}</span>
                    <span class="transaction-party">{{ transaction.party }}</span>
                  </span>
                  <span class="transaction-value" :class="{ 'value-positive': transaction.amount.startsWith('+') }">{{ transaction.amount }}</span>
                  <span class="transaction-time">{{ transaction.time }}</span>
                </div>
              </div>
            </SectionPanel>
          </section>

          <section class="quick-actions" aria-labelledby="quick-actions-title">
            <div>
              <p id="quick-actions-title" class="section-kicker">Quick actions</p>
              <p class="quick-actions-help">Jump into the workflows you use most.</p>
            </div>
            <div class="quick-action-buttons">
              <button class="quick-action" type="button" @click="createNewOrder"><span class="quick-action-icon">+</span>New order</button>
              <button class="quick-action" type="button" @click="selectView('Production')"><span class="quick-action-icon">▣</span>Start production</button>
              <button class="quick-action" type="button" @click="selectView('Purchases')"><span class="quick-action-icon">↙</span>Record purchase</button>
              <button class="quick-action" type="button" @click="selectView('Accounting')"><span class="quick-action-icon">₿</span>Post expense</button>
            </div>
          </section>
        </template>

        <section v-else class="empty-workspace">
          <div class="empty-workspace-icon" aria-hidden="true">✦</div>
          <p class="eyebrow">Atropaten workspace</p>
          <h1>{{ currentView.title }}</h1>
          <p>{{ currentView.description }}</p>
          <button class="button button-secondary" type="button" @click="selectView('Dashboard')">Back to dashboard</button>
        </section>
      </main>

      <footer class="status-bar" aria-label="Workspace status">
        <span class="status-live"><span class="sync-dot" aria-hidden="true"></span> Local mode</span>
        <span class="status-divider" aria-hidden="true"></span>
        <span>Last updated just now</span>
        <span class="status-spacer"></span>
        <span>Cash <strong>$4,820</strong></span>
        <span>Bank <strong>$18,640</strong></span>
        <span>Receivable <strong>$12,680</strong></span>
        <span>Payable <strong>$6,240</strong></span>
      </footer>
    </div>

    <Transition name="toast">
      <div v-if="toastMessage" class="toast" role="status">{{ toastMessage }}</div>
    </Transition>
  </div>
</template>
