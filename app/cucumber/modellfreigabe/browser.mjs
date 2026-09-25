import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-modellfreigabe-chrome-'))
const chrome = spawn('google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu',
  '--remote-debugging-port=0', `--user-data-dir=${directory}`, 'about:blank'
], { stdio: 'ignore' })
let socket
let session
let nextID = 0
const pending = new Map()
const responses = new Map()

async function debuggingPort() {
  for (let attempt = 0; attempt < 100; attempt++) {
    try { return Number((await readFile(join(directory, 'DevToolsActivePort'), 'utf8')).split('\n')[0]) }
    catch { await new Promise(resolve => setTimeout(resolve, 50)) }
  }
  throw new Error('Chrome DevTools did not start')
}

function send(method, params = {}, sessionId = session) {
  const id = ++nextID
  socket.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }))
  return new Promise((resolve, reject) => pending.set(id, { resolve, reject }))
}

async function connect() {
  const port = await debuggingPort()
  const version = await (await fetch(`http://127.0.0.1:${port}/json/version`)).json()
  socket = new WebSocket(version.webSocketDebuggerUrl)
  socket.onmessage = ({ data }) => {
    const event = JSON.parse(data)
    if (event.method === 'Network.responseReceived' &&
        /\/api\/|\/organisationen\/|\/settings\//.test(event.params.response.url)) {
      responses.set(event.params.requestId, {
        url: event.params.response.url,
        status: event.params.response.status,
        mimeType: event.params.response.mimeType
      })
    }
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
  await send('Network.enable')
}

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function snapshot() {
  return evaluate(`(() => ({
    url: location.href,
    heading: document.querySelector('h1')?.textContent?.trim() || '',
    text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
    html: document.documentElement.outerHTML,
    fields: [...document.querySelectorAll('input, select, textarea')].map(el => ({name: el.name, value: el.value, type: el.type})),
    links: [...document.querySelectorAll('a')].map(el => ({text: el.textContent.trim(), href: el.href})),
    forms: [...document.querySelectorAll('form')].map(el => ({action: el.action, method: el.method}))
  }))()`)
}

async function waitFor(input) {
  let value
  for (let attempt = 0; attempt < 100; attempt++) {
    try {
      value = await snapshot()
      if ((!input.heading || value.heading === input.heading) &&
          (!input.text || value.text.includes(input.text)) &&
          (!input.urlSuffix || value.url.endsWith(input.urlSuffix)) &&
          (!input.expression || await evaluate(input.expression))) return value
    } catch { /* Navigation temporarily replaces the execution context. */ }
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not reach expected state: ${JSON.stringify({ input, value })}`)
}

async function responseBodies() {
  const values = []
  for (const [requestId, meta] of responses) {
    try {
      const result = await send('Network.getResponseBody', { requestId })
      values.push({ ...meta, body: result.body })
    } catch { /* Chrome can discard cached or redirected response bodies. */ }
  }
  return values
}

async function command(input) {
  if (input.navigate) await send('Page.navigate', { url: input.navigate })
  if (input.click) await evaluate(`(() => {
    const el = [...document.querySelectorAll('a,button,input[type=submit]')]
      .find(el => el.textContent?.trim() === ${JSON.stringify(input.click)} || el.value === ${JSON.stringify(input.click)})
    if (!el) throw new Error('Action missing: ' + ${JSON.stringify(input.click)})
    el.click()
  })()`)
  if (input.submit) await evaluate(`(() => {
    const form = document.querySelector(${JSON.stringify(input.submit)})
    if (!form) throw new Error('Form missing: ' + ${JSON.stringify(input.submit)})
    form.requestSubmit()
  })()`)
  if (input.evaluate) await evaluate(input.evaluate)
  return { ok: true, page: await waitFor(input.wait || {}), responses: input.responses ? await responseBodies() : [] }
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
