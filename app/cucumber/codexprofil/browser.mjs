import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-codex-profil-chrome-'))
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
  text: (document.querySelector('main') || document.body)?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  workspaceEnabled: document.querySelector('input[name=workspace_enabled]')?.checked || false,
  writeEnabled: document.querySelector('input[name=write_enabled]')?.checked || false,
  pathInput: !!document.querySelector('input[name=workspace_path], input[name=path], input[name=directory]'),
  links: [...document.querySelectorAll('main a')].map(link => link.textContent.trim())
}))()`

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function page(previous) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    if (value?.url && value.text && (!previous || value.epoch !== previous.epoch || value.url !== previous.url || value.heading !== previous.heading || value.text !== previous.text)) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error('Browser page did not render')
}

async function command(input) {
  const previous = input.url || input.click || input.clickText || input.submit ? await evaluate(snapshot) : undefined
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.click) await evaluate(`document.querySelector(${JSON.stringify(input.click)})?.click()`)
  if (input.clickText) await evaluate(`(() => { const link = [...document.querySelectorAll('main a')].find(item => item.textContent.trim() === ${JSON.stringify(input.clickText)}); if (!link) throw new Error('link missing'); link.click() })()`)
  if (input.setCheckbox) await evaluate(`(() => { const field = document.querySelector('input[name=${input.setCheckbox.name}]'); if (!field) throw new Error('checkbox missing'); field.checked = ${input.setCheckbox.value}; field.dispatchEvent(new Event('change', { bubbles: true })) })()`)
  if (input.submit) await evaluate(`document.querySelector('main form')?.requestSubmit()`)
  return { ok: true, page: await page(previous) }
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
