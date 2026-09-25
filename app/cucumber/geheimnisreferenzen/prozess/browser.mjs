import { spawn } from 'node:child_process'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const input = JSON.parse(await new Promise(resolve => {
  let data = ''
  process.stdin.on('data', chunk => { data += chunk })
  process.stdin.on('end', () => resolve(data))
}))
const directory = await mkdtemp(join(tmpdir(), 'acp-story79-chrome-'))
const chrome = spawn('google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu',
  '--remote-debugging-port=0', `--user-data-dir=${directory}`, 'about:blank'
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
  throw new Error('Chrome DevTools did not start')
}

function send(method, params = {}, sessionId = session) {
  const id = ++nextID
  socket.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }))
  return new Promise((resolve, reject) => pending.set(id, { resolve, reject }))
}

async function connect() {
  const version = await (await fetch(`http://127.0.0.1:${await port()}/json/version`)).json()
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
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function navigate(url, expectedPath) {
  await send('Page.navigate', { url })
  for (let attempt = 0; attempt < 100; attempt++) {
    const path = await evaluate('location.pathname')
    if (path === expectedPath && await evaluate('document.readyState') === 'complete') return
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Chrome did not navigate to ${expectedPath}`)
}

async function statusPage() {
  await navigate(input.base + '/settings', '/settings')
  const link = await evaluate("document.querySelector('#settings-provider-status-link')?.getAttribute('href')")
  if (link !== '/settings/modellanbieter/status') throw new Error('Settings status link missing')
  await navigate(input.base + link, '/settings/modellanbieter/status')
  return await evaluate(`(() => {
    const card = id => document.querySelector('#connection-' + id)
    const value = (id, index) => card(id)?.querySelectorAll('dd')[index]?.textContent?.trim() || ''
    const auth = id => value(id, 0)
    const status = id => card(id)?.querySelector('[role="status"]')?.textContent?.trim() || ''
    const ref = id => card(id)?.querySelector('code')?.textContent?.trim() || ''
    return { ok: document.querySelector('h1')?.textContent?.trim() === 'Verbindungsstatus',
      codex: status('codex-abo'), openrouter: status('openrouter'),
      codexAuth: auth('codex-abo'), openrouterAuth: auth('openrouter'),
      codexRef: ref('codex-abo'), openrouterRef: ref('openrouter'),
      html: document.documentElement.outerHTML }
  })()`)
}

async function disconnect() {
  await navigate(input.base + '/settings/modellanbieter/openrouter', '/settings/modellanbieter/openrouter')
  const form = await evaluate("Boolean(document.querySelector('form[action=\"/settings/modellanbieter/openrouter/trennen\"]'))")
  if (!form) throw new Error('OpenRouter disconnect form missing')
  await evaluate("document.querySelector('form[action=\"/settings/modellanbieter/openrouter/trennen\"]').requestSubmit()")
  let confirmed = false
  for (let attempt = 0; attempt < 100; attempt++) {
    confirmed = await evaluate("document.querySelector('.provider-status')?.textContent?.trim() === 'nicht eingerichtet' && document.body.textContent.includes('Verbindung getrennt')")
    if (confirmed) break
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  if (!confirmed) throw new Error('disconnect response was not confirmed in Chrome')
  return statusPage()
}

try {
  if (!input.base.startsWith('http://127.0.0.1:')) throw new Error('non-local Settings URL')
  await connect()
  process.stdout.write(JSON.stringify(input.action === 'disconnect' ? await disconnect() : await statusPage()) + '\n')
} catch (error) {
  process.stdout.write(JSON.stringify({ ok: false, error: error.message }) + '\n')
} finally {
  socket?.close()
  chrome.kill()
  await new Promise(resolve => { chrome.once('exit', resolve); setTimeout(resolve, 1000) })
  await rm(directory, { recursive: true, force: true, maxRetries: 5, retryDelay: 100 })
}
