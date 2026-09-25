import { spawn } from 'node:child_process'
import { createInterface } from 'node:readline'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const directory = await mkdtemp(join(tmpdir(), 'acp-choice-chrome-'))
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

async function evaluate(expression) {
  const result = await send('Runtime.evaluate', { expression, returnByValue: true })
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.text)
  return result.result.value
}

async function snapshot(expected) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const page = await evaluate(`(() => {
      const section = document.querySelector('section[aria-labelledby="provider-openrouter"]')
      const details = [...(section?.querySelectorAll('details') || [])].find(item => item.querySelector('summary')?.textContent.includes('openai/gpt-4'))
      const evidence = details?.querySelector('p')
      const evidenceText = evidence?.textContent?.replace(/\\s+/g,' ').trim() || ''
      const capabilities = Object.fromEntries([...(details?.querySelectorAll('li') || [])].map(item => [
        item.querySelector('strong')?.textContent?.replace(/:$/, '').trim(),
        {text:item.textContent.replace(/\\s+/g,' ').trim(),checkedAt:item.querySelector('time')?.getAttribute('datetime') || ''}
      ]))
      return {url:location.href,text:document.querySelector('main')?.textContent?.replace(/\\s+/g,' ').trim() || '',html:document.querySelector('main')?.outerHTML || '',selected:document.querySelector('section[aria-label="Modellwahlstatus"]')?.textContent?.trim() || '',expanded:details?.open || false,modelEvidence:{source:evidenceText.match(/Quelle:\\s*([^·]+)/)?.[1]?.trim() || '',observedAt:evidence?.querySelector('time')?.getAttribute('datetime') || '',status:evidenceText.includes('Nicht nachgewiesen') ? 'Nicht nachgewiesen' : ''},capabilities}
    })()`)
    if (page.text && (!expected || page.selected.includes(expected))) return page
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error('Browser page did not render')
}

async function command(input) {
  if (input.url) await send('Page.navigate', { url: input.url })
  if (input.provider) await evaluate(`(() => {
    const form = [...document.querySelectorAll('form')].find(item => item.querySelector('input[name=provider]')?.value === ${JSON.stringify(input.provider)})
    if (!form) throw new Error('Anbieterformular fehlt: ${input.provider}')
    for (const [name,value] of Object.entries(${JSON.stringify({ connection_reference: input.connection, model: input.model })})) {
      const field = form.querySelector('[name='+CSS.escape(name)+']')
      if (!field) throw new Error('Formularfeld fehlt: '+name)
      field.value = value
      field.dispatchEvent(new Event('change', {bubbles:true}))
    }
    window.__choiceForm = form
  })()`)
  if (input.submit) await evaluate(`(() => { if (!window.__choiceForm) throw new Error('Modellwahlformular fehlt'); window.__choiceForm.requestSubmit() })()`)
  if (input.expand) await evaluate(`(() => { const details = [...document.querySelectorAll('section[aria-labelledby="provider-openrouter"] details')].find(item => item.querySelector('summary')?.textContent.includes(${JSON.stringify(input.expand)})); if (!details) throw new Error('Fähigkeitsdetails fehlen'); details.querySelector('summary').click() })()`)
  return { ok: true, page: await snapshot(input.submit ? 'OpenRouter' : undefined) }
}

try {
  await connect()
  for await (const line of createInterface({ input: process.stdin })) {
    try { process.stdout.write(JSON.stringify(await command(JSON.parse(line)))+'\n') }
    catch (error) { process.stdout.write(JSON.stringify({ok:false,error:String(error)})+'\n') }
  }
} finally {
  chrome.kill()
  ws?.close()
  await rm(directory, { recursive:true, force:true })
}
