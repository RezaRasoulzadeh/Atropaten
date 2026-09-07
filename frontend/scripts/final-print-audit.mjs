import fs from 'node:fs/promises'
import assert from 'node:assert/strict'

const { chromium } = await import(process.env.PLAYWRIGHT_MODULE || 'playwright')
const output = '/tmp/atropaten-m6-015-print'
await fs.mkdir(output, { recursive: true })

const browser = await chromium.launch({
  executablePath: '/usr/bin/google-chrome',
  headless: true,
  args: ['--no-sandbox'],
})
const page = await browser.newPage({ viewport: { width: 1280, height: 1000 } })
const errors = []
page.on('pageerror', (error) => errors.push(error.message))
await page.clock.setFixedTime(new Date('2026-03-21T12:00:00Z'))
await page.addInitScript(() => {
  window.go = {
    main: {
      App: new Proxy({}, {
        get: (_, name) => async (...args) => {
          const response = await fetch(`http://127.0.0.1:4174/call/${String(name)}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(args),
          })
          if (!response.ok) throw new Error(await response.text())
          return response.json()
        },
      }),
    },
  }
})
await page.goto('http://127.0.0.1:5178')
await page.waitForTimeout(900)
await page.getByRole('button', { name: 'Reports', exact: true }).first().click()
await page.getByLabel('Record ID', { exact: true }).waitFor()

const cases = [
  { kind: 'quote', field: 'Record ID', value: 'QUO-DEMO-03', expected: ['QUO-1003', 'چاپ پوستر'] },
  { kind: 'invoice', field: 'Record ID', value: 'INV-DEMO-01', expected: ['INV-1001', 'استودیو طرحِ فردا', '120,000,000'] },
  { kind: 'payment_receipt', field: 'Record ID', value: 'PAY-DEMO-01', expected: ['PAY-1001', '120,000,000', 'INV-DEMO-01'] },
  { kind: 'customer_statement', field: 'Party ID', value: 'CUS-DEMO-01', expected: ['نگارستان رنگین‌کمان', 'Opening balance', 'Debit', 'Balance'] },
  { kind: 'supplier_statement', field: 'Party ID', value: 'SUP-DEMO-01', expected: ['تأمین کاغذ', 'Opening balance', 'Debit', 'Balance'] },
]
const results = []

for (const item of cases) {
  await page.getByLabel('Document', { exact: true }).selectOption(item.kind)
  await page.waitForTimeout(100)
  const field = page.getByLabel(item.field, { exact: true })
  await field.fill(item.value)
  const response = page.waitForResponse('**/call/GetPrintDocument')
  await page.getByRole('button', { name: 'Load preview', exact: true }).click()
  await response
  await page.waitForFunction((expected) => document.querySelector('.print-document')?.textContent?.includes(expected), item.expected[0])
  const text = await page.locator('.print-document').first().innerText()
  for (const expected of item.expected) assert.match(text, new RegExp(expected.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))

  await page.emulateMedia({ media: 'print' })
  assert.equal(await page.locator('#app').isVisible(), false)
  assert.equal(await page.locator('.print-output').isVisible(), true)
  const printStyles = await page.locator('.print-output .print-document').evaluate((element) => {
    const style = getComputedStyle(element)
    const table = element.querySelector('table')
    const tableStyle = table ? getComputedStyle(table) : null
    return {
      background: style.backgroundColor,
      color: style.color,
      tableBackground: tableStyle?.backgroundColor || null,
      tableColor: tableStyle?.color || null,
    }
  })
  assert.equal(printStyles.background, 'rgb(255, 255, 255)')
  assert.equal(printStyles.color, 'rgb(0, 0, 0)')
  if (printStyles.tableBackground) assert.equal(printStyles.tableBackground, 'rgb(255, 255, 255)')
  if (printStyles.tableColor) assert.equal(printStyles.tableColor, 'rgb(0, 0, 0)')
  await page.screenshot({ path: `${output}/${item.kind}.png` })
  results.push({ ...item, printStyles, textLength: text.length })
  await page.emulateMedia({ media: 'screen' })
}

assert.deepEqual(errors, [])
await browser.close()
await fs.writeFile(`${output}/results.json`, JSON.stringify({ viewport: 1280, results }, null, 2))
console.log(JSON.stringify({ viewport: 1280, cases: results.length, pageErrors: errors, status: 'passed' }, null, 2))
