import { createHash } from 'node:crypto'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'

const websiteDir = resolve(fileURLToPath(new URL('..', import.meta.url)))
const targets = [
  join(websiteDir, 'docs/reference/rules'),
  join(websiteDir, 'docs/.vitepress/theme/components/RuleCatalogData.js')
]

function files(path) {
  if (!statSync(path).isDirectory()) return [path]
  return readdirSync(path).sort().flatMap((name) => files(join(path, name)))
}

function snapshot() {
  const values = new Map()
  for (const file of targets.flatMap(files)) {
    values.set(relative(websiteDir, file), createHash('sha256').update(readFileSync(file)).digest('hex'))
  }
  return values
}

const before = snapshot()
const generation = spawnSync(process.execPath, ['scripts/generate-rule-catalog.mjs'], { cwd: websiteDir, encoding: 'utf8' })
if (generation.status !== 0) throw new Error(generation.stderr || generation.stdout)
const after = snapshot()
const changed = [...new Set([...before.keys(), ...after.keys()])].filter((name) => before.get(name) !== after.get(name))
if (changed.length > 0) {
  console.error(`Generated Rule Catalog output was stale:\n${changed.map((name) => `- ${name}`).join('\n')}`)
  process.exit(1)
}
console.log(`Generated Rule Catalog output is current (${after.size - 1} detail pages).`)
