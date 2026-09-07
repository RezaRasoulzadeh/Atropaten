// Development-only final QA. Uses injected bridge failures and never mutates demo data.
import assert from 'node:assert/strict'
const { chromium } = await import(process.env.PLAYWRIGHT_MODULE || 'playwright')

const browser = await chromium.launch({
  executablePath: '/usr/bin/google-chrome',
  headless: true,
  args: ['--no-sandbox'],
})
const page = await browser.newPage({ viewport: { width: 1280, height: 1000 } })
const errors = []
page.on('pageerror', (error) => errors.push(error.message))
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
await page.goto('http://127.0.0.1:5178')
await page.waitForTimeout(1000)

const nav = async (name) => {
  await page.getByRole('button', { name, exact: true }).first().click()
  await page.waitForTimeout(500)
}
const dialogConfirm = async (label) => {
  const dialog = page.locator('dialog[open]')
  await dialog.waitFor()
  await dialog.getByRole('button', { name: label, exact: true }).click()
}

async function expectFailure(endpoint, prepare, action, expected) {
  let requests = 0
  await page.route(`**/call/${endpoint}`, async (route) => {
    requests += 1
    await new Promise((resolve) => setTimeout(resolve, 350))
    await route.fulfill({
      status: 500,
      contentType: 'text/plain',
      body: `Injected final QA failure: ${endpoint}`,
      headers: { 'Access-Control-Allow-Origin': 'http://127.0.0.1:5178' },
    })
  })
  await prepare()
  await action()
  await page.waitForTimeout(500)
  assert.equal(requests, 1, `${endpoint} should issue exactly one request`)
  const errorText = (await page.locator('.toast .alert-error').allTextContents()).join(' ')
  assert.match(errorText, new RegExp(expected, 'i'), `${endpoint} should show an error toast`)
  assert.equal(
    await page.locator('.toast .alert-success').filter({ hasText: endpoint }).count(),
    0,
    `${endpoint} must not show a success toast`,
  )
  await page.unroute(`**/call/${endpoint}`)
}

await nav('Customers')
await expectFailure(
  'DeleteCustomer',
  async () => {
    await page.locator('main [role="button"]').filter({ hasText: 'استودیو طرح' }).first().click()
    await page.getByRole('button', { name: 'Delete', exact: true }).click()
    await dialogConfirm('Delete customer')
  },
  async () => {},
  'Injected final QA failure: DeleteCustomer',
)

await nav('Materials')
await expectFailure(
  'AdjustMaterialStock',
  async () => {
    await page.locator('main .data-table tbody tr').first().click()
    await page.getByLabel('Quantity delta', { exact: true }).fill('2.5')
    await page.getByLabel(/Unit cost \(/).fill('1000')
    await page.getByLabel('Reason / note', { exact: true }).fill('Final QA injected failure')
  },
  async () => page.getByRole('button', { name: 'Record movement', exact: true }).click(),
  'Injected final QA failure: AdjustMaterialStock',
)

await nav('Purchases')
await expectFailure(
  'CancelPurchase',
  async () => {
    await page.locator('main .data-table tbody tr').first().click()
    await page.getByRole('button', { name: 'Cancel / reverse', exact: true }).click()
    await dialogConfirm('Cancel purchase')
  },
  async () => {},
  'Injected final QA failure: CancelPurchase',
)

await nav('Accounting')
await page.getByRole('tab', { name: 'Payments', exact: true }).click()
await page.waitForTimeout(300)
await expectFailure(
  'ReversePayment',
  async () => {},
  async () => page.getByRole('button', { name: 'Reverse', exact: true }).first().click(),
  'Injected final QA failure: ReversePayment',
)

await nav('Checks')
await expectFailure(
  'TransitionCheck',
  async () => {
    await page.locator('main .data-table tbody tr').first().click()
    const transition = page.locator('aside').getByRole('button', { name: /Cleared|Returned|Rejected|Cancelled/ }).first()
    if (await transition.count()) return
    await page.locator('aside').getByRole('button', { name: /Received|Issued|Deposited|Delivered|Cleared|Returned|Rejected|Cancelled/ }).first().waitFor()
  },
  async () => {
    const transition = page.locator('aside').getByRole('button', { name: /Received|Issued|Deposited|Delivered|Cleared|Returned|Rejected|Cancelled/ }).first()
    if (await transition.count()) await transition.click()
    else await page.locator('aside').getByRole('button', { name: /Received|Issued|Deposited|Delivered|Cleared|Returned|Rejected|Cancelled/ }).first().click()
  },
  'Injected final QA failure: TransitionCheck',
)

await nav('Loans')
await expectFailure(
  'CreateLoan',
  async () => {
    await page.getByRole('button', { name: 'New loan', exact: true }).click()
    await page.getByLabel('Counterparty', { exact: true }).fill('Final QA loan')
    await page.getByLabel('Principal (Rial)', { exact: true }).fill('1000000')
  },
  async () => page.getByRole('button', { name: 'Open loan', exact: true }).click(),
  'Injected final QA failure: CreateLoan',
)

await nav('Owners')
await expectFailure(
  'DeleteOwner',
  async () => {
    await page.locator('main .data-table tbody tr').first().click()
    await page.getByRole('button', { name: 'Delete', exact: true }).click()
    await dialogConfirm('Delete owner')
  },
  async () => {},
  'Injected final QA failure: DeleteOwner',
)

assert.deepEqual(errors, [])
await page.close()
await browser.close()
console.log(JSON.stringify({ viewport: 1280, cases: 7, pageErrors: errors, status: 'passed' }, null, 2))
