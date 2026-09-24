import { cp, mkdir, rm } from 'node:fs/promises'

await mkdir('dist/templates', { recursive: true })
await mkdir('dist/fixtures', { recursive: true })
await cp('templates', 'dist/templates', { recursive: true })
await cp('fixtures', 'dist/fixtures', { recursive: true })
await cp('node_modules/htmx.org/dist/htmx.min.js', 'dist/assets/htmx.min.js')
await rm('../bridge/dist', { recursive: true, force: true })
await cp('dist', '../bridge/dist', { recursive: true })
