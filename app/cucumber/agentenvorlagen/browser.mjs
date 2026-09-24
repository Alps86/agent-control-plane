import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-agent-chrome-'))
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
    if (event.error) item.reject(new Error(event.error.message))
    else item.resolve(event.result)
  }
  await new Promise((resolve, reject) => { ws.onopen = resolve; ws.onerror = reject })
  const target = await send('Target.createTarget', { url: 'about:blank' }, null)
  session = (await send('Target.attachToTarget', { targetId: target.targetId, flatten: true }, null)).sessionId
  await send('Page.enable')
  await send('Runtime.enable')
}

const snapshot = `(() => ({
  url: location.href,
  epoch: performance.timeOrigin,
  heading: document.querySelector('main h1')?.textContent?.trim() || '',
  text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  alert: document.querySelector('main [role=alert]')?.textContent?.trim() || '',
  cards: [...document.querySelectorAll('main article')].map(card => card.textContent?.replace(/\\s+/g, ' ').trim() || ''),
  fields: Object.fromEntries([...document.querySelectorAll('main input, main select, main textarea')]
    .filter(element => element.name).map(element => [element.name, element.value])),
  options: [...document.querySelectorAll('main select option')].map(option => option.textContent?.trim() || ''),
  teamEntries: [...document.querySelectorAll('main section p')]
    .filter(paragraph => paragraph.querySelector('a[href*="vorlage="]'))
    .map(paragraph => paragraph.textContent?.replace(/\\s+/g, ' ').trim() || '')
}))()`

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function page(wait, previous) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    const changed = !previous || value.epoch !== previous.epoch || value.url !== previous.url || value.text !== previous.text || value.alert !== previous.alert
    const matches = !wait || (!wait.text || value.text.includes(wait.text)) && (!wait.url || value.url.includes(wait.url)) && (!wait.heading || value.heading === wait.heading)
    if (value.heading && matches && changed) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render ${JSON.stringify(wait)}`)
}

async function command(input) {
  const previous = await evaluate(snapshot)
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.click) {
    const clicked = await evaluate(`(() => { const element = document.querySelector(${JSON.stringify(input.click)}); if (!element) return false; element.click(); return true })()`)
    if (!clicked) throw new Error(`Element fehlt: ${input.click}`)
  }
  if (input.fill) {
    const filled = await evaluate(`(() => { const element = document.querySelector('[name="${input.fill.name}"]'); if (!element) return false; element.value = ${JSON.stringify(input.fill.value)}; element.dispatchEvent(new Event('input', {bubbles:true})); element.dispatchEvent(new Event('change', {bubbles:true})); return true })()`)
    if (!filled) throw new Error(`Feld fehlt: ${input.fill.name}`)
  }
  if (input.submit) {
    const submitted = await evaluate(`(() => { const form = document.querySelector(${JSON.stringify(input.submit)}); if (!form) return false; form.requestSubmit(); return true })()`)
    if (!submitted) throw new Error(`Formular fehlt: ${input.submit}`)
  }
  return { ok: true, page: await page(input.wait, input.fill ? undefined : previous) }
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
  await rm(directory, { recursive: true, force: true })
}
