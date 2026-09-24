import { spawn } from 'node:child_process'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const address = process.argv[2]
const width = Number(process.argv[3])
const directory = await mkdtemp(join(tmpdir(), 'acp-experimente-chrome-'))
const chrome = spawn('google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu',
  '--remote-debugging-port=0', `--user-data-dir=${directory}`, 'about:blank'
], { stdio: 'ignore' })
let socket
let session
let id = 0
const pending = new Map()

async function port() {
  for (let attempt = 0; attempt < 100; attempt++) {
    try {
      return Number((await readFile(join(directory, 'DevToolsActivePort'), 'utf8')).split('\n')[0])
    } catch {
      await new Promise(resolve => setTimeout(resolve, 50))
    }
  }
  throw new Error('Chrome DevTools did not start')
}

function send(method, params = {}, sessionId = session) {
  const call = ++id
  socket.send(JSON.stringify({ id: call, method, params, ...(sessionId ? { sessionId } : {}) }))
  return new Promise((resolve, reject) => pending.set(call, { resolve, reject }))
}

async function connect() {
  const host = await port()
  const version = await (await fetch(`http://127.0.0.1:${host}/json/version`)).json()
  socket = new WebSocket(version.webSocketDebuggerUrl)
  socket.onmessage = ({ data }) => {
    const message = JSON.parse(data)
    const item = pending.get(message.id)
    if (!item) return
    pending.delete(message.id)
    if (message.error) item.reject(new Error(message.error.message))
    else item.resolve(message.result)
  }
  await new Promise((resolve, reject) => { socket.onopen = resolve; socket.onerror = reject })
  const target = await send('Target.createTarget', { url: 'about:blank' }, null)
  session = (await send('Target.attachToTarget', { targetId: target.targetId, flatten: true }, null)).sessionId
  await send('Page.enable')
  await send('Runtime.enable')
}

async function snapshot() {
  for (let attempt = 0; attempt < 100; attempt++) {
    const expression = `location.href === ${JSON.stringify(address)} && document.readyState === 'complete'`
    const result = await send('Runtime.evaluate', { expression, returnByValue: true })
    if (result.result.value) break
    if (attempt === 99) throw new Error('Inventarseite wurde im Browser nicht vollständig geladen')
    await new Promise(resolve => setTimeout(resolve, 50))
  }

  const expression = `(async () => {
    const required = ['Paperclip-Reifegrad', 'Nutzen', 'ACP-Abhängigkeiten', 'Entscheidung und Begründung', 'Primärquellen']
    const named = ['Plugin SDK und Manager', 'Isolated Workspaces', 'Cases API', 'Chat-Style Tasks', 'Environments', 'External Objects', 'Status Cards']
    const fragment = await fetch(location.href, { headers: { 'HX-Request': 'true' } }).then(response => response.text())
    const parsed = new DOMParser().parseFromString(fragment, 'text/html')
    const css = await fetch('/assets/index.css').then(response => response.ok)
    const cards = [...document.querySelectorAll('main [data-source-key]')]
    const links = [...document.querySelectorAll('article .experiments-sources a')]
    return {
    title: document.title,
    heading: document.querySelector('main h1')?.textContent?.trim() || '',
    groups: document.querySelectorAll('main [data-group-key]').length,
    entries: cards.length,
    mobileWide: document.documentElement.scrollWidth > document.documentElement.clientWidth,
    source: [...document.querySelectorAll('.experiments-meta span')].map(x => x.textContent.trim()).find(x => x.startsWith('Quellenstand:')) || '',
    fieldsComplete: cards.every(card => required.every(label => [...card.querySelectorAll('dt')].some(node => node.textContent.trim() === label)) && [...card.querySelectorAll('dd')].every(node => node.textContent.trim())),
    statusesComplete: cards.every(card => card.querySelector('.experiments-status')?.textContent.trim() === 'ACP: In ACP noch nicht nachgewiesen'),
    legendComplete: ['Zuordnen', 'Später evaluieren', 'Abgelöst', 'Verweis'].every(label => [...document.querySelectorAll('.experiments-legend dt')].some(node => node.textContent.trim() === label)),
    decisionsComplete: ['zuordnen', 'spaeter-evaluieren', 'abgeloest', 'verweis'].every(kind => document.querySelector('[data-kind="' + kind + '"]')),
    namedComplete: named.every(name => cards.some(card => card.querySelector('h3')?.textContent.trim() === name)),
    sourcesSafe: links.length >= 23 && links.every(link => link.protocol === 'https:' && link.hostname === 'docs.paperclip.ing'),
    relatedVisible: document.querySelector('[data-source-key="api-status-cards-api"] .experiments-related')?.getAttribute('href') === '#source-experimentell-status-cards',
    laterComplete: [...document.querySelectorAll('[data-kind="spaeter-evaluieren"]')].length >= 5 && document.querySelector('.experiments-legend')?.textContent.includes('Folgeentscheid'),
    noFalseClaim: !document.querySelector('main')?.textContent.includes('Bereits in ACP verfügbar') && cards.every(card => card.querySelector('.experiments-status')?.textContent.includes('nicht nachgewiesen')),
    headingsComplete: [...document.querySelectorAll('[data-group-key] h2')].length === 2 && !!document.querySelector('main h1'),
    assetsLoaded: css && [...document.styleSheets].some(sheet => sheet.href?.endsWith('/assets/index.css')),
    fragmentSame: parsed.querySelectorAll('[data-source-key]').length === cards.length && [...parsed.querySelectorAll('[data-source-key]')].every((entry, index) => entry.getAttribute('data-source-key') === cards[index].getAttribute('data-source-key')),
    fragmentNoShell: !fragment.toLowerCase().includes('<!doctype') && !fragment.toLowerCase().includes('<html')
  }})()`
  const result = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true })
  return result.result.value
}

async function closeChrome() {
  if (chrome.exitCode !== null || chrome.signalCode !== null) return
  const exited = new Promise(resolve => chrome.once('exit', resolve))
  if (socket?.readyState === WebSocket.OPEN) {
    try {
      socket.send(JSON.stringify({ id: ++id, method: 'Browser.close' }))
    } catch {
      chrome.kill('SIGTERM')
    }
  } else {
    chrome.kill('SIGTERM')
  }
  const timeout = setTimeout(() => chrome.kill('SIGKILL'), 2000)
  await exited
  clearTimeout(timeout)
}

async function removeProfile() {
  for (let attempt = 0; attempt < 5; attempt++) {
    try {
      await rm(directory, { recursive: true, force: true })
      return
    } catch (error) {
      if (!['ENOTEMPTY', 'EBUSY'].includes(error.code) || attempt === 4) throw error
      await new Promise(resolve => setTimeout(resolve, 100 * (attempt + 1)))
    }
  }
}

try {
  await connect()
  await send('Emulation.setDeviceMetricsOverride', { width, height: 850, deviceScaleFactor: 1, mobile: width <= 600 })
  await send('Page.navigate', { url: address })
  process.stdout.write(JSON.stringify(await snapshot()))
} catch (error) {
  process.stderr.write(`${error.stack || error}\n`)
  process.exitCode = 1
} finally {
  await closeChrome()
  socket?.close()
  await removeProfile()
}
