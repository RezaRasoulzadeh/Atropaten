<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InvoiceRecord } from '../../api/invoices'
import type { ShopSettingsRecord } from '../../api/reports'
import { convertRial, formatMoney, type CurrencyUnit } from '../../utils/currency'
import { formatLocalizedNumber } from '../../utils/number'
import { formatLocalizedQuantityUnits } from '../../utils/quantity'
import { formatDate } from '../../utils/date'
import logoDark from '../../assets/logo-dark.png'

const props = defineProps<{
  invoice: InvoiceRecord
  shop: ShopSettingsRecord | null
  currencyUnit: CurrencyUnit
  documentType?: 'pre' | 'final'
}>()
const { locale } = useI18n()
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
    <table class="invoice-print-table">
      <colgroup><col class="invoice-col-index" /><col class="invoice-col-description" /><col class="invoice-col-quantity" /><col class="invoice-col-price" /><col class="invoice-col-price" /></colgroup>
      <thead>
        <tr class="invoice-print-layout-row">
          <td colspan="5" class="invoice-print-layout-cell">
            <header class="invoice-print-header">
              <div class="invoice-print-brand">
                <img class="invoice-print-logo" :src="logoDark" :alt="`${shopName} logo`" />
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
                <strong class="invoice-print-number-label"><bdi>{{ invoice.invoiceNumber }}</bdi></strong>
                <span class="invoice-print-header-date">
                  {{ $t("Invoice date") }} · <bdi>{{ date(invoice.issueDate) }}</bdi>
                </span>
                <span class="invoice-print-status">{{ $ui(invoice.status) }}</span>
              </div>
            </header>
          </td>
        </tr>
      </thead>
      <tbody class="invoice-print-intro">
        <tr class="invoice-print-layout-row">
          <td colspan="5" class="invoice-print-layout-cell">
            <section class="invoice-print-title-row">
              <div class="invoice-print-customer">
                <div class="invoice-print-customer-line">
                  <p class="invoice-print-label">{{ $t("Billed to") }}</p>
                  <h2>{{ invoice.customerName?.trim() || $t("Walk-in customer") }}</h2>
                </div>
                <div v-if="invoice.customerPhone || shop?.registrationId || shop?.taxId" class="invoice-print-customer-meta">
                  <span v-if="invoice.customerPhone"><bdi>{{ invoice.customerPhone }}</bdi></span>
                  <span v-if="invoice.customerPhone && (shop?.registrationId || shop?.taxId)" aria-hidden="true">·</span>
                  <span v-if="shop?.registrationId"><span>{{ $t("Registration ID") }}</span> <bdi>{{ shop.registrationId }}</bdi></span>
                  <span v-if="shop?.registrationId && shop?.taxId" aria-hidden="true">·</span>
                  <span v-if="shop?.taxId"><span>{{ $t("Tax ID") }}</span> <bdi>{{ shop.taxId }}</bdi></span>
                </div>
              </div>
            </section>
          </td>
        </tr>
        <tr class="invoice-print-customer-gap" aria-hidden="true">
          <td colspan="5"><div></div></td>
        </tr>
        <tr class="invoice-print-columns">
          <th class="invoice-print-index" scope="col">#</th>
          <th scope="col">{{ $t("Description") }}</th>
          <th scope="col">{{ $t("Quantity") }}</th>
          <th scope="col">{{ $t("Unit price") }}</th>
          <th scope="col">{{ $t("Line total") }}</th>
        </tr>
      </tbody>
      <tbody class="invoice-print-lines">
        <tr v-for="(line, index) in invoice.items" :key="line.id" class="invoice-print-item">
          <td class="invoice-print-index">{{ formatLocalizedNumber(index + 1) }}</td>
          <td class="invoice-print-description"><strong>{{ line.description }}</strong></td>
          <td class="invoice-print-quantity">
            <bdi>{{ formatLocalizedQuantityUnits(line.quantity) }}</bdi>
            <span>{{ $ui(line.quantityUnit) }}</span>
          </td>
          <td class="invoice-print-number"><bdi>{{ amount(line.unitPriceRial) }}</bdi></td>
          <td class="invoice-print-number"><strong><bdi>{{ amount(line.lineTotalRial) }}</bdi></strong></td>
        </tr>
        <tr v-if="!invoice.items.length" class="invoice-print-item"><td colspan="5" class="invoice-print-empty">{{ $t("No invoice lines.") }}</td></tr>
      </tbody>
      <tbody class="invoice-print-summary">
        <tr class="invoice-print-layout-row">
          <td colspan="5" class="invoice-print-layout-cell">
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
          </td>
        </tr>
      </tbody>
      <tfoot aria-hidden="true">
        <tr class="invoice-print-layout-row">
          <td colspan="5" class="invoice-print-layout-cell"><div class="invoice-print-footer-reserve"></div></td>
        </tr>
      </tfoot>
    </table>
    <footer class="invoice-print-footer" :data-footer="shop?.documentFooter?.trim() || ''">
      <span v-if="shop?.documentFooter?.trim()" class="invoice-print-footer-text">{{ shop.documentFooter.trim() }}</span>
    </footer>
  </article>
