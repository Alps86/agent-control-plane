import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-projektort-chrome-'))
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
  alert: document.querySelector('main [role=alert]')?.textContent?.trim() || '',
  pathInput: Boolean(document.querySelector('main input[name=path], main textarea[name=path], main input[name=repository_url]')),
  enabled: Boolean(document.querySelector('main input[name=kind][value=managed_directory]:checked, main input[name=enabled]:checked'))
}))()`
async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}
async function page(mode, previousEpoch) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    const ready = mode === 'detail' ? value?.url.match(/\/projekte\/[^/]+$/)
      : mode === 'location' ? value?.url.endsWith('/ausfuehrungsort')
      : mode === 'error' ? Boolean(value?.alert || value?.text.match(/nicht gefunden|verweigert|not found/i))
      : Boolean(value?.heading)
    if (ready && (!previousEpoch || value.epoch !== previousEpoch || mode === 'error')) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render ${mode}`)
}
async function command(input) {
  const previousEpoch = input.url || input.click || input.submit ? await evaluate('performance.timeOrigin') : undefined
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.click) await evaluate(`(() => { const item = document.querySelector(${JSON.stringify(input.click)}); if (!item) throw new Error('Link or button missing: ' + ${JSON.stringify(input.click)}); item.click() })()`)
  if (input.enable) await evaluate(`(() => { const item = document.querySelector('input[name=kind][value=managed_directory], input[name=enabled]'); if (!item) throw new Error('Managed location control missing'); item.checked = true; item.dispatchEvent(new Event('change', { bubbles: true })) })()`)
  if (input.submit) await evaluate(`document.querySelector('main form')?.requestSubmit()`)
  return { ok: true, page: await page(input.mode, previousEpoch) }
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
