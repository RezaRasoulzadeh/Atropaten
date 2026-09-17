<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InvoiceRecord } from '../../api/invoices'
import type { ShopSettingsRecord } from '../../api/reports'
import { convertRial, formatMoney, type CurrencyUnit } from '../../utils/currency'
import { formatLocalizedNumber } from '../../utils/number'
import { formatQuantityUnits } from '../../utils/quantity'
import { formatDate } from '../../utils/date'

const props = defineProps<{
  invoice: InvoiceRecord
  shop: ShopSettingsRecord | null
  currencyUnit: CurrencyUnit
  documentType?: 'pre' | 'final'
}>()
const { locale } = useI18n()
const documentTitle = computed(() =>
  (props.documentType ?? (props.invoice.status === 'Draft' ? 'pre' : 'final')) === 'pre'
    ? 'Pre-invoice' : 'Final invoice',
)
const shopName = computed(() => props.shop?.shopName?.trim() || 'Atropaten')

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}
function amount(value: number) {
  return formatLocalizedNumber(convertRial(value, props.currencyUnit), {
    maximumFractionDigits: props.currencyUnit === 'Toman' && value % 10 !== 0 ? 1 : 0,
  })
}
function date(value: string) {
  return value ? formatDate(value) : '—'
}
</script>

<template>
  <article class="invoice-print-document" :dir="locale === 'fa' ? 'rtl' : 'ltr'" :lang="locale">
    <header class="invoice-print-header">
      <div class="invoice-print-brand">
        <div class="invoice-print-mark" aria-hidden="true">{{ shopName.slice(0, 1) }}</div>
        <div class="invoice-print-business">
          <h2>{{ shopName }}</h2>
          <p v-if="shop?.shopSubtitle" class="invoice-print-muted">{{ shop.shopSubtitle }}</p>
          <p v-if="shop?.address" class="invoice-print-muted">{{ shop.address }}</p>
          <p v-if="shop?.phone || shop?.email" class="invoice-print-muted">
            <bdi v-if="shop?.phone">{{ shop.phone }}</bdi>
            <span v-if="shop?.phone && shop?.email"> · </span>
            <bdi v-if="shop?.email">{{ shop.email }}</bdi>
          </p>
          <p v-if="shop?.website" class="invoice-print-muted"><bdi>{{ shop.website }}</bdi></p>
        </div>
      </div>
      <div class="invoice-print-heading">
        <h1>{{ $ui(documentTitle) }}</h1>
        <strong><bdi>{{ invoice.invoiceNumber }}</bdi></strong>
        <span class="invoice-print-status">{{ $ui(invoice.status) }}</span>
      </div>
    </header>

    <section class="invoice-print-title-row">
      <div>
        <p class="invoice-print-label">{{ $t("Billed to") }}</p>
        <h2>{{ invoice.customerName?.trim() || $t("Walk-in customer") }}</h2>
        <p v-if="invoice.customerPhone" class="invoice-print-muted"><bdi>{{ invoice.customerPhone }}</bdi></p>
      </div>
      <div v-if="shop?.registrationId || shop?.taxId" class="invoice-print-registration">
        <p v-if="shop?.registrationId"><span>{{ $t("Registration ID") }}</span> <bdi>{{ shop.registrationId }}</bdi></p>
        <p v-if="shop?.taxId"><span>{{ $t("Tax ID") }}</span> <bdi>{{ shop.taxId }}</bdi></p>
      </div>
    </section>

    <section class="invoice-print-meta">
      <div><span>{{ $t("Invoice date") }}</span><strong>{{ date(invoice.issueDate) }}</strong></div>
      <div><span>{{ $t("Due date") }}</span><strong>{{ invoice.dueDate ? date(invoice.dueDate) : $t("On receipt") }}</strong></div>
      <div><span>{{ $t("Order reference") }}</span><strong><bdi>{{ invoice.orderId || '—' }}</bdi></strong></div>
      <div><span>{{ $t("Currency") }}</span><strong>{{ $ui(currencyUnit === 'Toman' ? 'Iranian Toman' : 'Iranian Rial') }}</strong></div>
    </section>

    <section class="invoice-print-section">
      <div class="invoice-print-section-heading">
        <h2>{{ $t("Invoice lines") }}</h2>
        <span>{{ formatLocalizedNumber(invoice.items.length) }} {{ $ui(invoice.items.length === 1 ? 'line' : 'lines') }}</span>
      </div>
      <table class="invoice-print-table">
        <colgroup><col class="invoice-col-index" /><col class="invoice-col-description" /><col class="invoice-col-quantity" /><col class="invoice-col-price" /><col class="invoice-col-price" /></colgroup>
        <thead>
          <tr>
            <th class="invoice-print-index" scope="col">#</th>
            <th scope="col">{{ $t("Description") }}</th>
            <th scope="col">{{ $t("Quantity") }}</th>
            <th class="invoice-print-number" scope="col">{{ $t("Unit price") }}</th>
            <th class="invoice-print-number" scope="col">{{ $t("Line total") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(line, index) in invoice.items" :key="line.id">
            <td class="invoice-print-index">{{ formatLocalizedNumber(index + 1) }}</td>
            <td><strong>{{ line.description }}</strong><span v-if="line.notes" class="invoice-print-subline">{{ line.notes }}</span></td>
            <td><bdi>{{ formatQuantityUnits(line.quantity) }}</bdi><span class="invoice-print-subline">{{ $ui(line.quantityUnit) }}</span></td>
            <td class="invoice-print-number"><bdi>{{ amount(line.unitPriceRial) }}</bdi></td>
            <td class="invoice-print-number"><strong><bdi>{{ amount(line.lineTotalRial) }}</bdi></strong></td>
          </tr>
          <tr v-if="!invoice.items.length"><td colspan="5" class="invoice-print-empty">{{ $t("No invoice lines.") }}</td></tr>
        </tbody>
      </table>
    </section>

    <section class="invoice-print-bottom">
      <div class="invoice-print-notes">
        <template v-if="invoice.notes || shop?.documentNotes">
          <p class="invoice-print-label">{{ $t("Notes") }}</p>
          <p v-if="invoice.notes" class="invoice-print-note-text">{{ invoice.notes }}</p>
          <p v-if="shop?.documentNotes" class="invoice-print-note-text">{{ shop.documentNotes }}</p>
        </template>
      </div>
      <dl class="invoice-print-totals">
        <div><dt>{{ $t("Subtotal") }}</dt><dd><bdi>{{ money(invoice.subtotalRial) }}</bdi></dd></div>
        <div><dt>{{ $t("Discount") }}</dt><dd><bdi>{{ money(invoice.discountRial) }}</bdi></dd></div>
        <div class="invoice-print-total"><dt>{{ $t("Total") }}</dt><dd><bdi>{{ money(invoice.totalRial) }}</bdi></dd></div>
        <div><dt>{{ $t("Paid") }}</dt><dd><bdi>{{ money(invoice.paidRial) }}</bdi></dd></div>
        <div class="invoice-print-remaining"><dt>{{ $t("Remaining") }}</dt><dd><bdi>{{ money(invoice.remainingRial) }}</bdi></dd></div>
      </dl>
    </section>

    <footer class="invoice-print-footer">
      <span>{{ shop?.documentFooter?.trim() || $t("Thank you for your business.") }}</span>
      <bdi>{{ invoice.invoiceNumber }}</bdi>
    </footer>
  </article>
</template>

<style scoped>
.invoice-print-document {
  box-sizing: border-box;
  width: 210mm;
  max-width: 100%;
  margin: 0 auto;
  padding: 12mm;
  color: #202d3a;
  background: #fff;
  color-scheme: light;
  font-family: "Estedad", sans-serif;
  font-size: 9pt;
  font-weight: 400;
  line-height: 1.7;
  font-synthesis: none;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
.invoice-print-document *, .invoice-print-document *::before, .invoice-print-document *::after { box-sizing: border-box; }
.invoice-print-document strong, .invoice-print-document h1, .invoice-print-document h2 { font-weight: 700; }
.invoice-print-header, .invoice-print-title-row, .invoice-print-section-heading, .invoice-print-footer {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 6mm;
}
.invoice-print-header { padding-bottom: 7mm; border-bottom: 0.7mm solid #234c52; }
.invoice-print-brand { display: flex; align-items: flex-start; gap: 3mm; min-width: 0; flex: 1; }
.invoice-print-business { min-width: 0; }
.invoice-print-mark { display: grid; flex: 0 0 12mm; height: 12mm; place-items: center; border: 0.4mm solid #234c52; color: #234c52; border-radius: 3mm; font-size: 20pt; font-weight: 700; line-height: 1; }
.invoice-print-business h2 { margin: 0; font-size: 15pt; line-height: 1.5; }
.invoice-print-muted { margin: 1mm 0 0; color: #53626d; font-size: 8pt; }
.invoice-print-heading { flex: 0 1 65mm; min-width: 0; text-align: end; }
.invoice-print-heading h1 { margin: 0; font-size: 20pt; line-height: 1.5; color: #234c52; }
.invoice-print-heading > strong { display: block; margin-top: 1mm; font-size: 11pt; }
.invoice-print-status { display: inline-block; margin-top: 2mm; padding: 0.5mm 3mm; border: 0.25mm solid #b5c7c7; border-radius: 1mm; color: #234c52; font-size: 8pt; font-weight: 600; }
.invoice-print-title-row { margin-top: 6mm; align-items: center; }
.invoice-print-label { margin: 0; color: #53626d; font-size: 8pt; font-weight: 600; }
.invoice-print-title-row h2 { margin: 1mm 0 0; font-size: 13pt; }
.invoice-print-registration { font-size: 8pt; text-align: end; }
.invoice-print-registration p { margin: 1mm 0; }
.invoice-print-registration span { color: #53626d; margin-inline-end: 2mm; }
.invoice-print-meta { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 4mm; margin-top: 6mm; padding: 4mm 0; border-block: 0.25mm solid #cad4d9; }
.invoice-print-meta span { display: block; color: #53626d; font-size: 7.5pt; }
.invoice-print-meta strong { display: block; margin-top: 1mm; font-size: 8pt; font-weight: 600; }
.invoice-print-section { margin-top: 7mm; }
.invoice-print-section-heading { align-items: center; margin-bottom: 3mm; break-after: avoid; }
.invoice-print-section-heading h2 { margin: 0; font-size: 10pt; }
.invoice-print-section-heading > span { color: #53626d; font-size: 8pt; }
.invoice-print-table { width: 100%; border-collapse: collapse; table-layout: fixed; }
.invoice-col-index { width: 5%; }
.invoice-col-description { width: 38%; }
.invoice-col-quantity { width: 13%; }
.invoice-col-price { width: 22%; }
.invoice-print-table th, .invoice-print-table td { padding: 3mm 2mm; border-bottom: 0.25mm solid #dce2e5; text-align: start; vertical-align: top; }
.invoice-print-table th { border-top: 0.4mm solid #234c52; border-bottom: 0.4mm solid #234c52; background: #f0f5f4; color: #234c52; font-size: 8pt; font-weight: 600; }
.invoice-print-table td { font-size: 8.5pt; }
.invoice-print-table tbody tr:nth-child(even) { background: #f8faf9; }
.invoice-print-index { color: #53626d; text-align: center !important; }
.invoice-print-number { text-align: end !important; }
.invoice-print-subline { display: block; margin-top: 1mm; color: #53626d; font-size: 7.5pt; white-space: pre-wrap; }
.invoice-print-empty { padding: 8mm !important; color: #53626d; text-align: center !important; }
.invoice-print-bottom { display: flex; align-items: flex-start; gap: 8mm; margin-top: 6mm; }
.invoice-print-notes { flex: 1; min-width: 0; }
.invoice-print-note-text { margin: 2mm 0 0; white-space: pre-wrap; font-size: 8pt; orphans: 3; widows: 3; }
.invoice-print-totals { flex: 0 0 47%; min-width: 0; margin: 0; break-inside: avoid; }
.invoice-print-totals div { display: flex; justify-content: space-between; align-items: baseline; gap: 3mm; padding: 2mm 0; border-bottom: 0.25mm solid #dce2e5; }
.invoice-print-totals dt { color: #53626d; }
.invoice-print-totals dd { min-width: 0; margin: 0; font-weight: 600; text-align: end; }
.invoice-print-totals .invoice-print-total { border-block: 0.5mm solid #234c52; margin-top: 1mm; padding-block: 3mm; font-size: 11pt; }
.invoice-print-total dt, .invoice-print-total dd { color: #234c52; font-weight: 700; }
.invoice-print-remaining dt, .invoice-print-remaining dd { color: #202d3a; font-weight: 700; }
.invoice-print-footer { margin-top: 9mm; padding-top: 3mm; border-top: 0.25mm solid #cad4d9; color: #53626d; font-size: 7.5pt; }
.invoice-print-footer > span { flex: 1; min-width: 0; }

@media print {
  .invoice-print-document { width: 100%; max-width: none; margin: 0; padding: 0; print-color-adjust: exact; -webkit-print-color-adjust: exact; }
  .invoice-print-table thead { display: table-header-group; }
  .invoice-print-table tr, .invoice-print-meta, .invoice-print-header, .invoice-print-title-row, .invoice-print-footer { break-inside: avoid; }
}
@media screen and (max-width: 640px) {
  .invoice-print-document { padding: 5mm; }
  .invoice-print-header { flex-wrap: wrap; }
  .invoice-print-heading { flex-basis: 100%; text-align: start; }
  .invoice-print-meta { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .invoice-print-bottom { flex-direction: column; }
  .invoice-print-totals { width: 100%; flex-basis: auto; }
}
</style>
