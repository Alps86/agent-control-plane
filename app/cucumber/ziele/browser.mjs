import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-goals-chrome-'))
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
  text: document.querySelector('main')?.textContent?.replace(/\\s+/g, ' ').trim() || '',
  alert: document.querySelector('main [role=alert]')?.textContent?.trim() || '',
  name: document.querySelector('input[name=name]')?.value ?? '',
  links: [...document.querySelectorAll('main a')].map(link => link.textContent.trim()),
  goals: [...document.querySelectorAll('main article h2')].map(item => item.textContent.trim()),
  goal_cards: [...document.querySelectorAll('main [data-goal-id]')].filter(item => item.hasAttribute('data-goal-status')).map(item => ({
    id: item.dataset.goalId, status: item.dataset.goalStatus, parent_id: item.dataset.parentGoalId || '',
    name: item.querySelector('h2')?.textContent?.trim() || ''
  })),
  child_errors: [...document.querySelectorAll('main form[data-action="child"]')].map(form => ({
    id: form.dataset.goalId, message: form.querySelector('[role=alert]')?.textContent?.trim() || ''
  })),
  move_errors: [...document.querySelectorAll('main form[data-action="reparent"]')].map(form => ({
    id: form.dataset.goalId, message: form.querySelector('[role=alert]')?.textContent?.trim() || ''
  })),
  project_paths: [...document.querySelectorAll('main [data-project-id]')].map(item => item.textContent.replace(/\\s+/g, ' ').trim())
}))()`

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function page(mode, previousEpoch) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const value = await evaluate(snapshot)
    const ready = mode === 'error' ? Boolean(value?.alert)
      : mode === 'goals' ? value?.heading === 'Zielübersicht'
      : mode === 'goal-form' ? value?.heading === 'Ziel anlegen'
      : mode === 'project-form' ? value?.heading === 'Projekt anlegen'
      : mode === 'projects' ? value?.heading === 'Projektübersicht'
      : mode === 'org-form' ? value?.heading === 'Organisation anlegen'
      : mode === 'org-list' ? value?.heading === 'Organisationsübersicht'
      : mode === 'org-detail' ? Boolean(value?.url.match(/\/organisationen\/[^/]+$/))
      : Boolean(value?.heading)
    if (ready && (!previousEpoch || value.epoch !== previousEpoch || mode === 'error')) return value
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error(`Browser page did not render ${mode}`)
}

async function command(input) {
  const previousEpoch = input.url || input.click || input.submit ? (await evaluate('performance.timeOrigin')) : undefined
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.click) await evaluate(`(() => {
    const element = document.querySelector(${JSON.stringify(input.click)})
    if (!element) throw new Error('Action missing: ' + ${JSON.stringify(input.click)})
    element.click()
  })()`)
  if (input.check) await evaluate(`(() => {
    if (!document.querySelector(${JSON.stringify(input.check)})) throw new Error('Element missing: ' + ${JSON.stringify(input.check)})
  })()`)
  if (input.fill !== undefined) await evaluate(`(() => {
    const name = document.querySelector(${JSON.stringify(input.field || 'input[name=name]')})
    if (!name) throw new Error('Field missing: ' + ${JSON.stringify(input.field || 'input[name=name]')})
    name.value = ${JSON.stringify(input.fill)}
    name.dispatchEvent(new Event('input', { bubbles: true }))
    name.dispatchEvent(new Event('change', { bubbles: true }))
  })()`)
  if (input.select !== undefined) await evaluate(`(() => {
    const select = document.querySelector(${JSON.stringify(input.field || 'select[name=goal_id]')})
    if (!select) throw new Error('Select missing: ' + ${JSON.stringify(input.field || 'select[name=goal_id]')})
    const option = [...select.options].find(item => item.textContent.trim() === ${JSON.stringify(input.select)})
    if (!option) throw new Error('Option missing: ' + ${JSON.stringify(input.select)})
    select.value = option.value
    select.dispatchEvent(new Event('change', { bubbles: true }))
  })()`)
  if (input.submit) await evaluate(`document.querySelector(${JSON.stringify(input.form || 'main form')})?.requestSubmit()`)
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
