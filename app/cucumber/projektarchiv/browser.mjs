import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-project-archive-'))
const chrome = spawn('google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu', '--remote-debugging-port=0',
  `--user-data-dir=${directory}`, 'about:blank'
], { stdio: 'ignore' })
let socket
let session
let nextID = 0
const pending = new Map()

async function port() {
  for (let attempt = 0; attempt < 100; attempt++) {
    try { return Number((await readFile(join(directory, 'DevToolsActivePort'), 'utf8')).split('\n')[0]) }
    catch { await new Promise(resolve => setTimeout(resolve, 50)) }
  }
  throw new Error('Chrome did not start')
}

function send(method, params = {}, sessionId = session) {
  const id = ++nextID
  socket.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }))
  return new Promise((resolve, reject) => pending.set(id, { resolve, reject }))
}

async function connect() {
  const address = await port()
  const version = await (await fetch(`http://127.0.0.1:${address}/json/version`)).json()
  socket = new WebSocket(version.webSocketDebuggerUrl)
  socket.onmessage = ({ data }) => {
    const event = JSON.parse(data)
    const item = pending.get(event.id)
    if (!item) return
    pending.delete(event.id)
    if (event.error) item.reject(new Error(event.error.message))
    else item.resolve(event.result)
  }
  await new Promise((resolve, reject) => { socket.onopen = resolve; socket.onerror = reject })
  const target = await send('Target.createTarget', { url: 'about:blank' }, null)
  session = (await send('Target.attachToTarget', { targetId: target.targetId, flatten: true }, null)).sessionId
  await send('Page.enable')
  await send('Runtime.enable')
}

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.exception?.description || result.exceptionDetails.text)
  return result.result.value
}

const snapshot = `(() => ({
  url: location.href,
  epoch: performance.timeOrigin,
  heading: document.querySelector('main h1')?.textContent?.trim() || '',
  text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  links: [...document.querySelectorAll('main a[href]')].map(link => link.getAttribute('href')),
  forms: [...document.querySelectorAll('main form[action]')].map(form => form.getAttribute('action'))
}))()`

async function page(expected, previousEpoch) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    if (value?.heading && (!previousEpoch || value.epoch !== previousEpoch) && (expected === '*' || value.heading === expected)) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render ${expected}`)
}

async function command(input) {
  const previousEpoch = input.url || input.click || input.submit ? await evaluate('performance.timeOrigin') : undefined
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.click) await evaluate(`(() => { const node=document.querySelector(${JSON.stringify(input.click)}); if (!node) throw new Error('Action missing: '+${JSON.stringify(input.click)}); node.click() })()`)
  if (input.submit) await evaluate(`(() => { const form=document.querySelector(${JSON.stringify(input.submit)}); if (!form) throw new Error('Form missing: '+${JSON.stringify(input.submit)}); form.requestSubmit() })()`)
  return { ok: true, page: await page(input.expected, previousEpoch) }
}

try {
  await connect()
  for await (const line of createInterface({ input: process.stdin })) {
    try { process.stdout.write(JSON.stringify(await command(JSON.parse(line))) + '\n') }
    catch (error) { process.stdout.write(JSON.stringify({ ok: false, error: error.message }) + '\n') }
  }
} finally {
  socket?.close()
  chrome.kill()
  await rm(directory, { recursive: true, force: true })
}
