import htmx from 'htmx.org'
import './style.css'

const views = ['organization', 'projects', 'agents', 'tasks']
const params = new URLSearchParams(location.search)
const view = views.includes(params.get('view')) ? params.get('view') : 'organization'
const fixture = ['empty', 'error', 'alternate'].includes(params.get('fixture')) ? params.get('fixture') : 'default'
const file = fixture === 'default' ? view : `${view}-${fixture}`
const root = document.querySelector('#app')

const escape = (value) => String(value ?? '').replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char])
const link = (target) => `/?view=${views.includes(target) ? target : 'organization'}${fixture === 'default' ? '' : `&fixture=${fixture}`}`
const initials = (item) => escape(item.Initials || item.Code || '·')
const badge = (label, tone = '') => `<span class="badge ${escape(tone)}">${escape(label)}</span>`

function shell(data) {
  const organization = data.View.Organization || {}
  const navigation = Array.isArray(data.Navigation) ? data.Navigation.filter((item) => item && views.includes(item.Key)) : []
  document.title = data.PageTitle || 'Agent Control Plane'
  root.innerHTML = `
    <div class="app-shell">
      <aside class="sidebar" aria-label="Hauptnavigation">
        <a class="brand" href="${link('organization')}"><span class="brand-mark">✳</span><span>agent<span class="brand-light">control</span><small>CONTROL PLANE</small></span></a>
        <div class="side-label">ARBEITSBEREICH</div>
        <div class="workspace"><span class="workspace-logo">${initials(organization)}</span><span class="workspace-copy"><strong>${escape(organization.Name || 'Organisation')}</strong><small>Organisation</small></span><span class="chevron">⌄</span></div>
        <div class="side-label nav-label">NAVIGATION</div>
        <nav>${navigation.map((item) => `<a href="${escape(link(item.Key))}" class="nav-item ${item.Key === view ? 'active' : ''}" ${item.Key === view ? 'aria-current="page"' : ''}><span class="nav-icon" aria-hidden="true">${escape(item.Icon)}</span>${escape(item.Label)}${item.Key === 'tasks' && organization.ActiveTasks ? `<span class="nav-count">${escape(organization.ActiveTasks)}</span>` : ''}</a>`).join('')}</nav>
        <div class="sidebar-bottom"><div class="preview-indicator"><span class="live-dot"></span> JSON-Vorschau</div><button class="sidebar-help" hx-get="/fragments/preview-help.html" hx-target="#preview-help" hx-swap="innerHTML">ⓘ &nbsp; Über diese Vorschau</button><div id="preview-help"></div></div>
      </aside>
      <div class="main-wrap"><header class="topbar"><div class="breadcrumb">Arbeitsbereich <span>/</span> <strong>${escape(navigation.find((item) => item.Key === view)?.Label || 'Übersicht')}</strong></div><div class="top-actions"><span class="preview-pill">● &nbsp; Vorschau</span><span class="avatar-user" aria-label="Betreiber">B</span></div></header><main id="content" class="content" tabindex="-1">${content(data)}</main></div>
    </div>`
  htmx.process(root)
}

function intro(eyebrow, title, description, action = '') {
  return `<div class="page-heading"><div><div class="eyebrow">${escape(eyebrow)}</div><h1>${escape(title)}</h1><p>${escape(description)}</p></div>${action}</div>`
}

function emptyState(title, description, symbol) {
  return `<section class="empty-state"><div class="empty-icon">${symbol}</div><h2>${escape(title)}</h2><p>${escape(description)}</p><a href="${link('organization')}" class="text-link">Zur Übersicht <span aria-hidden="true">→</span></a></section>`
}

function organization(data) {
  const item = data.View.Organization
  const highlights = data.View.Highlights || []
  const recent = data.View.RecentTasks || []
  const heading = intro('ORGANISATION · ' + (item.Update || 'Vorschau'), 'Guten Tag, ' + item.Name, item.Summary || item.Description)
  return `${heading}
    <section class="hero"><div><div class="hero-kicker">IHRE ORGANISATION</div><h2>${escape(item.Name)}</h2><p>${escape(item.Description)}</p><a class="hero-link" href="${link('projects')}">Projekte ansehen <span>↗</span></a></div><div class="hero-decoration" aria-hidden="true"><span>✳</span></div></section>
    <div class="section-heading"><h2>Auf einen Blick</h2><span>Aktueller Stand</span></div>
    <div class="stats-grid">${highlights.length ? highlights.map((metric, index) => `<article class="stat-card"><div class="stat-top"><span class="stat-icon tone-${index}">${['▦','◉','☷'][index]}</span><span class="stat-arrow">↗</span></div><strong>${escape(metric.Value)}</strong><div class="stat-label">${escape(metric.Label)}</div><small>${escape(metric.Description)}</small></article>`).join('') : `<article class="stat-card"><strong>0</strong><div class="stat-label">Noch keine Einträge</div><small>Ihre Arbeitsbereiche erscheinen hier.</small></article>`}</div>
    <div class="section-heading recent-heading"><h2>Aktuelle Aufgaben</h2><a href="${link('tasks')}" class="text-link">Alle Aufgaben ansehen →</a></div>
    ${recent.length ? `<div class="table-card"><div class="table-scroll"><table><thead><tr><th>AUFGABE</th><th>PROJEKT</th><th>STATUS</th><th>ZUSTÄNDIG</th></tr></thead><tbody>${recent.map((task) => `<tr><td><span class="task-code">${escape(task.Code)}</span><strong>${escape(task.Title)}</strong></td><td>${escape(task.Project)}</td><td>${badge(task.Status, task.Status === 'In Arbeit' ? 'blue' : task.Status === 'In Prüfung' ? 'violet' : '')}</td><td><span class="small-avatar">${initials(task)}</span>${escape(task.Assignee)}</td></tr>`).join('')}</tbody></table></div></div>` : emptyState('Noch keine Aufgaben', 'Sobald Aufgaben angelegt werden, sehen Sie sie hier.', '☷')}`
}

