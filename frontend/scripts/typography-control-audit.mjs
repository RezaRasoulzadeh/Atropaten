// Development-only browser validation for M6-012. Requires the ui_preview Go bridge.
import fs from 'node:fs/promises'

const { chromium } = await import(process.env.PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({
  executablePath: '/usr/bin/google-chrome',
  headless: true,
  args: ['--no-sandbox'],
})
const output = process.env.UI_AUDIT_OUTPUT || '/tmp/atropaten-m6-012-controls'
await fs.mkdir(output, { recursive: true })

const widths = process.env.AUDIT_WIDTH ? [Number(process.env.AUDIT_WIDTH)] : [1024, 1280, 1600]
const results = []
const waitForVisible = async (locator, timeout = 8000) => {
  await locator.waitFor({ state: 'visible', timeout })
  return locator
}

const readMetric = (element) => {
  const rect = element.getBoundingClientRect()
  const style = getComputedStyle(element)
  return {
    tag: element.tagName.toLowerCase(),
    label: element.getAttribute('aria-label') || element.getAttribute('placeholder') || element.textContent?.trim().slice(0, 36) || '',
    className: element.className,
    width: Math.round(rect.width * 10) / 10,
    height: Math.round(rect.height * 10) / 10,
    fontSize: style.fontSize,
    lineHeight: style.lineHeight,
    fontWeight: style.fontWeight,
    paddingBlock: `${style.paddingTop} ${style.paddingBottom}`,
    border: `${style.borderWidth} ${style.borderStyle}`,
    borderColor: style.borderColor,
    textAlign: style.textAlign,
    alignItems: style.alignItems,
    outline: style.outlineStyle,
    boxShadow: style.boxShadow,
  }
}

const captureMetrics = async (page, state) => page.evaluate((stateName) => {
  const readMetric = (element) => {
    const rect = element.getBoundingClientRect()
    const style = getComputedStyle(element)
    return {
      tag: element.tagName.toLowerCase(),
      label: element.getAttribute('aria-label') || element.getAttribute('placeholder') || element.textContent?.trim().slice(0, 36) || '',
      className: element.className,
      width: Math.round(rect.width * 10) / 10,
      height: Math.round(rect.height * 10) / 10,
      cssHeight: style.height,
      fontSize: style.fontSize,
      lineHeight: style.lineHeight,
      fontWeight: style.fontWeight,
      paddingBlock: `${style.paddingTop} ${style.paddingBottom}`,
      border: `${style.borderWidth} ${style.borderStyle}`,
      borderColor: style.borderColor,
      textAlign: style.textAlign,
      alignItems: style.alignItems,
      outline: style.outlineStyle,
      boxShadow: style.boxShadow,
      transform: style.transform,
      scale: style.scale,
    }
  }
  const first = (selector) => document.querySelector(selector)
  const named = (selector, label) => [...document.querySelectorAll(selector)].find((element) =>
    (element.getAttribute('aria-label') || '').toLowerCase().includes(label.toLowerCase()),
  )
  const picks = {
    toolbarSearch: first('.navbar input[type="search"]'),
    toolbarSelect: named('.navbar select', 'currency') || first('.navbar select'),
    toolbarButton: named('.navbar button.btn', 'new order') || first('.navbar button.btn'),
    formInput: first('main input.input:not([type="search"]):not([readonly])'),
    formSelect: first('main select.select'),
    formButton: first('main button[form].btn, main form button.btn:not(.btn-sm):not(.btn-xs)'),
    textarea: first('main textarea.textarea, main textarea'),
    jalaliDate: first('main input[aria-haspopup="dialog"]'),
    moneyOrQuantity: first('main input[inputmode="decimal"]'),
    inspectorControl: first('main aside input.input, main aside select.select, main aside textarea'),
    dialogButton: first('dialog[open] button.btn'),
    tableAction: first('main .data-table button.btn, main [role="button"] .btn'),
    registerRow: first('main [role="button"]'),
  }
  const metrics = {}
  for (const [name, element] of Object.entries(picks)) {
    if (element) metrics[name] = readMetric(element)
  }
  const editableFocus = first('main input.input, main select.select, main textarea')
  if (editableFocus) {
    editableFocus.focus()
    metrics.focusedEditable = readMetric(editableFocus)
  }
  const typeScale = [...document.querySelectorAll('main h1, main h2, main h3, main p, main label, main th, main .badge')]
    .map((element) => readMetric(element))
    .filter((metric) => metric.label)
    .slice(0, 40)
  return { state: stateName, metrics, typeScale }
}, state)

const installBridge = async (page) => {
  await page.addInitScript(() => {
    window.go = { main: { App: new Proxy({}, { get: (_, name) => async (...args) => {
      const response = await fetch(`http://127.0.0.1:4174/call/${String(name)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(args),
      })
      if (!response.ok) throw new Error(await response.text())
      return response.json()
    } }) } }
  })
}

for (const width of widths) {
  const page = await browser.newPage({ viewport: { width, height: 1000 } })
  const errors = []
  page.on('pageerror', (error) => errors.push(error.message))
  await installBridge(page)
  await page.goto('http://127.0.0.1:5178')
  await page.waitForTimeout(1000)

  const record = async (state, screenshot = true) => {
    if (screenshot) await page.screenshot({ path: `${output}/${width}-${state}.png`, fullPage: true })
    results.push({ width, ...(await captureMetrics(page, state)), errors: [...errors] })
  }

  await record('dashboard')

  await page.getByRole('button', { name: 'Accounting', exact: true }).first().click()
  await page.waitForTimeout(350)
  const expensesTab = page.getByRole('tab', { name: 'Expenses', exact: true })
  if (await expensesTab.count()) {
    await expensesTab.click()
    await page.waitForTimeout(250)
    await record('accounting-expenses')
  }

  await page.getByRole('button', { name: 'Customers', exact: true }).first().click()
  await page.waitForTimeout(350)
  await page.getByRole('button', { name: 'New customer', exact: true }).click()
  await page.waitForTimeout(200)
  await record('customer-form')
  await page.getByRole('button', { name: 'Cancel', exact: true }).click()

  await page.getByRole('button', { name: 'Materials', exact: true }).first().click()
  await page.waitForTimeout(350)
  await page.getByRole('button', { name: 'New material', exact: true }).click()
  await page.waitForTimeout(200)
  await record('material-form')

  await page.getByRole('button', { name: 'Orders', exact: true }).first().click()
  await page.waitForTimeout(900)
  await (await waitForVisible(page.getByRole('button', { name: /ORD-1004/ }).first())).click()
  await page.waitForTimeout(300)
  await record('order-detail')
  const calendarButton = page.getByRole('button', { name: 'Open Jalali calendar', exact: true }).first()
  if (await calendarButton.count()) {
    await calendarButton.click()
    await page.waitForTimeout(100)
    await record('jalali-calendar')
    await page.keyboard.press('Escape')
  }

  await page.getByRole('button', { name: 'Checks', exact: true }).first().click()
  await page.waitForTimeout(350)
  const newCheck = page.getByRole('button', { name: 'New check', exact: true })
  if (await newCheck.count()) {
    await newCheck.click()
    await page.waitForTimeout(200)
    const inspectorCalendar = page.getByRole('button', { name: 'Open Jalali calendar', exact: true }).first()
    if (await inspectorCalendar.count()) {
      await inspectorCalendar.click()
      await page.waitForTimeout(100)
      await page.locator('[role="dialog"][aria-label="Jalali calendar"]').scrollIntoViewIfNeeded()
      await page.waitForTimeout(100)
      const clipping = await page.evaluate(() => {
        const popover = document.querySelector('[role="dialog"][aria-label="Jalali calendar"]')
        if (!popover) return null
        const rect = popover.getBoundingClientRect()
        const clippedBy = []
        let ancestor = popover.parentElement
        while (ancestor) {
          const style = getComputedStyle(ancestor)
          const clips = ['hidden', 'clip'].includes(style.overflow) || ['hidden', 'clip'].includes(style.overflowY)
          if (clips) {
            const box = ancestor.getBoundingClientRect()
            const outside = rect.left < box.left || rect.right > box.right || rect.top < box.top || rect.bottom > box.bottom
            if (outside) clippedBy.push({ tag: ancestor.tagName.toLowerCase(), className: ancestor.className, rect: box.toJSON() })
          }
          ancestor = ancestor.parentElement
        }
        return { rect: rect.toJSON(), clippedBy }
      })
      await record('inspector-jalali-calendar')
      results[results.length - 1].clipping = clipping
      await page.keyboard.press('Escape')
    }
  }

  await page.getByRole('button', { name: 'Customers', exact: true }).first().click()
  await page.waitForTimeout(350)
  const customerRow = page.locator('main [role="button"]').filter({ hasText: 'استودیو طرح فردا' }).first()
  if (await customerRow.count()) await customerRow.click()
  const deleteButton = page.getByRole('button', { name: 'Delete', exact: true })
  if (await deleteButton.count()) {
    await deleteButton.click()
    await page.waitForTimeout(100)
    await record('confirmation-dialog')
    await page.keyboard.press('Escape')
  }

  await page.close()
}

await browser.close()
await fs.writeFile(`${output}/results.json`, JSON.stringify(results, null, 2))
console.log(JSON.stringify(results, null, 2))
