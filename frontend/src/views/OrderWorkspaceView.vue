<script setup lang="ts">
import { computed, ref } from 'vue'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceTabs from '../components/WorkspaceTabs.vue'
import type { CommercialStatus, FulfillmentStatus, Order, OrderItem, PaymentStatus, Priority, ProductionStatus } from '../data/orders'
import { formatMoney, type CurrencyUnit } from '../utils/currency'
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  ChevronDown,
  CircleAlert,
  CreditCard,
  Eye,
  EyeOff,
  FileText,
  History as HistoryIcon,
  Info,
  Minus,
  MoreHorizontal,
  PackageOpen,
  Plus,
  Upload,
  X,
} from 'lucide-vue-next'

type BadgeTone = 'blue' | 'green' | 'amber' | 'red' | 'slate'
type DraftItem = { id: string; service: string; specification: string; quantity: number; unit: string }
type HistoryEvent = { date: string; title: string; detail: string; amountRial?: number; tone: BadgeTone }

const props = withDefaults(defineProps<{ order: Order; isDraft?: boolean; currencyUnit: CurrencyUnit }>(), { isDraft: false })

const emit = defineEmits<{
  back: []
  notify: [message: string]
}>()

const tabs = ['Overview', 'Items', 'Production', 'Payments', 'Files', 'History']
const activeTab = ref('Overview')
const showMargins = ref(true)
const draftCustomer = ref('')
const draftPromisedDate = ref('')
const draftPriority = ref<Priority>('Normal')
const draftNotes = ref('')
const draftItems = ref<DraftItem[]>([
  { id: 'draft-item-1', service: 'Business cards', specification: 'Choose stock, size, color, and finishing', quantity: 500, unit: 'pcs' },
])

const orderBalance = computed(() => props.order.totalRial - props.order.paidRial)
const orderProfit = computed(() => props.order.items.reduce((total, item) => total + item.sellingPriceRial - item.estimatedCostRial, 0))
const orderMargin = computed(() => props.order.totalRial ? Math.round((orderProfit.value / props.order.totalRial) * 100) : 0)

const payments = [
  { date: '12 Aug 2024 · 09:04', method: 'Bank transfer', reference: 'TRX-882104', amountRial: 46_000_000, status: 'Allocated' },
  { date: 'Deposit requested', method: 'Pending', reference: '—', amountRial: 0, status: 'Not received' },
]

const files = [
  { name: 'mehr-studio-business-cards.pdf', detail: 'PDF · 4.8 MB · added 12 Aug', kind: 'Proof' },
  { name: 'menu-final-artwork.zip', detail: 'ZIP · 18.2 MB · added 11 Aug', kind: 'Artwork' },
]

const history: HistoryEvent[] = [
  { date: '12 Aug · 10:22', title: 'Production started', detail: 'Business cards moved to the digital press queue.', tone: 'blue' as BadgeTone },
  { date: '12 Aug · 09:04', title: 'Deposit recorded', detail: 'Bank transfer received and allocated.', amountRial: 46_000_000, tone: 'green' as BadgeTone },
  { date: '11 Aug · 16:40', title: 'Order confirmed', detail: 'Commercial status changed from Quoted to Confirmed.', tone: 'blue' as BadgeTone },
  { date: '11 Aug · 14:18', title: 'Proof approved', detail: 'Mehr Studio approved the first proof by email.', tone: 'green' as BadgeTone },
]

function notify(message: string) {
  emit('notify', message)
}

function currency(amountRial: number) {
  return formatMoney(amountRial, props.currencyUnit)
}

function valueOrDash(amountRial: number) {
  return amountRial ? currency(amountRial) : '—'
}

function commercialTone(status: CommercialStatus): BadgeTone {
  return status === 'Confirmed' ? 'blue' : status === 'Closed' ? 'green' : status === 'Quoted' ? 'amber' : 'slate'
}

function fulfillmentTone(status: FulfillmentStatus): BadgeTone {
  return status === 'In production' ? 'blue' : status === 'Ready' || status === 'Delivered' ? 'green' : 'slate'
}

function paymentTone(status: PaymentStatus): BadgeTone {
  return status === 'Paid' ? 'green' : status === 'Partially paid' ? 'amber' : 'red'
}

