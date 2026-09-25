import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-bericht-chrome-'))
const chrome = spawn('google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu',
  '--remote-debugging-port=0', `--user-data-dir=${directory}`, 'about:blank'
], { stdio: 'ignore' })
let ws
let session
let nextID = 0
const pending = new Map()

async function port() {
  for (let attempt = 0; attempt < 100; attempt++) {
    try { return Number((await readFile(join(directory, 'DevToolsActivePort'), 'utf8')).split('\n')[0]) }
    catch { await new Promise(resolve => setTimeout(resolve, 50)) }
  }
  throw new Error('Chrome DevTools did not start')
}

function send(method, params = {}, sessionId = session) {
  const id = ++nextID
  ws.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }))
  return new Promise((resolve, reject) => pending.set(id, { resolve, reject }))
}

async function connect() {
  const address = await port()
  const version = await (await fetch(`http://127.0.0.1:${address}/json/version`)).json()
  ws = new WebSocket(version.webSocketDebuggerUrl)
  ws.onmessage = ({ data }) => {
    const event = JSON.parse(data)
    const item = pending.get(event.id)
    if (!item) return
    pending.delete(event.id)
    if (event.error) { item.reject(new Error(event.error.message)); return }
    item.resolve(event.result)
  }
  await new Promise((resolve, reject) => { ws.onopen = resolve; ws.onerror = reject })
  const target = await send('Target.createTarget', { url: 'about:blank' }, null)
  session = (await send('Target.attachToTarget', { targetId: target.targetId, flatten: true }, null)).sessionId
  await send('Page.enable')
  await send('Runtime.enable')
  await send('Page.addScriptToEvaluateOnNewDocument', { source: 'window.__alertCalls = 0; window.alert = () => { window.__alertCalls++ }' })
}

const snapshot = `(() => ({
  url: location.href,
  epoch: performance.timeOrigin,
  heading: document.querySelector('main h1')?.textContent?.trim() || '',
  text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  status: Number(document.querySelector('main')?.dataset.status || 200),
  alert: document.querySelector('main [role=alert]')?.textContent?.trim() || '',
  reports: Object.fromEntries([...document.querySelectorAll('main article[id^=agent-]')].map(a => [a.id, a.querySelector('p')?.textContent?.replace(/\\s+/g, ' ').trim() || '']))
}))()`

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function page(mode, urlContains, textContains, textAbsent, alertContains, previous) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    const changed = !previous || value.epoch !== previous.epoch || value.url !== previous.url || value.text !== previous.text || value.alert !== previous.alert
    if (changed && value?.heading && (!mode || value.heading === mode) && (!urlContains || value.url.includes(urlContains)) && (!textContains || value.text.includes(textContains)) && (!textAbsent || !value.text.includes(textAbsent)) && (!alertContains || value.alert.includes(alertContains))) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render ${mode}`)
}

async function command(input) {
  const previous = await evaluate(snapshot)
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.injectOption) await evaluate(`(() => { const item=${JSON.stringify(input.injectOption)}; const element=document.querySelector('[name="'+item.field+'"]'); if(!element) throw Error('field '+item.field+' missing'); element.add(new Option(item.value,item.value)) })()`)
  if (input.select) await evaluate(`(() => { const field=document.querySelector(${JSON.stringify(input.select.selector)}); if(!field) throw Error('select missing'); field.value=${JSON.stringify(input.select.value)}; field.dispatchEvent(new Event('change',{bubbles:true})) })()`)
  if (input.fields) await evaluate(`(() => { for (const [name,value] of Object.entries(${JSON.stringify(input.fields)})) {
    const element=document.querySelector('[name="'+name+'"]'); if(!element) throw Error('field '+name+' missing');
    element.value=value; element.dispatchEvent(new Event('input',{bubbles:true})); element.dispatchEvent(new Event('change',{bubbles:true}));
  }})()`)
  if (input.submit) await evaluate(`(() => { const form=document.querySelector(${JSON.stringify(input.submit)}); if(!form) throw Error('form missing'); form.requestSubmit() })()`)
  if (input.click) await evaluate(`(() => { const element=document.querySelector(${JSON.stringify(input.click)}); if(!element) throw Error('element missing'); element.click() })()`)
  return { ok: true, page: await page(input.mode, input.urlContains, input.textContains, input.textAbsent, input.alertContains, input.url || input.submit || input.click ? previous : undefined) }
}

try {
  await connect()
  for await (const line of createInterface({ input: process.stdin })) {
    try { process.stdout.write(JSON.stringify(await command(JSON.parse(line))) + '\n') }
    catch (error) { process.stdout.write(JSON.stringify({ ok: false, error: error.message }) + '\n') }
  }
} finally {
  ws?.close()
  chrome.kill()
  await rm(directory, { recursive: true, force: true, maxRetries: 8, retryDelay: 100 })
}
