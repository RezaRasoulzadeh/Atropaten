<script setup lang="ts">
import type { InvoiceRecord } from '../../api/invoices'
import type { ShopSettingsRecord } from '../../api/reports'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatQuantityUnits } from '../../utils/quantity'
import { formatDateTime } from '../../utils/date'

const props = defineProps<{
  invoice: InvoiceRecord
  shop: ShopSettingsRecord | null
  currencyUnit: CurrencyUnit
  documentType?: 'pre' | 'final'
}>()

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}

function shopValue(value: string | undefined, fallback = '—') {
  return value?.trim() || fallback
}
</script>

<template>
  <article class="invoice-print-document">
    <header class="invoice-print-header">
      <div class="invoice-print-brand">
        <div class="invoice-print-mark">A</div>
        <div>
          <p class="invoice-print-kicker">{{ shopValue(shop?.shopName, 'Atropaten') }}</p>
          <h1>{{ shopValue(shop?.shopSubtitle, 'Commercial invoice') }}</h1>
          <p v-if="shop?.address" class="invoice-print-muted">{{ shop.address }}</p>
          <p v-if="shop?.phone || shop?.email" class="invoice-print-muted">
            {{ [shop?.phone, shop?.email].filter(Boolean).join(' · ') }}
          </p>
        </div>
      </div>
      <div class="invoice-print-heading">
        <p class="invoice-print-kicker">{{ props.documentType === 'pre' ? 'Pre-invoice' : 'Final invoice' }}</p>
        <strong>{{ invoice.invoiceNumber }}</strong>
        <span>{{ formatDateTime(invoice.issueDate) }}</span>
      </div>
    </header>

    <section class="invoice-print-title-row">
      <div>
        <p class="invoice-print-label">Billed to</p>
        <h2>{{ shopValue(invoice.customerName, 'Walk-in customer') }}</h2>
        <p v-if="invoice.customerPhone" class="invoice-print-muted">{{ invoice.customerPhone }}</p>
      </div>
      <div class="invoice-print-status">
        <span class="invoice-print-label">Status</span>
        <strong>{{ invoice.status }}</strong>
      </div>
    </section>

    <section class="invoice-print-meta">
      <div><span>Invoice date</span><strong>{{ formatDateTime(invoice.issueDate) }}</strong></div>
      <div><span>Due date</span><strong>{{ invoice.dueDate ? formatDateTime(invoice.dueDate) : 'On receipt' }}</strong></div>
      <div><span>Order reference</span><strong>{{ shopValue(invoice.orderId) }}</strong></div>
      <div><span>Currency</span><strong>{{ props.currencyUnit === 'Toman' ? 'Iranian Toman' : 'Iranian Rial' }}</strong></div>
    </section>

    <section class="invoice-print-section">
      <div class="invoice-print-section-heading">
        <div>
          <p class="invoice-print-kicker">Commercial snapshot</p>
          <h2>Invoice lines</h2>
        </div>
        <span>{{ invoice.items.length }} {{ invoice.items.length === 1 ? 'line' : 'lines' }}</span>
      </div>
      <table class="invoice-print-table">
        <thead>
          <tr>
            <th class="invoice-print-index">#</th>
            <th>Description</th>
            <th>Quantity</th>
            <th class="invoice-print-number">Unit price</th>
            <th class="invoice-print-number">Line total</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(line, index) in invoice.items" :key="line.id">
            <td class="invoice-print-index">{{ index + 1 }}</td>
            <td>
              <strong>{{ line.description }}</strong>
              <span v-if="line.notes" class="invoice-print-subline">{{ line.notes }}</span>
            </td>
            <td>{{ formatQuantityUnits(line.quantity) }} {{ line.quantityUnit }}</td>
            <td class="invoice-print-number">{{ money(line.unitPriceRial) }}</td>
            <td class="invoice-print-number"><strong>{{ money(line.lineTotalRial) }}</strong></td>
          </tr>
          <tr v-if="!invoice.items.length">
            <td colspan="5" class="invoice-print-empty">No invoice lines.</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="invoice-print-bottom">
      <div v-if="invoice.notes || shop?.documentNotes" class="invoice-print-notes">
        <p class="invoice-print-label">Notes</p>
        <p v-if="invoice.notes" class="invoice-print-note-text">{{ invoice.notes }}</p>
        <p v-if="shop?.documentNotes" class="invoice-print-note-text">{{ shop.documentNotes }}</p>
      </div>
      <dl class="invoice-print-totals">
        <div><dt>Subtotal</dt><dd>{{ money(invoice.subtotalRial) }}</dd></div>
        <div><dt>Discount</dt><dd>{{ money(invoice.discountRial) }}</dd></div>
        <div class="invoice-print-total"><dt>Total</dt><dd>{{ money(invoice.totalRial) }}</dd></div>
        <div class="invoice-print-paid"><dt>Paid</dt><dd>{{ money(invoice.paidRial) }}</dd></div>
        <div class="invoice-print-remaining"><dt>Remaining</dt><dd>{{ money(invoice.remainingRial) }}</dd></div>
      </dl>
    </section>

    <footer class="invoice-print-footer">
      <span>{{ shopValue(shop?.documentFooter, 'Thank you for your business.') }}</span>
      <span>{{ invoice.invoiceNumber }} · {{ invoice.status }}</span>
    </footer>
  </article>
</template>

<style scoped>
.invoice-print-document {
  box-sizing: border-box;
  width: 210mm;
  max-width: 100%;
  min-height: 280mm;
  margin: 0 auto;
  padding: 14mm;
  color: #172033;
  background: #fff;
  font-family: Arial, sans-serif;
  font-size: 10pt;
  line-height: 1.45;
}