function projects(data) {
  const items = data.View.Projects || []
  return `${intro('ARBEITSBEREICH / PROJEKTE', 'Projekte', 'Alle Vorhaben Ihrer Organisation an einem Ort.', `<span class="heading-count">${items.length} Projekte</span>`)}
    ${items.length ? `<div class="project-grid">${items.map((item) => `<article class="project-card"><div class="project-top"><div class="project-icon ${escape(item.Color)}">${escape(item.Code)}</div>${badge(item.Status, item.Status === 'Aktiv' ? 'green' : '')}</div><h2>${escape(item.Name)}</h2><p>${escape(item.Description)}</p><div class="project-progress"><div><span>Fortschritt</span><strong>${escape(item.Progress)}%</strong></div><div class="progress-track"><span style="width:${Math.max(0,Math.min(100,Number(item.Progress)||0))}%"></span></div></div><div class="project-footer"><span>☷ &nbsp; ${escape(item.Tasks)} Aufgaben</span><span aria-hidden="true">↗</span></div></article>`).join('')}</div>` : emptyState('Noch keine Projekte', 'Für diese Organisation sind noch keine Projekte vorhanden.', '▦')}`
}

function agents(data) {
  const items = data.View.Agents || []
  return `${intro('ARBEITSBEREICH / AGENTEN', 'Agenten', 'Ihr Team und seine Aufgabenbereiche im Überblick.', `<span class="heading-count">${items.length} Agenten</span>`)}
    ${items.length ? `<div class="agent-grid">${items.map((item) => `<article class="agent-card"><div class="agent-card-top"><div class="agent-avatar ${escape(item.Color)}">${initials(item)}</div>${badge(item.Status, item.Status === 'Bereit' ? 'green' : item.Status === 'In Arbeit' ? 'blue' : '')}</div><h2>${escape(item.Name)}</h2><div class="agent-role">${escape(item.Role)}</div><p>${escape(item.Description)}</p><div class="agent-footer"><span class="type-pill">${escape(item.Type)}</span><span>Profilvorschau</span></div></article>`).join('')}</div>` : emptyState('Noch keine Agenten', 'Sobald Agenten eingerichtet sind, erscheinen sie hier.', '◉')}`
}

function tasks(data) {
  const items = data.View.Tasks || []
  return `${intro('ARBEITSBEREICH / AUFGABEN', 'Aufgaben', 'Alles, woran Ihr Team gerade arbeitet.', `<span class="heading-count">${items.length} Aufgaben</span>`)}
    ${data.Errors?.length ? `<section class="error-state" role="alert"><div class="error-icon">!</div><div><h2>Aufgaben nicht verfügbar</h2><p>${escape(data.Errors[0])}</p><a href="${link('organization')}" class="text-link">Zur Übersicht →</a></div></section>` : items.length ? `<div class="task-toolbar"><div><span class="toolbar-tab active">Alle Aufgaben <b>${items.length}</b></span></div><span class="toolbar-caption">Nach Status sortiert</span></div><div class="table-card"><div class="table-scroll"><table><thead><tr><th>AUFGABE</th><th>PROJEKT</th><th>PRIORITÄT</th><th>STATUS</th><th>ZUSTÄNDIG</th></tr></thead><tbody>${items.map((item) => `<tr><td><span class="task-code">${escape(item.Code)}</span><strong>${escape(item.Title)}</strong></td><td>${escape(item.Project)}</td><td><span class="priority ${item.Priority === 'Hoch' ? 'high' : ''}">● &nbsp;${escape(item.Priority)}</span></td><td>${badge(item.Status, item.Status === 'In Arbeit' ? 'blue' : item.Status === 'In Prüfung' ? 'violet' : '')}</td><td><span class="small-avatar">${initials(item)}</span>${escape(item.Assignee)}</td></tr>`).join('')}</tbody></table></div></div>` : emptyState('Noch keine Aufgaben', 'In dieser Organisation gibt es derzeit keine Aufgaben.', '☷')}`
}

function content(data) {
  const notice = data.Notice ? `<div class="notice" role="status">${escape(data.Notice)}</div>` : ''
  return notice + ({ organization, projects, agents, tasks })[view](data)
}

fetch(`/fixtures/${file}.json`)
  .then((response) => { if (!response.ok) throw new Error('Fixture unavailable'); return response.json() })
  .then(shell)
  .catch(() => { root.innerHTML = `<main class="load-error" role="alert"><h1>Vorschau nicht verfügbar</h1><p>Die Beispieldaten konnten nicht geladen werden.</p><a href="/?view=organization">Zur Übersicht</a></main>` })
