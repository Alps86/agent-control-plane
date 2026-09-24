import { spawn } from 'node:child_process'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const [url, mode] = process.argv.slice(2)
const directory = await mkdtemp(join(tmpdir(), 'story09-chrome-'))
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
  const result = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function waitFor(state) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const current = await evaluate("document.getElementById('codex-status')?.dataset.state")
    if (current === state) return
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  const diagnostics = await evaluate(`({state:document.getElementById('codex-status')?.dataset.state, htmx:typeof htmx, html:document.getElementById('codex-status')?.textContent, ready:document.readyState})`)
  throw new Error(`browser did not reach ${state}: ${JSON.stringify(diagnostics)}`)
}

try {
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
  await send('Page.navigate', { url })
  await waitFor('idle')
  await new Promise(resolve => setTimeout(resolve, 300))
  await evaluate("document.getElementById('codex-start').click()")
  await waitFor('pending')
  const codeShown = await evaluate("!document.getElementById('codex-verification').hidden && !!document.getElementById('codex-user-code').textContent")
  if (!codeShown) throw new Error('start did not show a provider code')
  const probe = await evaluate(`(() => {
    const source = document.getElementById('codex-poller');
    const xhr = {responseText: JSON.stringify({state:'pending', verificationUrl:'https://example.test/verify', userCode:'PROBE-CODE'})};
    source.dispatchEvent(new CustomEvent('htmx:beforeRequest', {bubbles:true, detail:{elt:source, xhr}}));
    source.dispatchEvent(new CustomEvent('htmx:afterRequest', {bubbles:true, detail:{elt:source, xhr, successful:true}}));
    const stale = {responseText: JSON.stringify({state:'pending', verificationUrl:'https://example.test/verify', userCode:'STALE-CODE'})};
    source.dispatchEvent(new CustomEvent('htmx:beforeRequest', {bubbles:true, detail:{elt:source, xhr:stale}}));
    window.story09StaleXHR = stale;
    return document.getElementById('codex-user-code').textContent.trim();
  })()`)
  if (probe !== 'PROBE-CODE') throw new Error('HTMX response handler did not apply serial probe')
  if (mode === 'connected') await evaluate("document.getElementById('codex-refresh').click()")
  else await evaluate("document.getElementById('codex-cancel').click()")
  await waitFor(mode)

  // This status request was assigned a serial before the terminal HTTP request.
  await evaluate(`(() => {
    const source = document.getElementById('codex-poller');
    source.dispatchEvent(new CustomEvent('htmx:afterRequest', {bubbles:true, detail:{elt:source, xhr:window.story09StaleXHR, successful:true}}));
  })()`)
  const result = await evaluate(`(() => {
    const status = document.getElementById('codex-status');
    const code = document.getElementById('codex-user-code');
    return {state:status.dataset.state, text:status.textContent.trim(), codeVisible:!document.getElementById('codex-verification').hidden, code:code.textContent.trim()};
  })()`)
  process.stdout.write(JSON.stringify(result) + '\n')
} catch (error) {
  process.stdout.write(JSON.stringify({ error: error.message }) + '\n')
} finally {
  socket?.close()
  chrome.kill('SIGKILL')
  await new Promise(resolve => chrome.once('exit', resolve))
  await rm(directory, { recursive: true, force: true, maxRetries: 10, retryDelay: 100 })
}
