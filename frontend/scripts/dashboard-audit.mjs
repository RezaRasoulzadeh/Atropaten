// Development-only regression audit; requires the marked demo preview bridge.
import fs from 'node:fs/promises'
const { chromium } = await import(process.env.PLAYWRIGHT_MODULE || 'playwright')
const output =
  process.env.DASHBOARD_AUDIT_OUTPUT || '/tmp/atropaten-dashboard-audit'
await fs.mkdir(output, { recursive: true })
import assert from 'node:assert/strict'
const browser = await chromium.launch({
  executablePath:
    process.env.CHROME_PATH ||
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  headless: true,
})
const results = []
for (const width of [1024, 1600]) {
  const page = await browser.newPage({ viewport: { width, height: 1000 } })
  const errors = []
  page.on('pageerror', (e) => errors.push(e.message))
  await page.addInitScript(() => {
    window.go = {
      main: {
        App: new Proxy(
          {},
          {
            get:
              (_, name) =>
              async (...args) => {
                const r = await fetch(
                  'http://127.0.0.1:4174/call/' + String(name),
                  {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(args),
                  },
                )
                if (!r.ok) throw new Error(await r.text())
                return r.json()
              },
          },
        ),
      },
    }
  })
  await page.goto('http://127.0.0.1:5178')
  await page
    .getByRole('heading', { name: 'Sales & gross profit', exact: true })
    .waitFor()
  await page
    .getByRole('button', { name: /Open order / })
    .first()
    .waitFor()
  assert.equal(
    await page
      .getByRole('heading', { name: 'Needs attention', exact: true })
      .count(),
    0,
  )
  await page
    .getByRole('button', { name: 'Refresh', exact: true })
    .waitFor({ state: 'visible' })
  await page
    .getByRole('button', { name: 'Dashboard period', exact: true })
    .click()
  await page.getByRole('option', { name: 'Last 90 days', exact: true }).click()
  await page.getByText('View chart data', { exact: true }).click()
  await page
    .getByRole('region', { name: 'Daily sales and gross profit' })
    .waitFor()
  await page.getByText('View chart data', { exact: true }).click()

  await page.screenshot({
    path: `${output}/dashboard-${width}.png`,
    fullPage: true,
  })
  const geometry = await page.evaluate(() => ({
    documentOverflow: document.documentElement.scrollWidth > innerWidth,
    mainOverflow:
      document.querySelector('main').scrollWidth >
      document.querySelector('main').clientWidth,
  }))
  assert.equal(geometry.documentOverflow, false)
  assert.equal(geometry.mainOverflow, false)
  await page.getByRole('button', { name: /Notifications, / }).click()
  await page
    .getByRole('heading', { name: 'Notifications', exact: true })
    .waitFor()
  await page.getByRole('button', { name: 'Mark all read', exact: true }).click()
  await page.getByLabel('Unread only', { exact: true }).check()
  await page.getByText('No unread notifications.', { exact: true }).waitFor()
  await page.getByLabel('Unread only', { exact: true }).uncheck()
  await page.screenshot({
    path: `${output}/notifications-${width}.png`,
    fullPage: true,
  })
  await page.keyboard.press('Escape')
  assert.equal(await page.locator('#notification-center').count(), 0)
  await page
    .getByRole('heading', { name: 'Orders needing attention', exact: true })
    .scrollIntoViewIfNeeded()
  await page.screenshot({ path: `${output}/orders-${width}.png` })
  await page
    .getByRole('button', { name: /Open order / })
    .first()
    .click()
  await page
    .locator('main')
    .getByRole('button', { name: 'Orders', exact: true })
    .waitFor({ timeout: 10000 })
  assert.deepEqual(errors, [])
  await page.getByRole('button', { name: 'Dashboard', exact: true }).click()
  await page
    .getByRole('heading', { name: 'Sales & gross profit', exact: true })
    .waitFor()
  await page.route('**/call/GetDashboard', (route) =>
    route.fulfill({
      status: 503,
      body: 'Dashboard unavailable for test',
      headers: { 'Access-Control-Allow-Origin': 'http://127.0.0.1:5178' },
    }),
  )
  await page.getByRole('button', { name: 'Refresh', exact: true }).click()
  await page
    .getByText('Showing the last loaded dashboard.', { exact: true })
    .waitFor()
  await page.getByRole('button', { name: /Notifications, / }).click()
  await page
    .getByText('Showing the last loaded notifications.', { exact: true })
    .waitFor()
  await page
    .getByRole('button', { name: 'Close notifications', exact: true })
    .click()
  await page.unroute('**/call/GetDashboard')
  await page.route('**/call/GetDashboard', async (route) => {
    const response = await route.fetch()
    const data = await response.json()
    for (const key of Object.keys(data)) {
      if (typeof data[key] === 'number') data[key] = 0
      else if (Array.isArray(data[key])) data[key] = []
    }
    await route.fulfill({ response, json: data })
  })
  await page.getByRole('button', { name: 'Refresh', exact: true }).click()
  await page.getByText('No orders need follow-up.', { exact: true }).waitFor()
  await page
    .getByText('No sales or cost activity in this period.', { exact: true })
    .waitFor()
  await page.getByRole('button', { name: /Notifications, / }).click()
  await page.getByText('No notifications right now.', { exact: true }).waitFor()
  await page.screenshot({ path: `${output}/empty-${width}.png` })

  results.push({
    width,
    ...geometry,
    notificationRead: true,
    openOrder: true,
    errors,
  })
  await page.close()
}
console.log(JSON.stringify(results, null, 2))
await browser.close()
