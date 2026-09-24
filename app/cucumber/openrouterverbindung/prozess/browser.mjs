import { spawn } from 'node:child_process'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const input = JSON.parse(await new Promise(resolve => {
  let data = ''
  process.stdin.on('data', chunk => { data += chunk })
  process.stdin.on('end', () => resolve(data))
}))
const directory = await mkdtemp(join(tmpdir(), 'acp-story23-process-chrome-'))
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

async function view() {
  const result = await send('Runtime.evaluate', {
    expression: `(() => ({ heading: document.querySelector('h1')?.textContent?.trim(),
      status: document.querySelector('.provider-status')?.textContent?.trim(),
      text: document.querySelector('main')?.textContent || '',
      fieldType: document.querySelector('#openrouter-key')?.type,
      fieldValue: document.querySelector('#openrouter-key')?.value,
      html: document.documentElement.outerHTML }))()`,
    returnByValue: true
  })
  return result.result.value
}

async function check() {
  if (!input.url.startsWith('http://127.0.0.1:')) throw new Error('non-local Settings URL')
  await connect()
  await send('Page.navigate', { url: input.url })
  for (let attempt = 0; attempt < 100; attempt++) {
    let page
    try { page = await view() }
    catch { await new Promise(resolve => setTimeout(resolve, 50)); continue }
    if (page.heading === 'OpenRouter' && page.status === 'nicht eingerichtet') {
      if (!page.text.includes('Separate Abrechnung')) throw new Error('billing notice missing')
      if (page.fieldType !== 'password' || page.fieldValue !== '') throw new Error('secret field unsafe')
      if (page.html.includes(input.key)) throw new Error('secret in browser DOM')
      return { ok: true, heading: page.heading, status: page.status, secretField: page.fieldType }
    }
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error('mounted Settings page did not render in Chrome')
}

try {
  process.stdout.write(JSON.stringify(await check()) + '\n')
} catch (error) {
  process.stdout.write(JSON.stringify({ ok: false, error: error.message }) + '\n')
} finally {
  socket?.close()
  chrome.kill()
  await new Promise(resolve => { chrome.once('exit', resolve); setTimeout(resolve, 1000) })
  await rm(directory, { recursive: true, force: true, maxRetries: 5, retryDelay: 100 })
}