function priorityTone(priority: Priority): BadgeTone {
  return priority === 'Urgent' ? 'red' : priority === 'High' ? 'amber' : priority === 'Normal' ? 'blue' : 'slate'
}

function productionTone(status: ProductionStatus): BadgeTone {
  return status === 'In progress' ? 'blue' : status === 'Ready to print' || status === 'Delivered' ? 'green' : status === 'Waiting for proof' ? 'amber' : 'slate'
}

function itemProfit(item: OrderItem) {
  return item.sellingPriceRial - item.estimatedCostRial
}

function itemMargin(item: OrderItem) {
  return item.sellingPriceRial ? Math.round((itemProfit(item) / item.sellingPriceRial) * 100) : 0
}

function addDraftItem() {
  draftItems.value.push({ id: `draft-item-${draftItems.value.length + 1}`, service: 'New service', specification: 'Choose service parameters', quantity: 1, unit: 'pcs' })
}

function removeDraftItem(id: string) {
  draftItems.value = draftItems.value.filter((item) => item.id !== id)
}

function saveDraft() {
  notify('Draft order kept in this workspace for this visual prototype.')
}

function formatDraftCustomer() {
  return draftCustomer.value || 'Select customer'
}

function formatDraftPromised() {
  return draftPromisedDate.value || 'Set promised date'
}
</script>

