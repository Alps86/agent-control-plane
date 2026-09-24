import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-chrome-'))
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
    try {
      return Number((await readFile(join(directory, 'DevToolsActivePort'), 'utf8')).split('\n')[0])
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 50))
    }
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
  title: document.title,
  heading: document.querySelector('main h1')?.textContent?.trim() || '',
  text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  alert: document.querySelector('[role=alert]')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  links: [...document.querySelectorAll('nav a')].map(a => ({ label: a.textContent.trim(), href: a.getAttribute('href'), visible: a.getBoundingClientRect().width > 0 && a.getBoundingClientRect().right <= innerWidth })),
  active: document.querySelector('nav a[aria-current=page]')?.textContent?.trim() || '',
  horizontal: document.documentElement.scrollWidth > document.documentElement.clientWidth,
  help: document.querySelector('#preview-help')?.textContent?.trim() || ''
}))()`

async function page(expected) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const result = await send('Runtime.evaluate', { expression: snapshot, returnByValue: true })
    if (result.result.value?.heading && (!expected || result.result.value.url === expected)) return result.result.value
    await new Promise((resolve) => setTimeout(resolve, 50))
  }
  throw new Error('Browser page did not render')
}

async function command(input) {
  let expected = input.url
  if (input.width) await send('Emulation.setDeviceMetricsOverride', { width: input.width, height: 800, deviceScaleFactor: 1, mobile: true })
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.click) {
    const target = await send('Runtime.evaluate', { expression: `document.querySelector(${JSON.stringify(input.click)})?.href`, returnByValue: true })
    expected = target.result.value
    const expression = `document.querySelector(${JSON.stringify(input.click)})?.click()`
    await send('Runtime.evaluate', { expression })
  }
  if (input.help) {
    const location = await send('Runtime.evaluate', { expression: 'location.href', returnByValue: true })
    expected = location.result.value
    await send('Runtime.evaluate', { expression: `document.querySelector('.sidebar-help')?.click()` })
  }
  if (input.scroll) {
    const location = await send('Runtime.evaluate', { expression: 'location.href', returnByValue: true })
    expected = location.result.value
    await send('Runtime.evaluate', { expression: `document.querySelector(${JSON.stringify(input.scroll)})?.scrollIntoView()` })
  }
  let result = await page(expected)
  if (input.help) {
    for (let attempt = 0; attempt < 100 && !result.help; attempt++) {
      await new Promise((resolve) => setTimeout(resolve, 50))
      result = await page(expected)
    }
  }
  if (input.assets) {
    const expression = `Promise.all([...document.querySelectorAll('link[rel=stylesheet], script[src]')].map(async element => ({ path: new URL(element.href || element.src).pathname, status: (await fetch(element.href || element.src)).status })))`
    const assets = await send('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true })
    result.assets = assets.result.value
  }
  return { ok: true, page: result }
}

try {
  await connect()
  const lines = createInterface({ input: process.stdin })
  for await (const line of lines) {
    try { process.stdout.write(JSON.stringify(await command(JSON.parse(line))) + '\n') }
    catch (error) { process.stdout.write(JSON.stringify({ ok: false, error: error.message }) + '\n') }
  }
} finally {
  ws?.close()
  chrome.kill()
  await rm(directory, { recursive: true, force: true })
}
