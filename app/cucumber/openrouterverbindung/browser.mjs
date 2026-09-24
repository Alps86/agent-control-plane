import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-openrouter-chrome-'))
const chrome = spawn('google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu',
  '--remote-debugging-port=0', `--user-data-dir=${directory}`, 'about:blank'
], { stdio: 'ignore' })
let socket
let session
let nextID = 0
const pending = new Map()
let lastMutationHeaders = {}
const requestMethods = new Map()
const nativeOrigins = []

async function debuggingPort() {
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
  socket.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }))
  return new Promise((resolve, reject) => pending.set(id, { resolve, reject }))
}

async function connect() {
  const port = await debuggingPort()
  const version = await (await fetch(`http://127.0.0.1:${port}/json/version`)).json()
  socket = new WebSocket(version.webSocketDebuggerUrl)
  socket.onmessage = ({ data }) => {
    const event = JSON.parse(data)
    if (event.method === 'Network.requestWillBeSent') {
      requestMethods.set(event.params.requestId, event.params.request.method)
    }
    if (event.method === 'Network.requestWillBeSentExtraInfo' && requestMethods.get(event.params.requestId) === 'POST') {
      const headers = event.params.headers || {}
      lastMutationHeaders = { Origin: headers.Origin || headers.origin, Host: headers.Host || headers.host }
      nativeOrigins.push(lastMutationHeaders.Origin)
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
  const answer = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true })
  if (answer.exceptionDetails) throw new Error(answer.exceptionDetails.text)
  return answer.result.value
}

const snapshot = `(() => ({
  heading: document.querySelector('h1')?.textContent?.trim() || '',
  text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  body: document.body?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  status: document.querySelector('.provider-status')?.textContent?.trim() || '',
  secretFieldType: document.querySelector('#openrouter-key')?.type || '',
  secretFieldValue: document.querySelector('#openrouter-key')?.value || '',
  html: document.documentElement.outerHTML,
  url: location.href
}))()`

async function page(status) {
  let last
  for (let attempt = 0; attempt < 100; attempt++) {
    try {
      const result = await evaluate(snapshot)
      last = result
      if (result?.heading === 'OpenRouter' && result.status === status) return result
    } catch {
      // A form submission may replace the page execution context here.
    }
    await new Promise((resolve) => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render status ${status}: ${JSON.stringify({ url: last?.url, status: last?.status, body: last?.body?.slice(0, 200), headers: lastMutationHeaders })}`)
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function noSecret(view, keys) {
  for (const key of keys) assert(!view.html.includes(key), 'Secret appeared in browser DOM')
  assert(view.secretFieldType === 'password', 'Key field is not a password input')
  assert(view.secretFieldValue === '', 'Key field retained a secret after navigation')
}

async function apiBody(keys) {
  const body = await evaluate(`fetch('/api/settings/modellanbieter/openrouter').then(r => r.text())`)
  for (const key of keys) assert(!body.includes(key), 'Secret appeared in public JSON response')
  return JSON.parse(body)
}

async function submit(action, key) {
  const expression = `(() => {
    const form = document.querySelector('form[action=' + ${JSON.stringify(JSON.stringify(action))} + ']')
    if (!form) throw new Error('Missing form: ' + ${JSON.stringify(action)})
    if (${JSON.stringify(key !== undefined)}) form.querySelector('[name=key]').value = ${JSON.stringify(key ?? '')}
    form.requestSubmit()
  })()`
  await evaluate(expression)
}

async function flow(input) {
  assert(input.url?.startsWith('http://127.0.0.1:'), 'Expected loopback Settings URL')
  await send('Page.navigate', { url: input.url })
  let view = await page('nicht eingerichtet')
  assert(view.text.includes('Separate Abrechnung'), 'Separate billing notice missing')
  assert(view.text.includes('ausdrückliche Wahl'), 'Conscious OpenRouter choice missing')
  noSecret(view, [input.key, input.replacement])

  await submit('/settings/modellanbieter/openrouter', input.key)
  view = await page('nicht geprüft')
  noSecret(view, [input.key, input.replacement])
  assert(view.text.includes('openrouter-central'), 'Central reference missing after save')
  await apiBody([input.key, input.replacement])

  await submit('/settings/modellanbieter/openrouter/pruefen')
  view = await page('einsatzbereit')
  noSecret(view, [input.key, input.replacement])
  await apiBody([input.key, input.replacement])

  await submit('/settings/modellanbieter/openrouter', input.replacement)
  view = await page('nicht geprüft')
  noSecret(view, [input.key, input.replacement])
  assert(view.text.includes('openrouter-central'), 'Central reference changed after replacement')
  await apiBody([input.key, input.replacement])

  await submit('/settings/modellanbieter/openrouter/pruefen')
  view = await page('einsatzbereit')
  noSecret(view, [input.key, input.replacement])
  await apiBody([input.key, input.replacement])

  await submit('/settings/modellanbieter/openrouter/trennen')
  view = await page('nicht eingerichtet')
  noSecret(view, [input.key, input.replacement])
  assert(view.text.includes('Verbindung getrennt'), 'Disconnect notice missing')
  const api = await apiBody([input.key, input.replacement])
  assert(api.connected === false, 'Connection survived disconnect')
  assert(nativeOrigins.length === 5, `Expected five native form submissions, got ${nativeOrigins.length}`)
  assert(nativeOrigins.every(origin => origin === new URL(input.url).origin), `Wrong native Origins: ${nativeOrigins}`)
  return { ok: true, billing: true, saved: true, checked: true, replaced: true, disconnected: true, origins: nativeOrigins }
}

try {
  await connect()
  const lines = createInterface({ input: process.stdin })
  for await (const line of lines) {
    try { process.stdout.write(JSON.stringify(await flow(JSON.parse(line))) + '\n') }
    catch (error) { process.stdout.write(JSON.stringify({ ok: false, error: error.message }) + '\n') }
  }
} finally {
  socket?.close()
  chrome.kill()
  await new Promise((resolve) => { chrome.once('exit', resolve); setTimeout(resolve, 1000) })
  await rm(directory, { recursive: true, force: true, maxRetries: 5, retryDelay: 100 })
}
