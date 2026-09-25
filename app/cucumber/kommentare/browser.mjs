import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-comments-chrome-'))
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
    if (event.error) { item.reject(new Error(event.error.message)); return }
    item.resolve(event.result)
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
  text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  alert: document.querySelector('main [role=alert]')?.textContent?.trim() || '',
  title: document.querySelector('input[name=title]')?.value ?? '',
  description: document.querySelector('textarea[name=description]')?.value ?? '',
  priority: document.querySelector('select[name=priority]')?.value ?? '',
  assignee: document.querySelector('select[name=assignee_id]')?.value ?? '',
  assigneeMultiple: document.querySelector('select[name=assignee_id]')?.multiple ?? false,
  assigneeSelectCount: document.querySelectorAll('select[name=assignee_id]').length,
  assigneeOptions: [...document.querySelectorAll('select[name=assignee_id] option')].map(option => option.textContent.trim()),
  titleError: document.querySelector('#aufgabe-title-error')?.textContent?.trim() || '',
  assigneeError: document.querySelector('#aufgabe-assignee-error')?.textContent?.trim() || '',
  projectError: document.querySelector('#aufgabe-project-error')?.textContent?.trim() || '',
  projectName: document.querySelector('input[name=project_name]')?.value ?? '',
  projectReadOnly: document.querySelector('input[name=project_name]')?.readOnly ?? false,
  taskLinks: [...document.querySelectorAll('main a[href*="/aufgaben/"]')].map(link => link.textContent.trim()),
  timestamps: [...document.querySelectorAll('main .comment time[datetime]')].map(time => ({ value: time.getAttribute('datetime'), text: time.textContent.trim() })),
  commentCount: document.querySelectorAll('main .comment').length,
  links: [...document.querySelectorAll('main .comment a[href]')].map(link => ({ href: link.getAttribute('href'), text: link.textContent.trim() }))
}))()`

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function page(mode, previousEpoch) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    const ready = mode === 'project-detail' ? value?.heading === 'Website'
      : mode === 'task-form' ? value?.heading === 'Aufgabe anlegen'
      : mode === 'task-detail' ? value?.heading === 'Startseite prüfen'
      : mode === 'task-list' ? value?.heading === 'Aufgaben von Website'
      : mode === 'form-error' ? value?.heading === 'Aufgabe anlegen' && Boolean(value?.alert)
      : mode === 'error' ? Boolean(value?.alert) || value?.heading === 'Aufgaben'
      : Boolean(value?.heading)
    if (ready && (!previousEpoch || value.epoch !== previousEpoch)) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render ${mode}`)
}

async function command(input) {
  const previousEpoch = input.url || input.click || input.submit ? await evaluate('performance.timeOrigin') : undefined
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.click) await evaluate(`(() => {
    const element = document.querySelector(${JSON.stringify(input.click)})
    if (!element) throw new Error('Action missing: ' + ${JSON.stringify(input.click)})
    element.click()
  })()`)
  if (input.fill !== undefined) await evaluate(`(() => {
    const field = document.querySelector(${JSON.stringify(input.field)})
    if (!field) throw new Error('Field missing: ' + ${JSON.stringify(input.field)})
    field.value = ${JSON.stringify(input.fill)}
    field.dispatchEvent(new Event('input', { bubbles: true }))
    field.dispatchEvent(new Event('change', { bubbles: true }))
  })()`)
  if (input.select !== undefined) await evaluate(`(() => {
    const field = document.querySelector(${JSON.stringify(input.field)})
    if (!field) throw new Error('Select missing: ' + ${JSON.stringify(input.field)})
    field.value = ${JSON.stringify(input.select)}
    field.dispatchEvent(new Event('change', { bubbles: true }))
  })()`)
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
