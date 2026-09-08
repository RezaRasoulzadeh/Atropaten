// Use only with the marked disposable demo preview: this audit creates/deletes a draft and pauses/resumes a demo job.
import fs from 'node:fs/promises'
const { chromium } = await import(process.env.PLAYWRIGHT_MODULE || 'playwright')
const output = process.env.PRODUCTION_AUDIT_OUTPUT || '/tmp/atropaten-production-audit'
await fs.mkdir(output, { recursive: true })
import assert from 'node:assert/strict'
const browser = await chromium.launch({
  executablePath:
    process.env.CHROME_PATH || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
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
                const r = await fetch('http://127.0.0.1:4174/call/' + String(name), {
                  method: 'POST',
                  headers: { 'Content-Type': 'application/json' },
                  body: JSON.stringify(args),
                })
                if (!r.ok) throw new Error(await r.text())
                return r.json()
              },
          },
        ),
      },
    }
  })
  await page.goto('http://127.0.0.1:5178')
  await page.locator('aside nav').getByRole('button', { name: 'Production', exact: true }).click()
  await page.locator('main article[role="button"]').first().waitFor()
  assert.equal(await page.getByRole('tab').count(), 0)
  assert.equal(await page.getByText('Job inspector', { exact: true }).count(), 0)
  await page.screenshot({ animations: 'disabled', path: `${output}/production-list-${width}.png` })
  const first = page.locator('main article[role="button"]').first()
  const title = (await first.innerText()).match(/JOB-\d+|PRD-\d+/)?.[0]
  await page
    .getByRole('searchbox', { name: 'Search production jobs', exact: true })
    .fill(title || '')
  await page.getByRole('button', { name: 'Production status', exact: true }).click()
  await page.getByRole('option', { name: 'In Progress', exact: true }).click()
  await first.focus()
  await page.keyboard.press('Enter')
  await page.getByRole('tab', { name: 'Overview', exact: true }).waitFor()
  assert.equal(
    await page.getByRole('heading', { name: 'Production queue', exact: true }).count(),
    0,
  )
  for (const tab of ['Overview', 'Reservations', 'Consumption', 'Outsourcing']) {
    await page.getByRole('tab', { name: tab, exact: true }).click()
    assert.equal(
      await page.getByRole('tab', { name: tab, exact: true }).getAttribute('aria-selected'),
      'true',
    )
    await page.screenshot({
      animations: 'disabled',
      path: `${output}/production-${tab.toLowerCase()}-${width}.png`,
    })
    const geometry = await page.evaluate(() => ({
      page: document.documentElement.scrollWidth > innerWidth,
      workspace:
        document.querySelector('main').scrollWidth > document.querySelector('main').clientWidth,
      fields: [
        ...document.querySelectorAll('main input, main textarea, main button.select'),
      ].filter(
        (e) => e.getBoundingClientRect().width && e.getBoundingClientRect().right > innerWidth,
      ).length,
    }))
    assert.deepEqual(geometry, { page: false, workspace: false, fields: 0 })
  }
  await page.getByRole('button', { name: 'Pause production', exact: true }).click()
  await page.getByRole('button', { name: 'Resume production', exact: true }).waitFor()
  await page.locator('main').getByRole('button', { name: 'Production', exact: true }).click()
  await page.getByText('No jobs match this queue', { exact: true }).waitFor()
  await page.getByRole('button', { name: 'Production status', exact: true }).click()
  await page.getByRole('option', { name: 'Paused', exact: true }).click()
  await page.locator('main article[role="button"]').first().click()
  await page.getByRole('button', { name: 'Resume production', exact: true }).click()
  await page.getByRole('button', { name: 'Pause production', exact: true }).waitFor()
  await page.locator('main').getByRole('button', { name: 'Production', exact: true }).click()
  await page.getByText('No jobs match this queue', { exact: true }).waitFor()
  await page.getByRole('button', { name: 'Production status', exact: true }).click()
  await page.getByRole('option', { name: 'All', exact: true }).click()

  assert.equal(
    await page.getByRole('searchbox', { name: 'Search production jobs', exact: true }).inputValue(),
    title || '',
  )
  await page.getByRole('button', { name: 'New production job', exact: true }).click()
  await page.getByRole('heading', { name: 'Job details', exact: true }).waitFor()
  assert.equal(
    await page.getByRole('heading', { name: 'Production queue', exact: true }).count(),
    0,
  )
  await page.screenshot({ animations: 'disabled', path: `${output}/production-new-${width}.png` })
  await page.getByRole('button', { name: 'Confirmed order', exact: true }).click()
  await page.getByRole('option').nth(1).click()
  await page.getByRole('button', { name: 'Create job', exact: true }).click()
  await page.getByRole('tab', { name: 'Overview', exact: true }).waitFor()
  await page.getByRole('button', { name: 'Delete job', exact: true }).click()
  await page.getByRole('button', { name: 'Delete permanently', exact: true }).click()
  await page.getByRole('heading', { name: 'Production queue', exact: true }).waitFor()
  assert.equal(await page.getByRole('tab').count(), 0)
  assert.deepEqual(errors, [])
  results.push({
    width,
    queueOnly: true,
    tabs: true,
    backPreservesSearch: true,
    statusFilterSurvivesTransitions: true,
    createAndDelete: true,
    errors,
  })
  await page.close()
}
console.log(JSON.stringify(results, null, 2))
await browser.close()
