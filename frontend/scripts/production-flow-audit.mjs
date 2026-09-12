// Run against the disposable ui_preview bridge, Vite :5178 and headless Chrome :9229.
import assert from 'node:assert/strict'
import fs from 'node:fs/promises'

const pages = await (await fetch('http://127.0.0.1:9229/json')).json()
const ws = new WebSocket(pages.find(p => p.type === 'page').webSocketDebuggerUrl)
await new Promise(resolve => ws.addEventListener('open', resolve, { once: true }))
let seq = 0
const pending = new Map()
const errors = []
ws.addEventListener('message', event => {
  const value = JSON.parse(event.data)
  if (value.method === 'Runtime.exceptionThrown') errors.push(value.params.exceptionDetails.exception?.description || value.params.exceptionDetails.text)
  if (value.id) {
    const request = pending.get(value.id)
    pending.delete(value.id)
    value.error ? request.reject(value.error) : request.resolve(value.result)
  }
})
function send(method, params = {}) {
  return new Promise((resolve, reject) => {
    const id = ++seq
    pending.set(id, { resolve, reject })
    ws.send(JSON.stringify({ id, method, params }))
  })
}
async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true })
  if (result.exceptionDetails) throw Error(JSON.stringify(result.exceptionDetails))
  return result.result.value
}
async function waitFor(expression) {
  for (let n = 0; n < 100; n++) {
    if (await evaluate(expression)) return
    await new Promise(resolve => setTimeout(resolve, 100))
  }
  throw Error(`Timed out: ${expression}\n${await evaluate('document.body.innerText')}\n${JSON.stringify(errors)}`)
}
const click = text => evaluate(`(() => { const button = [...document.querySelectorAll('button')].find(b => b.textContent.trim() === ${JSON.stringify(text)}); if(!button) throw Error('Missing button: '+${JSON.stringify(text)}); button.click(); })()`)
await send('Page.enable')
await send('Runtime.enable')
await send('Page.addScriptToEvaluateOnNewDocument', { source: `window.go={main:{App:new Proxy({}, {get:(_,name)=>async(...args)=>{const r=await fetch('http://127.0.0.1:4174/call/'+String(name),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(args)});if(!r.ok)throw Error(await r.text());return r.json()}})}};` })
const results = []
try {
  for (const width of [1024,1600]) {
    await send('Emulation.setDeviceMetricsOverride', { width, height: 1000, deviceScaleFactor: 1, mobile: false })
    await send('Page.navigate', { url: 'http://127.0.0.1:5178' })
    await waitFor(`!![...document.querySelectorAll('aside nav button')].find(b=>b.textContent.trim()==='Production')`)
    await click('Production')
    await waitFor(`!![...document.querySelectorAll('button')].find(b=>b.textContent.trim()==='Manage production')`)
    await click('Manage production')
    await waitFor(`!!document.querySelector('nav[aria-label="Production management steps"]')`)
    const steps = await evaluate(`[...document.querySelectorAll('nav[aria-label="Production management steps"] strong')].map(e=>e.textContent.trim())`)
    assert.deepEqual(steps, ['Overview','Materials','Outsourcing'])
    for (const step of steps) {
      await evaluate(`[...document.querySelectorAll('nav[aria-label="Production management steps"] button')].find(b=>b.querySelector('strong').textContent.trim()===${JSON.stringify(step)}).click()`)
      await new Promise(resolve => setTimeout(resolve,200))
      if (step === 'Overview') {
        const stacked = await evaluate(`(() => {const heading = name => [...document.querySelectorAll('h3')].find(e=>e.textContent.trim()===name).parentElement.getBoundingClientRect();const cost=heading('Cost'),schedule=heading('Schedule');return schedule.top>=cost.bottom && Math.abs(schedule.left-cost.left)<1;})()`)
        assert.equal(stacked,true,'Cost must be above Schedule')
      }
      if (step === 'Materials') {
        assert.equal(await evaluate(`document.body.textContent.includes('Planned material for in-house work')`),false)
        await click('Use material')
        await waitFor(`!!document.querySelector('input[type="range"]') && document.querySelector('input[type="range"]').getBoundingClientRect().height>0`)
        const reserved = await evaluate(`Number([...document.querySelectorAll('span')].find(e=>e.textContent.trim().startsWith('Reserved:')).textContent.trim().split(' ')[1])`)
        assert.ok(reserved>0)
        await evaluate(`(() => {const slider=document.querySelector('input[type="range"]');slider.value='50';slider.dispatchEvent(new Event('input',{bubbles:true}));})()`)
        const usage = `document.querySelector('input.pe-14')`
        assert.equal(Number(await evaluate(`${usage}.value`)),reserved/2)
        await click('Max')
        assert.equal(Number(await evaluate(`${usage}.value`)),reserved)
        await evaluate(`(() => {const input=${usage};input.value=${JSON.stringify(String(reserved+1))};input.dispatchEvent(new Event('input',{bubbles:true}));})()`)
        assert.equal(Number(await evaluate(`${usage}.value`)),reserved+1,'Typed usage above reservation must remain editable')
        assert.equal(Number(await evaluate(`document.querySelector('input[type="range"]').value`)),100)
      }
      const geometry = await evaluate(`({page:document.documentElement.scrollWidth>innerWidth,fields:[...document.querySelectorAll('main input,main textarea,main button.select')].filter(e=>e.getBoundingClientRect().width && e.getBoundingClientRect().right>innerWidth+1).length})`)
      assert.deepEqual(geometry, { page:false,fields:0 })
      const shot = await send('Page.captureScreenshot', { format:'png' })
      await fs.writeFile('/tmp/production-flow-'+width+'-'+step.toLowerCase()+'.png',Buffer.from(shot.data,'base64'))
    }
    results.push({width,steps,overflow:false})
  }
  assert.deepEqual(errors,[])
  console.log(JSON.stringify(results,null,2))
} finally { ws.close() }
