import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-org-chrome-'))
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
  name: document.querySelector('input[name=name]')?.value ?? '',
  description: document.querySelector('textarea[name=description]')?.value ?? '',
  cards: [...document.querySelectorAll('main article')].map(card => ({
    name: card.querySelector('h2')?.textContent?.trim() || '',
    description: card.querySelector('p')?.textContent?.trim() || ''
  }))
}))()`

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function page(mode, previous) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    if (value?.heading && (!previous || value.epoch !== previous) && (!mode || mode === 'error' && value.alert || mode === 'form' && value.heading === 'Organisation anlegen' || mode === 'detail' && value.url.match(/\/organisationen\/[^/]+$/) || mode === 'list' && value.heading === 'Organisationsübersicht')) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render ${mode}`)
}

async function command(input) {
  const previous = input.url || input.click || input.submit ? await evaluate('performance.timeOrigin') : undefined
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.click) await evaluate(`document.querySelector(${JSON.stringify(input.click)})?.click()`)
  if (input.fill) await evaluate(`(() => {
    const name = document.querySelector('input[name=name]')
    const description = document.querySelector('textarea[name=description]')
    name.value = ${JSON.stringify(input.fill.name)}
    description.value = ${JSON.stringify(input.fill.description)}
    name.dispatchEvent(new Event('input', { bubbles: true }))
    description.dispatchEvent(new Event('input', { bubbles: true }))
  })()`)
  if (input.submit) await evaluate(`document.querySelector('form[action="/organisationen"]')?.requestSubmit()`)
  return { ok: true, page: await page(input.mode, previous) }
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