.invoice-print-header,
.invoice-print-title-row,
.invoice-print-section-heading,
.invoice-print-bottom,
.invoice-print-footer {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1.25rem;
}

.invoice-print-header {
  padding-bottom: 1.25rem;
  border-bottom: 2px solid #172033;
}

.invoice-print-brand { display: flex; align-items: center; gap: 0.75rem; min-width: 0; }
.invoice-print-mark { display: grid; width: 2.75rem; height: 2.75rem; place-items: center; color: #fff; background: #089fbd; border-radius: 0.45rem; font-size: 1.3rem; font-weight: 700; }
.invoice-print-kicker { margin: 0; color: #0785a0; font-size: 8pt; font-weight: 700; letter-spacing: 0.12em; text-transform: uppercase; }
.invoice-print-brand h1 { margin: 0.1rem 0 0; font-size: 15pt; }
.invoice-print-heading { text-align: right; }
.invoice-print-heading strong { display: block; margin-top: 0.1rem; font-size: 18pt; letter-spacing: 0.04em; }
.invoice-print-heading span { display: block; margin-top: 0.2rem; color: #637083; font-size: 8.5pt; }
.invoice-print-muted { margin: 0.2rem 0 0; color: #637083; font-size: 8.5pt; }
.invoice-print-title-row { margin-top: 1.25rem; align-items: end; }
.invoice-print-title-row h2 { margin: 0.15rem 0 0; font-size: 14pt; }
.invoice-print-label { margin: 0; color: #637083; font-size: 8pt; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; }
.invoice-print-status { min-width: 7rem; padding: 0.55rem 0.75rem; border: 1px solid #b7c4cf; border-radius: 0.35rem; text-align: right; }
.invoice-print-status strong { display: block; margin-top: 0.15rem; font-size: 10pt; }
.invoice-print-meta { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.75rem; margin-top: 1.25rem; padding: 0.9rem 1rem; border: 1px solid #d7dee5; border-radius: 0.4rem; background: #f5f8fa; }
.invoice-print-meta div { min-width: 0; }
.invoice-print-meta span { display: block; color: #637083; font-size: 8pt; }
.invoice-print-meta strong { display: block; margin-top: 0.2rem; font-size: 9pt; overflow-wrap: anywhere; }
.invoice-print-section { margin-top: 1.5rem; }
.invoice-print-section-heading { align-items: end; padding-bottom: 0.6rem; border-bottom: 1px solid #b7c4cf; }
.invoice-print-section-heading h2 { margin: 0.1rem 0 0; font-size: 12pt; }
.invoice-print-section-heading > span { color: #637083; font-size: 8.5pt; }
.invoice-print-table { width: 100%; border-collapse: collapse; table-layout: fixed; }
.invoice-print-table th, .invoice-print-table td { padding: 0.65rem 0.5rem; border-bottom: 1px solid #d7dee5; text-align: left; vertical-align: top; }
.invoice-print-table th { color: #465467; background: #eef3f6; font-size: 8.5pt; font-weight: 700; }
.invoice-print-table td { font-size: 9pt; }
.invoice-print-table th:nth-child(1), .invoice-print-table td:nth-child(1) { width: 2.5rem; }
.invoice-print-table th:nth-child(2), .invoice-print-table td:nth-child(2) { width: 40%; }
.invoice-print-table th:nth-child(3), .invoice-print-table td:nth-child(3) { width: 16%; }
.invoice-print-table th:nth-child(4), .invoice-print-table td:nth-child(4), .invoice-print-table th:nth-child(5), .invoice-print-table td:nth-child(5) { width: 19%; }
.invoice-print-index { color: #637083; text-align: center !important; }
.invoice-print-number { text-align: right !important; white-space: nowrap; }
.invoice-print-subline { display: block; margin-top: 0.2rem; color: #637083; font-size: 8pt; white-space: pre-wrap; }
.invoice-print-empty { padding: 1.5rem !important; color: #637083; text-align: center !important; }
.invoice-print-bottom { align-items: flex-start; margin-top: 1.5rem; }
.invoice-print-notes { flex: 1 1 auto; max-width: 52%; }
.invoice-print-note-text { margin: 0.35rem 0 0; white-space: pre-wrap; }
.invoice-print-totals { width: 46%; margin: 0; }
.invoice-print-totals div { display: flex; justify-content: space-between; gap: 1rem; padding: 0.45rem 0; border-bottom: 1px solid #d7dee5; }
.invoice-print-totals dt { color: #637083; }
.invoice-print-totals dd { margin: 0; font-weight: 600; text-align: right; white-space: nowrap; }
.invoice-print-totals .invoice-print-total { margin-top: 0.25rem; padding: 0.7rem 0; border-top: 2px solid #172033; border-bottom: 2px solid #172033; font-size: 12pt; }
.invoice-print-totals .invoice-print-total dt, .invoice-print-totals .invoice-print-total dd { color: #0785a0; font-weight: 700; }
.invoice-print-paid dd { color: #078653; }
.invoice-print-remaining dd { color: #b77900; }
.invoice-print-footer { margin-top: 2.5rem; padding-top: 0.75rem; border-top: 1px solid #b7c4cf; color: #637083; font-size: 8pt; }

@media print {
  .invoice-print-document { width: auto; max-width: none; min-height: 0; margin: 0; padding: 0; font-size: 9.5pt; }
  .invoice-print-table thead { display: table-header-group; }
  .invoice-print-table tr, .invoice-print-meta, .invoice-print-bottom { break-inside: avoid; }
  .invoice-print-footer { break-inside: avoid; }
}

@media (max-width: 640px) {
  .invoice-print-meta { grid-template-columns: repeat(2, 1fr); }
  .invoice-print-bottom { flex-direction: column; }
  .invoice-print-notes, .invoice-print-totals { width: 100%; max-width: none; }
}
</style>
