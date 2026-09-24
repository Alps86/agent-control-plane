import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'settings-chrome-'))
const chrome = spawn('google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu', '--disable-dev-shm-usage',
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

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function waitFor(path) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const page = await evaluate(`(() => ({path:location.pathname, ready:document.readyState, providerTitle:document.querySelector('h1')?.textContent.trim() || '', links:document.querySelectorAll('#settings-codex-link, #settings-openrouter-link').length}))()`)
    if (page.path === path && page.ready === 'complete' && page.providerTitle) return page
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`browser did not reach ${path}`)
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

async function command(input) {
  if (input.action === 'open') {
    await send('Page.navigate', { url: input.url })
    const settings = await waitFor('/settings')
    if (settings.links !== 2) throw new Error('Settings links missing')
    const path = input.id === 'settings-codex-link' ? '/settings/modelle/codex' : '/settings/modellanbieter/openrouter'
    await evaluate(`document.getElementById(${JSON.stringify(input.id)})?.click()`)
    return await waitFor(path)
  }
  if (input.action === 'return') {
    const back = await evaluate("document.querySelector('#codex-settings-back-link, .provider-back')?.getAttribute('href')")
    if (back !== '/settings') throw new Error('provider back link missing')
    await evaluate("document.querySelector('#codex-settings-back-link, .provider-back').click()")
    return await waitFor('/settings')
  }
  throw new Error('unknown browser command')
}

try {
  await connect()
  for await (const line of createInterface({ input: process.stdin })) {
    try { process.stdout.write(JSON.stringify(await command(JSON.parse(line))) + '\n') }
    catch (error) { process.stdout.write(JSON.stringify({ error: error.message }) + '\n') }
  }
} finally {
  socket?.close()
  chrome.kill('SIGKILL')
  await new Promise(resolve => chrome.once('exit', resolve))
  await rm(directory, { recursive: true, force: true, maxRetries: 10, retryDelay: 100 })
}
