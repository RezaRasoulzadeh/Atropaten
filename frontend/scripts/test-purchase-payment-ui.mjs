import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { fileURLToPath } from 'node:url'
import { build } from 'esbuild'
import { parse, compileScript } from '@vue/compiler-sfc'

// Render the real wizard and composable without touching the desktop database.
const root = fileURLToPath(new URL('../', import.meta.url))
globalThis.window = { go: { main: { App: {
  ListPayments: async () => [], ListChecks: async () => [],
} } } }
const result = await build({
  stdin: {
    resolveDir: root,
    loader: 'ts',
    contents: `
      import { createSSRApp, h } from 'vue'
      import { renderToString } from '@vue/server-renderer'
      import Wizard from './src/features/purchases/PurchaseWorkspaceView.vue'
      import Preview from './src/features/purchases/PurchaseDetailPanel.vue'
      import { usePurchasesWorkspace } from './src/features/purchases/usePurchasesWorkspace'
      export { formatMoneyInputWhileTyping, parseMoneyInput } from './src/utils/currency'
      export async function render(purchase, bank = false, preview = false, step = 5) {
        const props = { currencyUnit: 'Toman', suppliers: [], materials: [] }
        return renderToString(createSSRApp({ setup() {
          const workspace = usePurchasesWorkspace(props, () => {})
          workspace.rows.value = [purchase]
          workspace.select(purchase.id)
          workspace.financial.value = [{ id: 'FIN-test', name: bank ? 'Bank' : 'Cash', type: bank ? 'bank' : 'cash', active: true }]
          if (!preview) workspace.openPayment()
          return () => h(preview ? Preview : Wizard, { ...props, workspace, initialStep: workspace.paymentMode.value ? step : 1 })
        } }))
      }
    `,
  },
  bundle: true, write: false, format: 'cjs', platform: 'node',
  external: ['vue', '@vue/server-renderer'],
  plugins: [{ name: 'vue', setup(builder) {
    builder.onLoad({ filter: /\.vue$/ }, async ({ path }) => {
      // The shared dropdown hides its menu while closed. Render its supplied
      // options here so the bank/check choices can be asserted without a DOM.
      if (path.endsWith('/SelectField.vue')) return {
        contents: `import { defineComponent, h } from 'vue'; export default defineComponent({ props: ['options'], setup: p => () => h('select', p.options.map(o => h('option', { value: o.value }, o.label))) })`,
        loader: 'js',
      }
      const source = await readFile(path, 'utf8')
      const { descriptor } = parse(source, { filename: path })
      return { contents: compileScript(descriptor, { id: path, inlineTemplate: true }).content, loader: 'ts' }
    })
  } }],
})
const module = { exports: {} }
new Function('require', 'module', 'exports', result.outputFiles[0].text)(createRequire(import.meta.url), module, module.exports)
const purchase = {
  id: 'PUR-ui', purchaseNumber: 'PUR-1001', status: 'Posted',
  supplierId: '', supplierName: 'Supplier', financialAccountId: '',
  totalRial: 10000, paidRial: 0, remainingRial: 10000, items: [],
  purchaseDate: '2026-09-12', supplierInvoiceNumber: '', notes: '',
  subtotalRial: 10000, discountRial: 0, shippingRial: 0, taxRial: 0, additionalCostsRial: 0,
}
const html = await module.exports.render(purchase)
assert.match(html, /type="range"[^>]*min="0"[^>]*max="100"[^>]*step="5"/, 'Payment slider must use 5% steps')
for (const tick of [0, 25, 50, 75, 100]) assert.ok(html.includes(`>${tick}</span>`), `Missing slider label ${tick}`)
for (const currency of ['Toman', 'Rial']) {
  assert.equal(module.exports.formatMoneyInputWhileTyping('1234567', currency), '1,234,567')
  assert.equal(module.exports.parseMoneyInput('1,234,567', currency), currency === 'Toman' ? 12345670 : 1234567)
}
const details = await module.exports.render(purchase, false, false, 1)
assert.ok(!details.includes('Payment account (optional)'), 'Payment account field is still in purchase details')
for (const label of ['Purchase payment', 'Pay from', 'Method', 'Amount (Toman)', 'Record payment', 'Step 5 of 5']) {
  assert.ok(html.includes(label), `Payment wizard is missing: ${label}`)
}
const bank = await module.exports.render({ ...purchase, financialAccountId: 'FIN-test' }, true)
assert.ok(bank.includes('Bank transfer'), 'Bank payment method is missing')
assert.ok(bank.includes('Check'), 'Bank check option is missing')
const preview = await module.exports.render(purchase, false, true)
assert.ok(preview.includes('Pay purchase'), 'Preview has no direct Pay purchase button')
const draft = await module.exports.render({ ...purchase, status: 'Draft' })
assert.ok(draft.includes('Post saved purchase'), 'Draft cannot be posted from payment step')
console.log('Purchase payment UI smoke tests passed (wizard, bank/check, preview, draft).')
