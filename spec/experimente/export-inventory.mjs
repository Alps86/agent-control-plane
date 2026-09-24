#!/usr/bin/env node
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';

const directory = fileURLToPath(new URL('.', import.meta.url));
const sourcePath = resolve(directory, 'inventar.md');
const targetPath = resolve(directory, '../../app/internal/adapter/web/experimente/inventory.json');
const markdown = readFileSync(sourcePath, 'utf8');
const sourceSha256 = createHash('sha256').update(markdown).digest('hex');
const sourceAsOf = markdown.match(/Stand der Primärquellenprüfung: (\d{2}\.\d{2}\.\d{4})\./)?.[1];
if (!sourceAsOf) throw Error('Quellenstand fehlt');

const benefitStarts = new Map([
  ['Cases API', 'Dokumentenreiche'],
  ['Attention API', 'lesender'],
  ['Decision Training API', 'unveränderliche'],
  ['Summary Slots API', 'Slots halten'],
  ['Resource Memberships API', 'persönliche Sidebar-Auswahl'],
  ['Folders API', 'getrennte, verschachtelte'],
  ['Inbox Agent Policy API', 'persönliche Richtlinie'],
  ['Status Cards API', 'identischer Flag'],
  ['Plugins API', 'REST-Schnittstelle'],
]);

const dependencies = new Map([
  ['Plugin SDK und Manager', ['INT-05']],
  ['Environments', ['RUN-01, PROJ-03, PERM-04']],
  ['Isolated Workspaces', ['PROJ-03 und RUN-01']],
  ['Experimental File Viewer', ['ART-01/02 und PROJ-03']],
  ['External Objects', ['TASK-05/ART-01 und INT-01–05']],
  ['Task Plan Decomposition Panel', ['TASK-03 / REVIEW-02']],
  ['Chat-Style Tasks', ['TASK-05/07, ACT-02, Work Modes und die geplanten Teamchats']],
  ['Task Watchdogs', ['TASK-03, RUN-/SCH-Lebenszyklus, Agentenrechte und Eskalation']],
  ['Status Cards', ['MON-01/02', 'MS-05/06, Rechte und Aktualisierung']],
  ['Auto-Create Recovery Tasks', ['TASK-03, SCH-01–06, Run-/Blockerzustand und Deduplizierung']],
  ['Server Info Debug View', ['OPS-01']],
  ['Auto-Restart Dev Server When Idle', ['ACP ist Go/Vite']],
  ['Connections v3 (Apps)', ['INT-01–05/PERM-02/05', 'CHV-/INT-Stories']],
  ['Cloud Sync', ['OPS-04', 'OPS-03']],
  ['Cases API', ['ORG-01, TASK-01 und ART-01']],
  ['Attention API', ['VIEW-02']],
  ['Decision Training API', ['Story-87', 'Story-48']],
  ['Summary Slots API', ['PROJ-01/03, ART-01 und Modellroute']],
  ['Resource Memberships API', ['Story-52', 'Story-65']],
  ['Folders API', ['Story-63', 'Story-69']],
  ['Inbox Agent Policy API', ['Story-51']],
  ['Status Cards API', ['Status-Cards-Evaluation']],
  ['Plugins API', ['Plugin-Evaluation']],
]);

