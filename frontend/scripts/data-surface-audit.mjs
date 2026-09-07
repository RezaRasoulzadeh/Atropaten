// M6-013 browser audit. Requires the ui_preview bridge and Playwright.
import fs from 'node:fs/promises'

const { chromium } = await import(process.env.PLAYWRIGHT_MODULE || 'playwright')
const output = process.env.UI_DATA_AUDIT_OUTPUT || '/tmp/atropaten-m6-013-data-surfaces'
const widths = [1024, 1280, 1600]
const contexts = ['Orders', 'Quotes', 'Production', 'Customers', 'Services', 'Materials', 'Machines', 'Purchases', 'Suppliers', 'Accounting', 'Invoices', 'Checks', 'Loans', 'Owners', 'Reports']
const stateContexts = new Set(['Orders', 'Customers', 'Materials', 'Machines', 'Purchases', 'Invoices', 'Checks', 'Loans', 'Owners', 'Reports'])
await fs.mkdir(output, { recursive: true })

const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] })
const results = []

for (const width of widths) {
  const page = await browser.newPage({ viewport: { width, height: 1000 } })
  const errors = []
  page.on('pageerror', (error) => errors.push(error.message))
  await page.addInitScript(() => {
    window.go = { main: { App: new Proxy({}, { get: (_, name) => async (...args) => {
      const response = await fetch(`http://127.0.0.1:4174/call/${String(name)}`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(args),
      })
      if (!response.ok) throw new Error(await response.text())
      return response.json()
    } }) } }
  })
  await page.goto('http://127.0.0.1:5178')
  await page.waitForTimeout(1000)

  for (const name of contexts) {
    await page.getByRole('button', { name, exact: true }).first().click()
    await page.waitForTimeout(450)
    const slug = name.toLowerCase().replaceAll(/[^a-z0-9]+/g, '-')
    const base = { width, name, errors: [...errors] }
    const table = page.locator('.data-table').first()
    const rows = page.locator('.data-table tbody tr')
    base.surface = await page.evaluate(() => ({
      documentOverflow: document.documentElement.scrollWidth > innerWidth,
      mainOverflow: document.querySelector('main')?.scrollWidth > document.querySelector('main')?.clientWidth,
      selectAlignment: [...document.querySelectorAll('select')].map((element) => getComputedStyle(element).textAlign),
      tableCount: document.querySelectorAll('.data-table').length,
      localOverflow: [...document.querySelectorAll('.data-table')].map((element) => ({ scroll: element.scrollWidth, client: element.clientWidth })),
    }))

    if (stateContexts.has(name)) {
      if (await rows.count()) {
        await rows.first().click()
        await page.waitForTimeout(150)
        await page.screenshot({ path: `${output}/${width}-${slug}-selected.png`, fullPage: true })
        await rows.first().hover()
        await page.screenshot({ path: `${output}/${width}-${slug}-hover.png`, fullPage: true })
        if ((await rows.first().getAttribute('tabindex')) !== null) {
          // Move focus away and back through keyboard navigation so :focus-visible is exercised.
          await rows.first().focus()
          await page.keyboard.press('Shift+Tab')
          await page.keyboard.press('Tab')
        }
        base.rowFocus = await rows.first().evaluate((element) => {
          const style = getComputedStyle(element)
          return { tabindex: element.getAttribute('tabindex'), outlineStyle: style.outlineStyle, outlineWidth: style.outlineWidth }
        })
        await page.screenshot({ path: `${output}/${width}-${slug}-focus.png`, fullPage: true })
      }

      const search = page.locator('main input[placeholder*="Search" i]').first()
      if (await search.count()) {
        await search.fill('__m6_013_no_results__')
        await page.waitForTimeout(100)
        await page.screenshot({ path: `${output}/${width}-${slug}-no-results.png`, fullPage: true })
        await search.fill('')
      }

      const control = page.locator('main input, main select, main textarea').first()
      if (await control.count()) {
        await control.focus()
        base.controlFocus = await control.evaluate((element) => {
          const style = getComputedStyle(element)
          return {
            tag: element.tagName,
            borderStyle: style.borderStyle,
            borderWidth: style.borderWidth,
            borderColor: style.borderColor,
            outline: style.outline,
            boxShadow: style.boxShadow,
          }
        })
      }
    }
    results.push(base)
  }
  await page.close()
}

await browser.close()
await fs.writeFile(`${output}/results.json`, JSON.stringify(results, null, 2))
console.log(JSON.stringify(results, null, 2))
