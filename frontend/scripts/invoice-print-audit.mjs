import assert from 'node:assert/strict'
import fs from 'node:fs/promises'

const debugPort = process.env.CHROME_DEBUG_PORT || '9229'
const output = process.env.INVOICE_PRINT_OUTPUT || '/tmp/atropaten-invoice-print-preview.pdf'
const pages = await (await fetch(`http://127.0.0.1:${debugPort}/json`)).json()
const pageTarget = pages.find((page) => page.type === 'page')
assert.ok(pageTarget, 'Chrome does not expose a page target')

const socket = new WebSocket(pageTarget.webSocketDebuggerUrl)
await new Promise((resolve) => socket.addEventListener('open', resolve, { once: true }))

let sequence = 0
const pending = new Map()
const pageErrors = []

socket.addEventListener('message', (event) => {
  const value = JSON.parse(event.data)
  if (value.method === 'Runtime.exceptionThrown') {
    pageErrors.push(value.params.exceptionDetails.exception?.description || value.params.exceptionDetails.text)
  }
  if (!value.id) return
  const request = pending.get(value.id)
  pending.delete(value.id)
  value.error ? request.reject(value.error) : request.resolve(value.result)
})

function send(method, params = {}) {
  return new Promise((resolve, reject) => {
    const id = ++sequence
    pending.set(id, { resolve, reject })
    socket.send(JSON.stringify({ id, method, params }))
  })
}

async function evaluate(expression) {
  const response = await send('Runtime.evaluate', {
    expression,
    returnByValue: true,
    awaitPromise: true,
  })
  if (response.exceptionDetails) throw new Error(JSON.stringify(response.exceptionDetails))
  return response.result.value
}

async function waitFor(expression) {
  for (let attempt = 0; attempt < 100; attempt += 1) {
    if (await evaluate(expression)) return
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  throw new Error(`Timed out waiting for: ${expression}`)
}

try {
  await send('Page.enable')
  await send('Runtime.enable')
  await send('Page.addScriptToEvaluateOnNewDocument', {
    source: `window.go={main:{App:new Proxy({}, {get:(_,name)=>async(...args)=>{const response=await fetch('http://127.0.0.1:4174/call/'+String(name),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(args)});if(!response.ok)throw Error(await response.text());return response.json()}})}};`,
  })
  await send('Emulation.setDeviceMetricsOverride', {
    width: 1280,
    height: 1000,
    deviceScaleFactor: 1,
    mobile: false,
  })
  await send('Page.navigate', { url: 'http://127.0.0.1:5178' })
  await waitFor(`!![...document.querySelectorAll('aside nav button')].find((button)=>button.textContent.trim()==='Invoices')`)
  await evaluate(`[...document.querySelectorAll('aside nav button')].find((button)=>button.textContent.trim()==='Invoices').click()`)
  await waitFor(`!![...document.querySelectorAll('tr')].find((row)=>row.textContent.includes('INV-1001'))`)
  await evaluate(`[...document.querySelectorAll('tr')].find((row)=>row.textContent.includes('INV-1001')).click()`)
  await waitFor(`!!document.querySelector('.print-output .invoice-print-document')`)
  await waitFor(`document.querySelector('.invoice-print-logo')?.complete === true`)

  await evaluate(`(() => {
    window.__invoicePrintCalled = false
    window.__invoicePreparedAtPrint = false
    window.print = () => {
      window.__invoicePrintCalled = true
      window.__invoicePreparedAtPrint = document.body.classList.contains('print-preparing')
      setTimeout(() => window.dispatchEvent(new Event('afterprint')), 0)
    }
    const button = [...document.querySelectorAll('button')].find((candidate) => candidate.textContent.includes('Print final invoice'))
    if (!button) throw new Error('Print final invoice button is unavailable')
    button.click()
  })()`)
  await waitFor(`window.__invoicePrintCalled === true`)
  await waitFor(`!document.body.classList.contains('print-preparing')`)
  const printPreparation = await evaluate(`({
    called: window.__invoicePrintCalled,
    preparedAtPrint: window.__invoicePreparedAtPrint,
    preparingClassCleaned: !document.body.classList.contains('print-preparing'),
  })`)
  assert.deepEqual(printPreparation, {
    called: true,
    preparedAtPrint: true,
    preparingClassCleaned: true,
  })

  await send('Emulation.setEmulatedMedia', { media: 'print' })
  const layout = await evaluate(`(() => {
    const documentElement = document.querySelector('.invoice-print-document')
    const logo = document.querySelector('.invoice-print-logo')
    const table = document.querySelector('.invoice-print-table')
    const footer = document.querySelector('.invoice-print-footer')
    return {
      appVisible: getComputedStyle(document.querySelector('#app')).display !== 'none',
      printVisible: getComputedStyle(document.querySelector('.print-output')).display !== 'none',
      logoWidth: Math.round(logo.getBoundingClientRect().width),
      tableWidth: Math.round(table.getBoundingClientRect().width),
      tableColumns: table.querySelectorAll('thead th').length,
      footerBelowTable: footer.getBoundingClientRect().top > table.getBoundingClientRect().bottom,
      documentOverflow: documentElement.scrollWidth > documentElement.clientWidth,
    }
  })()`)

  assert.equal(layout.appVisible, false)
  assert.equal(layout.printVisible, true)
  assert.ok(layout.logoWidth > 55)
  assert.equal(layout.tableColumns, 5)
  assert.equal(layout.footerBelowTable, true)
  assert.equal(layout.documentOverflow, false)
  assert.deepEqual(pageErrors, [])

  const pdf = await send('Page.printToPDF', {
    printBackground: true,
    preferCSSPageSize: true,
  })
  const pdfBuffer = Buffer.from(pdf.data, 'base64')
  const pdfText = pdfBuffer.toString('latin1')
  const mediaBox = pdfText.match(/\/MediaBox\s*\[0 0 ([\d.]+) ([\d.]+)\]/)
  const pageCountMatch = pdfText.match(/\/Count\s+(\d+)/)
  assert.ok(mediaBox, 'Generated PDF has no readable media box')
  assert.ok(pageCountMatch, 'Generated PDF has no readable page count')
  const pageSizePoints = {
    width: Math.round(Number(mediaBox[1])),
    height: Math.round(Number(mediaBox[2])),
  }
  assert.deepEqual(pageSizePoints, { width: 420, height: 595 })
  const pageCount = Number(pageCountMatch[1])
  assert.equal(pageCount, 1)
  await fs.writeFile(output, pdfBuffer)
  console.log(JSON.stringify({ output, pageErrors, printPreparation, layout, pageSizePoints, pageCount, status: 'passed' }, null, 2))
} finally {
  socket.close()
}