const plain = value => value.replace(/\*\*/g, '').replace(/`/g, '').replace(/\[([^\]]+)\]\([^)]+\)/g, '$1').trim();
const slug = value => value.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
const cells = line => line.split('|').slice(1, -1).map(value => value.trim());
const headings = [
  ['experimentell', 'Experimentelle Bedienflächen und Laufzeitfunktionen', 14],
  ['api', 'Weitere in der API-Navigation sichtbare Flächen', 9],
];

const groups = headings.map(([key, heading, expected]) => {
  const section = markdown.split(`## ${heading}\n`)[1]?.split('\n## ')[0];
  if (!section) throw Error(`Gruppe fehlt: ${heading}`);

  const rows = section.split('\n').filter(line => line.startsWith('| ') && !line.startsWith('| ---'));
  rows.shift();
  if (rows.length !== expected) throw Error(`${heading}: ${rows.length} statt ${expected} Zeilen`);

  const entries = rows.map(line => {
    const columns = cells(line);
    if (columns.length !== (key === 'api' ? 3 : 4)) throw Error(`Spalten fehlen: ${line}`);

    const sourceLinks = [...columns[0].matchAll(/\[([^\]]+)\]\((https:\/\/[^)]+)\)/g)].map(([, label, url]) => ({ label, url }));
    if (!sourceLinks.length || sourceLinks.some(link => {
      const url = new URL(link.url);
      return url.origin !== 'https://docs.paperclip.ing' || url.username || url.password || url.hash;
    })) throw Error(`Ungültige Quelle: ${line}`);

    const name = sourceLinks[0].label;
    const sourceKey = `${key}-${slug(name)}`;
    const combined = columns[1];
    const splitAt = key === 'api' ? combined.indexOf(benefitStarts.get(name) ?? '') : -1;
    if (key === 'api' && (!benefitStarts.has(name) || splitAt < 1)) throw Error(`Nutzen nicht trennbar: ${name}`);

    const paperclipMaturity = plain(key === 'api' ? combined.slice(0, splitAt) : combined).replace(/[;, ]+$/, '');
    const benefit = plain(key === 'api' ? combined.slice(splitAt) : columns[2]);
    const decisionText = columns[key === 'api' ? 2 : 3];
    const decisionPlain = plain(decisionText);
    const references = dependencies.get(name);
    if (!references || references.some(reference => !decisionPlain.includes(reference))) throw Error(`Abhängigkeiten nicht belegt: ${name}`);
    const decisions = [...decisionText.matchAll(/\*\*([^*]+)\*\*/g)]
      .map(([, label]) => kindOf(label)).filter(Boolean);
    const decisionParts = (decisions.length ? decisions : ['verweis'])
      .map(kind => ({ kind, reason: decisionPlain }));

    return {
      sourceKey, name, sourceLinks, paperclipMaturity, benefit,
      acpDependencies: references.join('; '), decisionParts,
      acpImplementationStatus: 'In ACP noch nicht nachgewiesen',
      ...(name === 'Status Cards API' ? { relatedSourceKey: 'experimentell-status-cards' } : {}),
      ...(name === 'Plugins API' ? { relatedSourceKey: 'experimentell-plugin-sdk-und-manager' } : {}),
    };
  });
  return { key, heading, entries };
});

const allEntries = groups.flatMap(group => group.entries);
if (new Set(allEntries.map(entry => entry.sourceKey)).size !== allEntries.length) throw Error('Quellkennungen nicht eindeutig');
if (new Set(allEntries.map(entry => entry.name)).size !== allEntries.length) throw Error('Quellnamen nicht eindeutig');
if (allEntries.length !== dependencies.size) throw Error('Quellenmenge und Abhängigkeitsregister weichen ab');

function kindOf(label) {
  const lower = label.toLowerCase();
  if (lower.startsWith('später')) return 'spaeter-evaluieren';
  if (lower.startsWith('abgelöst')) return 'abgeloest';
  if (lower.startsWith('zuordnen') || lower.startsWith('root-zuordnung')) return 'zuordnen';
  return null;
}

const inventory = { sourceAsOf, sourceSha256, groups };
const output = `${JSON.stringify(inventory, null, 2)}\n`;
if (process.argv.includes('--check')) {
  if (readFileSync(targetPath, 'utf8') !== output) throw Error('inventory.json ist nicht mit inventar.md synchron');
  console.log(`Inventar synchron: ${groups.reduce((count, group) => count + group.entries.length, 0)} Zeilen, SHA-256 ${sourceSha256}`);
} else {
  writeFileSync(targetPath, output);
  console.log(`Inventar erzeugt: ${targetPath}`);
}