<template>
  <div class="order-workspace">
    <div class="order-sticky-controls">
      <header class="order-workspace-header">
      <button class="back-button" type="button" @click="$emit('back')"><ArrowLeft :size="16" :stroke-width="1.8" aria-hidden="true" /> Orders</button>
      <div class="order-heading-copy">
        <p class="eyebrow">{{ isDraft ? 'Sales / New order' : 'Sales / Order workspace' }}</p>
        <div class="order-title-row">
          <h1>{{ isDraft ? 'New order draft' : order.id }}</h1>
          <span v-if="isDraft" class="draft-label">Unsaved</span>
        </div>
        <p class="heading-description">{{ isDraft ? 'Build the order composition before it is saved.' : `${order.customer} · ${order.items.length} line items` }}</p>
      </div>
      <div class="order-header-actions">
        <button class="button button-secondary" type="button" @click="notify('Order actions will be connected in a later milestone.')">More <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" /></button>
        <button class="button button-primary" type="button" @click="isDraft ? saveDraft() : notify('Editing is available when order persistence is introduced.')">{{ isDraft ? 'Save draft' : 'Edit order' }}</button>
      </div>
      </header>

      <section class="order-meta-strip" aria-label="Order summary">
      <div class="order-meta-cell order-customer-meta">
        <span class="meta-label">Customer</span>
        <span class="meta-value">{{ isDraft ? formatDraftCustomer() : order.customer }}</span>
        <span class="meta-detail">{{ isDraft ? 'Required to continue' : order.customerDetail.split(' · ')[0] }}</span>
      </div>
      <div class="order-meta-cell">
        <span class="meta-label">Created</span>
        <span class="meta-value">{{ order.created }}</span>
        <span class="meta-detail">Local time</span>
      </div>
      <div class="order-meta-cell">
        <span class="meta-label">Promised date</span>
        <span class="meta-value" :class="{ 'text-danger': !isDraft && order.promised.includes('Overdue') }">{{ isDraft ? formatDraftPromised() : order.promised }}</span>
        <span class="meta-detail">{{ isDraft ? 'Choose in Overview' : 'Delivery commitment' }}</span>
      </div>
      <div class="order-meta-cell">
        <span class="meta-label">Priority</span>
        <StatusBadge :label="isDraft ? draftPriority : order.priority" :tone="priorityTone(isDraft ? draftPriority : order.priority)" />
      </div>
      <div class="order-meta-statuses">
        <span class="meta-label">Order states</span>
        <div class="status-axis-list">
          <StatusBadge :label="order.commercialStatus" :tone="commercialTone(order.commercialStatus)" />
          <StatusBadge :label="order.fulfillmentStatus" :tone="fulfillmentTone(order.fulfillmentStatus)" />
          <StatusBadge :label="order.paymentStatus" :tone="paymentTone(order.paymentStatus)" />
        </div>
      </div>
      </section>

      <WorkspaceTabs :tabs="tabs" :active-tab="activeTab" @change="activeTab = $event" />
    </div>

    <section v-if="activeTab === 'Overview'" class="order-tab-content">
      <div class="order-overview-grid">
        <SectionPanel title="Order details" :subtitle="isDraft ? 'Set the basics before adding services.' : 'Customer and delivery context'" class="order-details-panel">
          <div v-if="isDraft" class="form-grid">
            <label class="form-field form-field-wide">
              <span>Customer</span>
              <select v-model="draftCustomer">
                <option value="">Select customer</option>
                <option>Mehr Studio</option>
                <option>Arman Foods</option>
                <option>Nika Events</option>
                <option>Pendar Clinic</option>
              </select>
            </label>
            <label class="form-field">
              <span>Promised date</span>
              <input v-model="draftPromisedDate" type="date" />
            </label>
            <label class="form-field">
              <span>Priority</span>
              <select v-model="draftPriority">
                <option>Urgent</option>
                <option>High</option>
                <option>Normal</option>
                <option>Low</option>
              </select>
            </label>
            <label class="form-field form-field-wide">
              <span>Notes</span>
              <textarea v-model="draftNotes" rows="3" placeholder="Add a production or delivery note"></textarea>
            </label>
          </div>
          <div v-else class="detail-grid">
            <div><span class="detail-label">Customer</span><strong>{{ order.customer }}</strong><span class="detail-muted">{{ order.customerDetail }}</span></div>
            <div><span class="detail-label">Created</span><strong>{{ order.created }}</strong><span class="detail-muted">Order intake</span></div>
            <div><span class="detail-label">Promised date</span><strong :class="{ 'text-danger': order.promised.includes('Overdue') }">{{ order.promised }}</strong><span class="detail-muted">Customer commitment</span></div>
            <div><span class="detail-label">Priority</span><StatusBadge :label="order.priority" :tone="priorityTone(order.priority)" /></div>
            <div class="detail-wide"><span class="detail-label">Notes</span><p class="detail-note">{{ order.notes }}</p></div>
          </div>
        </SectionPanel>

        <SectionPanel title="Order summary" subtitle="Commercial totals and payment position" class="order-summary-panel">
          <div class="money-summary">
            <div><span class="money-label">Total</span><strong>{{ isDraft ? 'To be priced' : currency(order.totalRial) }}</strong></div>
            <div><span class="money-label">Paid</span><strong class="value-positive">{{ isDraft ? '—' : currency(order.paidRial) }}</strong></div>
            <div><span class="money-label">Remaining</span><strong class="value-danger">{{ isDraft ? '—' : currency(orderBalance) }}</strong></div>
          </div>
          <div v-if="!isDraft" class="payment-progress-block">
            <div class="payment-progress-label"><span>Paid to date</span><strong>{{ Math.round((order.paidRial / order.totalRial) * 100) }}%</strong></div>
            <div class="payment-progress"><span :style="{ width: `${(order.paidRial / order.totalRial) * 100}%` }"></span></div>
          </div>
          <div v-else class="draft-summary-note"><span class="callout-icon" aria-hidden="true"><Info :size="15" :stroke-width="1.8" /></span><span>Prices will be calculated from the configured services in a later milestone.</span></div>
        </SectionPanel>

        <SectionPanel title="Status axes" subtitle="Tracked independently for operational clarity" class="status-axes-panel">
          <div class="status-axis-detail"><span><strong>Commercial</strong><small>Quote and confirmation</small></span><StatusBadge :label="order.commercialStatus" :tone="commercialTone(order.commercialStatus)" /></div>
          <div class="status-axis-detail"><span><strong>Fulfillment</strong><small>Production and delivery</small></span><StatusBadge :label="order.fulfillmentStatus" :tone="fulfillmentTone(order.fulfillmentStatus)" /></div>
          <div class="status-axis-detail"><span><strong>Payment</strong><small>Derived from payments</small></span><StatusBadge :label="order.paymentStatus" :tone="paymentTone(order.paymentStatus)" /></div>
        </SectionPanel>
      </div>

      <SectionPanel title="Order items" :subtitle="isDraft ? 'Add services from the Items tab.' : `${order.items.length} unrelated services in this order`" class="overview-items-panel">
        <div v-if="isDraft" class="overview-draft-items"><span class="empty-line-icon" aria-hidden="true"><Plus :size="15" :stroke-width="1.8" /></span><span>No service lines yet.</span><button class="text-button" type="button" @click="activeTab = 'Items'">Add a service <ArrowRight :size="14" :stroke-width="1.8" aria-hidden="true" /></button></div>
        <div v-else class="overview-item-list">
          <div v-for="item in order.items" :key="item.id" class="overview-item-row"><span class="item-index">{{ order.items.indexOf(item) + 1 }}</span><span class="overview-item-copy"><strong>{{ item.service }}</strong><small>{{ item.specification }}</small></span><span class="overview-item-qty">{{ item.quantity.toLocaleString() }} {{ item.unit }}</span><span class="overview-item-price">{{ currency(item.sellingPriceRial) }}</span><StatusBadge :label="item.productionStatus" :tone="productionTone(item.productionStatus)" /></div>
        </div>
      </SectionPanel>
    </section>

    <section v-else-if="activeTab === 'Items'" class="order-tab-content">
      <SectionPanel title="Order items" :subtitle="isDraft ? 'Compose unrelated services as separate line items.' : 'Each line keeps its own specification, cost, and production state.'" class="items-detail-panel">
        <template #action>
          <label class="margin-toggle"><input v-model="showMargins" type="checkbox" /> <Eye v-if="showMargins" :size="14" :stroke-width="1.8" aria-hidden="true" /><EyeOff v-else :size="14" :stroke-width="1.8" aria-hidden="true" /> Show cost & margin</label>
        </template>
        <div v-if="isDraft" class="draft-items-list">
          <div v-for="(item, index) in draftItems" :key="item.id" class="draft-item-row">
            <div class="draft-item-number">{{ String(index + 1).padStart(2, '0') }}</div>
            <label class="form-field"><span>Service</span><select v-model="item.service"><option>Business cards</option><option>A3 menu printing</option><option>Event banners</option><option>Artwork setup</option><option>New service</option></select></label>
            <label class="form-field draft-quantity"><span>Quantity</span><input v-model.number="item.quantity" min="1" type="number" /></label>
            <label class="form-field draft-unit"><span>Unit</span><select v-model="item.unit"><option>pcs</option><option>sheets</option><option>layouts</option><option>boards</option></select></label>
            <label class="form-field draft-specification"><span>Specification</span><input v-model="item.specification" type="text" /></label>
            <button v-if="draftItems.length > 1" class="remove-item-button" type="button" aria-label="Remove item" @click="removeDraftItem(item.id)"><X :size="15" :stroke-width="1.8" aria-hidden="true" /></button>
          </div>
          <div class="draft-items-footer"><button class="button button-secondary" type="button" @click="addDraftItem"><Plus class="button-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />Add service line</button><span class="draft-items-hint">A draft can contain multiple unrelated services.</span></div>
        </div>
        <div v-else class="item-cards">
          <article v-for="(item, index) in order.items" :key="item.id" class="item-card">
            <div class="item-card-top"><div class="item-card-title"><span class="item-index">{{ String(index + 1).padStart(2, '0') }}</span><div><h3>{{ item.service }}</h3><p>{{ item.specification }}</p></div></div><StatusBadge :label="item.productionStatus" :tone="productionTone(item.productionStatus)" /></div>
            <div class="item-card-metrics"><div><span>Quantity</span><strong>{{ item.quantity.toLocaleString() }} {{ item.unit }}</strong></div><div><span>Selling price</span><strong>{{ currency(item.sellingPriceRial) }}</strong></div><div v-if="showMargins"><span>Estimated cost</span><strong>{{ currency(item.estimatedCostRial) }}</strong></div><div v-if="showMargins"><span>Profit / margin</span><strong class="value-positive">{{ currency(itemProfit(item)) }} · {{ itemMargin(item) }}%</strong></div><button class="text-button" type="button" @click="notify(`Opening ${item.service} detail is available in the service configurator milestone.`)">Configure <ArrowRight :size="14" :stroke-width="1.8" aria-hidden="true" /></button></div>
          </article>
        </div>
      </SectionPanel>
      <SectionPanel v-if="isDraft" title="Draft totals" subtitle="Frontend-only placeholders until pricing is connected." class="draft-totals-panel">
        <div class="draft-total-row"><span>{{ draftItems.length }} service lines</span><strong>To be priced</strong></div>
        <div class="draft-total-row"><span>Deposit</span><strong>Not set</strong></div>
        <div class="draft-total-row total-row-emphasis"><span>Estimated total</span><strong>—</strong></div>
      </SectionPanel>
      <SectionPanel v-else title="Order profitability" subtitle="Cost visibility can be hidden when a customer is viewing the screen." class="profitability-panel">
        <div class="profitability-summary"><div><span>Estimated cost</span><strong>{{ currency(order.totalRial - orderProfit) }}</strong></div><div><span>Profit</span><strong class="value-positive">{{ currency(orderProfit) }}</strong></div><div><span>Margin</span><strong class="value-positive">{{ orderMargin }}%</strong></div><button class="button button-secondary" type="button" @click="showMargins = !showMargins">{{ showMargins ? 'Hide margin' : 'Show margin' }}</button></div>
      </SectionPanel>
    </section>

    <section v-else-if="activeTab === 'Production'" class="order-tab-content">
      <SectionPanel title="Production plan" :subtitle="isDraft ? 'Production jobs appear after services are configured.' : 'Work sequencing follows each item promised date.'" class="production-workspace-panel">
        <div v-if="isDraft" class="tab-empty-state"><div class="empty-workspace-icon" aria-hidden="true"><PackageOpen :size="21" :stroke-width="1.8" /></div><h2>No production jobs yet</h2><p>Add and configure service lines first. The production queue will be created per item.</p><button class="button button-secondary" type="button" @click="activeTab = 'Items'">Go to items</button></div>
        <div v-else class="production-job-list"><div v-for="(item, index) in order.items" :key="item.id" class="production-job-row"><div class="item-index">{{ String(index + 1).padStart(2, '0') }}</div><div class="production-job-copy"><strong>{{ item.service }}</strong><span>{{ item.quantity.toLocaleString() }} {{ item.unit }} · {{ order.promised }}</span></div><div class="production-progress"><div class="production-progress-label"><span>{{ item.productionStatus }}</span><strong>{{ item.productionStatus === 'Delivered' ? '100' : item.productionStatus === 'In progress' ? '72' : item.productionStatus === 'Ready to print' ? '100' : '18' }}%</strong></div><div class="progress-track"><span :style="{ width: item.productionStatus === 'Delivered' || item.productionStatus === 'Ready to print' ? '100%' : item.productionStatus === 'In progress' ? '72%' : '18%' }"></span></div></div><StatusBadge :label="item.productionStatus" :tone="productionTone(item.productionStatus)" /><button class="row-open-button" type="button" aria-label="View production job" @click="notify(`Production detail for ${item.service} is a future workflow.`)"><ArrowRight :size="16" :stroke-width="1.8" aria-hidden="true" /></button></div></div>
      </SectionPanel>
    </section>

    <section v-else-if="activeTab === 'Payments'" class="order-tab-content">
      <div class="payment-tab-grid">
        <SectionPanel title="Payment position" subtitle="Deposits and payments against this order" class="payment-position-panel">
          <div class="money-summary"><div><span class="money-label">Order total</span><strong>{{ isDraft ? 'To be priced' : currency(order.totalRial) }}</strong></div><div><span class="money-label">Paid</span><strong class="value-positive">{{ isDraft ? '—' : currency(order.paidRial) }}</strong></div><div><span class="money-label">Remaining</span><strong class="value-danger">{{ isDraft ? '—' : currency(orderBalance) }}</strong></div></div>
          <button class="button button-primary" type="button" @click="notify(isDraft ? 'Save the draft before recording a payment.' : 'Payment recording will be connected in the accounting milestone.')"><CreditCard :size="16" :stroke-width="1.8" aria-hidden="true" />Record payment</button>
        </SectionPanel>
        <SectionPanel title="Payment state" subtitle="Separate from commercial and fulfillment states" class="payment-state-panel"><div class="large-status"><StatusBadge :label="order.paymentStatus" :tone="paymentTone(order.paymentStatus)" /><p>{{ isDraft ? 'No payment has been recorded.' : order.paymentStatus === 'Paid' ? 'This order is fully settled.' : `${currency(orderBalance)} remains outstanding.` }}</p></div></SectionPanel>
      </div>
      <SectionPanel title="Payment activity" subtitle="Mock transaction history for the order" class="payment-history-panel"><div v-if="isDraft" class="tab-empty-inline"><span class="empty-line-icon" aria-hidden="true"><Minus :size="15" :stroke-width="1.8" /></span><span>No payment activity in a new draft.</span></div><div v-else class="payment-table-wrap"><table class="data-table payment-table"><thead><tr><th>Date</th><th>Method</th><th>Reference</th><th>Amount</th><th>Status</th></tr></thead><tbody><tr v-for="payment in payments" :key="`${payment.date}-${payment.method}`"><td>{{ payment.date }}</td><td class="table-primary">{{ payment.method }}</td><td>{{ payment.reference }}</td><td class="table-money">{{ valueOrDash(payment.amountRial) }}</td><td><StatusBadge :label="payment.status" :tone="payment.status === 'Allocated' ? 'green' : 'slate'" /></td></tr></tbody></table></div></SectionPanel>
    </section>

    <section v-else-if="activeTab === 'Files'" class="order-tab-content">
      <SectionPanel title="Order files" :subtitle="isDraft ? 'Attach artwork and proofs after the order is saved.' : 'Artwork, proofs, and customer-provided files'" class="files-panel">
        <div class="file-dropzone" :class="{ 'is-disabled': isDraft }"><span class="file-dropzone-icon" aria-hidden="true"><Upload :size="17" :stroke-width="1.8" /></span><div><strong>{{ isDraft ? 'File attachments are available after saving' : 'Drop files here or choose from your computer' }}</strong><p>PDF, ZIP, PNG, JPG up to 50 MB</p></div><button class="button button-secondary" type="button" :disabled="isDraft" @click="notify('A system file picker will be connected in a later milestone.')"><Upload :size="15" :stroke-width="1.8" aria-hidden="true" />Choose files</button></div>
        <div v-if="isDraft" class="tab-empty-inline"><span class="empty-line-icon" aria-hidden="true"><Minus :size="15" :stroke-width="1.8" /></span><span>No files attached to this draft.</span></div>
        <div v-else class="file-list"><div v-for="file in files" :key="file.name" class="file-row"><span class="file-type"><FileText :size="15" :stroke-width="1.8" aria-hidden="true" />{{ file.kind }}</span><span class="file-copy"><strong>{{ file.name }}</strong><small>{{ file.detail }}</small></span><button class="text-button" type="button" @click="notify(`Preview for ${file.name} is not connected yet.`)">Preview <ArrowRight :size="14" :stroke-width="1.8" aria-hidden="true" /></button><button class="icon-button" type="button" :aria-label="`More actions for ${file.name}`" @click="notify('File actions are coming soon.')"><MoreHorizontal :size="17" :stroke-width="1.8" aria-hidden="true" /></button></div></div>
      </SectionPanel>
    </section>

    <section v-else class="order-tab-content">
      <SectionPanel title="Order history" subtitle="A chronological view of important changes" class="history-panel"><div v-if="isDraft" class="tab-empty-state"><div class="empty-workspace-icon" aria-hidden="true"><HistoryIcon :size="21" :stroke-width="1.8" /></div><h2>No history yet</h2><p>Changes will appear here after the draft is saved.</p></div><div v-else class="history-timeline"><div v-for="event in history" :key="event.date" class="history-event"><span class="history-marker" :class="`tone-${event.tone}`" aria-hidden="true"><CheckCircle2 v-if="event.tone === 'green'" :size="15" :stroke-width="1.8" /><CircleAlert v-else :size="15" :stroke-width="1.8" /></span><div><span class="history-date">{{ event.date }}</span><h3>{{ event.title }}</h3><p>{{ event.detail }}<span v-if="event.amountRial"> {{ currency(event.amountRial) }}</span></p></div></div></div></SectionPanel>
    </section>
  </div>
</template>