</template>

<style scoped>
.invoice-print-document {
  --invoice-footer-height: 8mm;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  width: 148mm;
  max-width: 100%;
  min-height: 194mm;
  margin: 0 auto;
  padding: 8mm;
  color: #202d3a;
  background: #fff;
  color-scheme: light;
  font-family: "Estedad", sans-serif;
  font-size: 9pt;
  font-weight: 400;
  line-height: 1.7;
  font-synthesis: none;
  font-variant-numeric: tabular-nums;
  overflow-wrap: break-word;
}
.invoice-print-document *, .invoice-print-document *::before, .invoice-print-document *::after { box-sizing: border-box; }
.invoice-print-document strong, .invoice-print-document h1, .invoice-print-document h2 { font-weight: 700; }
.invoice-print-header, .invoice-print-title-row, .invoice-print-footer {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 6mm;
}
.invoice-print-header { position: relative; padding-bottom: 4.5mm; border-bottom: 0.8mm solid #243238; }
.invoice-print-header::after { position: absolute; inset-inline-start: 0; bottom: -0.8mm; width: 30mm; height: 0.8mm; content: ""; background: #f0a20b; }
.invoice-print-brand { display: flex; align-items: center; gap: 4mm; min-width: 0; flex: 1; }
.invoice-print-business { min-width: 0; }
.invoice-print-logo { display: block; flex: 0 0 18mm; width: 18mm; height: 17mm; object-fit: contain; object-position: center; }
.invoice-print-business h2 { margin: 0; font-size: 13pt; line-height: 1.4; }
.invoice-print-muted { margin: 0.7mm 0 0; color: #53626d; font-size: 7pt; }
.invoice-print-heading { flex: 0 1 43mm; min-width: 0; padding-inline-start: 4mm; border-inline-start: 0.4mm solid #d8dee0; text-align: end; }
.invoice-print-number-label { display: block; color: #243238; font-size: 11pt; }
.invoice-print-header-date { display: block; margin-top: 0.8mm; color: #53626d; font-size: 7pt; white-space: nowrap; }
.invoice-print-status { display: inline-block; margin-top: 2mm; padding: 0.6mm 3mm; border: 0.25mm solid #e0aa42; border-radius: 99mm; color: #714a00; background: #fff8e9; font-size: 7.5pt; font-weight: 600; }
.invoice-print-title-row { margin-top: 4mm; padding: 2.5mm 3mm; align-items: center; border: 0.25mm solid #ccd5d8; border-radius: 1.5mm; background: #fbfcfc; }
.invoice-print-customer { min-width: 0; flex: 1 1 auto; }
.invoice-print-customer-line { display: flex; min-width: 0; align-items: baseline; gap: 3mm; }
.invoice-print-label { margin: 0; color: #53626d; font-size: 8pt; font-weight: 600; }
.invoice-print-title-row h2 { min-width: 0; flex: 1 1 auto; margin: 0; font-size: 10pt; line-height: 1.45; overflow-wrap: break-word; }
.invoice-print-customer-meta { display: flex; flex-wrap: wrap; align-items: baseline; gap: 1mm 2mm; margin-top: 1mm; color: #53626d; font-size: 7.5pt; line-height: 1.4; }
.invoice-print-customer-meta span span { margin-inline-end: 1mm; }
.invoice-print-table { width: 100%; border-collapse: collapse; table-layout: fixed; }
.invoice-print-layout-cell { padding: 0 !important; border: 0 !important; background: transparent !important; }
.invoice-print-customer-gap td { height: 4mm; padding: 0 !important; border: 0 !important; line-height: 0; }
.invoice-print-customer-gap td div { height: 4mm; }
.invoice-col-index { width: 7%; }
.invoice-col-description { width: 37%; }
.invoice-col-quantity { width: 14%; }
.invoice-col-price { width: 21%; }
.invoice-print-columns th, .invoice-print-item td { padding: 2.7mm 2mm; border: 0.25mm solid #d2dade; text-align: start; vertical-align: top; }
.invoice-print-columns th { border-color: #243238; background: #243238; color: #fff; font-size: 7pt; font-weight: 600; text-align: center; vertical-align: middle; white-space: nowrap; }
.invoice-print-item td { font-size: 7.5pt; text-align: center; vertical-align: middle; }
.invoice-print-lines .invoice-print-item:nth-child(even) { background: #f6f8f8; }
.invoice-print-index { color: #53626d; text-align: center !important; }
.invoice-print-columns th.invoice-print-index { color: #fff; }
.invoice-print-description { text-align: start !important; vertical-align: top !important; }
.invoice-print-number { text-align: center !important; }
.invoice-print-number bdi, .invoice-print-index { white-space: nowrap; }
.invoice-print-quantity { white-space: nowrap; }
.invoice-print-quantity span { margin-inline-start: 1mm; color: #53626d; }
.invoice-print-subline { display: block; margin-top: 1mm; color: #53626d; font-size: 7.5pt; white-space: pre-wrap; }
.invoice-print-empty { padding: 8mm !important; color: #53626d; text-align: center !important; }
.invoice-print-bottom { display: flex; align-items: flex-start; gap: 6mm; margin-top: 4mm; }
.invoice-print-notes { flex: 1; min-width: 0; }
.invoice-print-note-text { margin: 2mm 0 0; white-space: pre-wrap; font-size: 8pt; orphans: 3; widows: 3; }
.invoice-print-totals { flex: 0 0 47%; min-width: 0; margin: 0; break-inside: avoid; }
.invoice-print-totals { padding: 3mm 4mm; border: 0.25mm solid #ccd5d8; border-radius: 1.5mm; }
.invoice-print-totals div { display: flex; justify-content: space-between; align-items: baseline; gap: 3mm; padding: 1.7mm 0; border-bottom: 0.25mm solid #dce2e5; }
.invoice-print-totals dt { color: #53626d; }
.invoice-print-totals dd { min-width: 0; margin: 0; font-weight: 600; text-align: end; }
.invoice-print-totals .invoice-print-total { margin: 1mm -4mm 0; padding: 2.5mm 4mm; border-block: 0; background: #243238; font-size: 10.5pt; }
.invoice-print-total dt, .invoice-print-total dd { color: #fff; font-weight: 700; }
.invoice-print-remaining dt, .invoice-print-remaining dd { color: #202d3a; font-weight: 700; }
.invoice-print-footer { width: 100%; margin-top: 4mm; padding-top: 3mm; border-top: 0.5mm solid #243238; color: #53626d; font-size: 7pt; }
.invoice-print-footer-text { display: block; flex: 1 1 auto; min-width: 0; white-space: pre-wrap; }
.invoice-print-footer-reserve { display: none; }

@media print {
  .invoice-print-document { display: block; width: 100%; max-width: none; min-height: 0; margin: 0; padding: 0; print-color-adjust: exact; -webkit-print-color-adjust: exact; }
  .invoice-print-table > thead { display: table-header-group; }
  .invoice-print-table > tfoot { display: table-footer-group; }
  .invoice-print-footer-reserve { display: block; height: var(--invoice-footer-height); }
  .invoice-print-footer {
    position: fixed;
    z-index: 1;
    inset-inline: 0;
    bottom: 0;
    width: 100%;
    height: var(--invoice-footer-height);
    margin: 0;
    padding-top: 2mm;
    background: #fff;
    line-height: 1.35;
  }
  .invoice-print-footer-text { max-height: 5mm; overflow: hidden; }
  .invoice-print-footer::before { display: block; max-height: 5mm; overflow: hidden; white-space: pre-wrap; content: attr(data-footer); }
  .invoice-print-footer-text { display: none; }
  .invoice-print-item, .invoice-print-header, .invoice-print-title-row, .invoice-print-summary, .invoice-print-footer { break-inside: avoid; }
}
@media screen and (max-width: 640px) {
  .invoice-print-document { min-height: 0; padding: 5mm; }
  .invoice-print-header { flex-wrap: wrap; }
  .invoice-print-heading { flex-basis: 100%; text-align: start; }
  .invoice-print-bottom { flex-direction: column; }
  .invoice-print-totals { width: 100%; flex-basis: auto; }
}
</style>
