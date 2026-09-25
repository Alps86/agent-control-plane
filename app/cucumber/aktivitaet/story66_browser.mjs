import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-activity-chrome-'))
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
  await send('Browser.setDownloadBehavior', { behavior: 'allow', downloadPath: directory, eventsEnabled: true }, null)
}

const snapshot = `(() => ({
  url: location.href,
  heading: document.querySelector('main h1')?.textContent?.trim() || '',
  text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  links: [...document.querySelectorAll('main a[href]')].map(link => link.getAttribute('href')),
  linkTexts: [...document.querySelectorAll('main a[href]')].map(link => link.textContent.trim()),
  times: [...document.querySelectorAll('main time[datetime]')].map(node => node.getAttribute('datetime'))
}))()`

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function page(mode) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    const ready = mode === 'activity' ? value?.heading.includes('Aktivität')
      : mode === 'task' ? value?.heading === 'Startseite prüfen'
      : mode === 'organization' ? value?.heading === 'Nordstern'
      : Boolean(value?.heading)
    if (ready) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render ${mode}`)
}

async function command(input) {
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.filters) {
    await page('activity')
    await evaluate(`(() => {
      const form = document.querySelector('form[aria-label="Aktivität filtern"]')
      if (!form) throw new Error('Aktivitätsfilter fehlt')
      const filters = ${JSON.stringify(input.filters)}
      for (const [name, value] of Object.entries(filters)) {
        const field = form.elements.namedItem(name)
        if (!field) throw new Error('Filterfeld fehlt: ' + name)
        field.value = value
        field.dispatchEvent(new Event('input', { bubbles: true }))
        field.dispatchEvent(new Event('change', { bubbles: true }))
      }
      form.requestSubmit()
    })()`)
    for (let attempt = 0; attempt < 100; attempt++) {
      const matched = await evaluate(`(() => {
        const query = new URLSearchParams(location.search)
        return Object.entries(${JSON.stringify(input.filters)}).every(([name, value]) => query.get(name) === value)
      })()`)
      if (matched) break
      await new Promise(resolve => setTimeout(resolve, 50))
    }
  }
  if (input.click) await evaluate(`(() => {
    const element = document.querySelector(${JSON.stringify(input.click)})
    if (!element) throw new Error('Action missing: ' + ${JSON.stringify(input.click)})
    element.click()
  })()`)
  if (input.download) {
    const target = join(directory, 'aktivitaet.csv')
    await rm(target, { force: true })
    await evaluate(`(() => {
      const link = document.querySelector('a[download="aktivitaet.csv"]')
      if (!link) throw new Error('CSV-Link fehlt')
      link.click()
    })()`)
    for (let attempt = 0; attempt < 100; attempt++) {
      try { return { ok: true, page: await page(input.mode), download: await readFile(target, 'utf8') } }
      catch { await new Promise(resolve => setTimeout(resolve, 50)) }
    }
    throw new Error('CSV-Download blieb aus')
  }
  return { ok: true, page: await page(input.mode) }
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
