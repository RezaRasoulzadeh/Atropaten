// Development-only browser validation. Requires the ui_preview Go test bridge.
// PLAYWRIGHT_MODULE points to an installed playwright module; no app dependency.
import fs from 'node:fs/promises'
const { chromium } = await import(process.env.PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({ executablePath: '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] })
const output = process.env.UI_AUDIT_OUTPUT || '/tmp/atropaten-ui-audit'
await fs.mkdir(output, { recursive: true })
const pages = process.argv.slice(2)
const contexts = pages.length ? pages : ['Dashboard','Orders','Quotes','Production','Customers','Services','Materials','Machines','Purchases','Suppliers','Accounting','Invoices','Checks','Loans','Owners','Reports','Settings']
const results = []
for (const width of [1024,1280,1600]) {
 const page = await browser.newPage({ viewport: { width, height: 1000 } })
 const errors=[]
 page.on('pageerror', e=>errors.push(e.message))
 await page.addInitScript(() => {
  window.go = { main: { App: new Proxy({}, { get: (_,name) => async (...args) => {
   const response=await fetch('http://127.0.0.1:4174/call/'+String(name), { method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(args) })
   if(!response.ok) throw new Error(await response.text())
   return response.json()
  } }) } }
 })
 await page.goto('http://127.0.0.1:5178')
 await page.waitForTimeout(1300)
 for(const name of contexts) {
  await page.getByRole('button',{name,exact:true}).first().click()
  await page.waitForTimeout(550)
  await page.screenshot({path: `${output}/${width}-${name.toLowerCase()}.png`,fullPage:true})
  const geometry=await page.evaluate(()=>({
   overflow: document.documentElement.scrollWidth > innerWidth,
   mainOverflow: document.querySelector('main').scrollWidth > document.querySelector('main').clientWidth,
   tables: [...document.querySelectorAll('.data-table')].map(e=>({scroll:e.scrollWidth,client:e.clientWidth})),
   heading:document.querySelector('main h1')?.textContent,
   selectAlignment:[...document.querySelectorAll('select')].map(e=>getComputedStyle(e).textAlign),
   alerts:[...document.querySelectorAll('main [role="alert"]')].map(e=>e.textContent.trim())
  }))
  results.push({width,name,...geometry,errors:[...errors]})
 }
 await page.close()
}
await browser.close()
await fs.writeFile(`${output}/results.json`,JSON.stringify(results,null,2))
console.log(JSON.stringify(results,null,2))
