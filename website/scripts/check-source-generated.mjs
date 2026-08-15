import { createHash } from 'node:crypto'
import { readFileSync, writeFileSync } from 'node:fs'
import { join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'

const websiteDir = resolve(fileURLToPath(new URL('..', import.meta.url)))
const repositoryDir = resolve(websiteDir, '..')
const targets = [
  join(websiteDir, 'docs/reference/config/generated-reference.md'),
  join(websiteDir, 'docs/reference/rules.md'),
  join(websiteDir, 'docs/reference/scanners.md')
]

function snapshot() {
  return new Map(targets.map((file) => [
    relative(websiteDir, file),
    createHash('sha256').update(readFileSync(file)).digest('hex')
  ]))
}

const before = snapshot()
const generation = spawnSync('go', ['run', './cmd/gen-all-docs/main.go'], {
  cwd: repositoryDir,
  encoding: 'utf8'
})

if (generation.status !== 0) {
  throw new Error(`Could not regenerate source-backed references:\n${generation.stderr || generation.stdout}`)
}

// The scanner registry groups tools by analyzer family, while the public
// documentation contract uses canonical policy domains. Keep this website-only
// normalization adjacent to the generated artifact until the source generator
// exposes policy domains directly.
const scannersPath = join(websiteDir, 'docs/reference/scanners.md')
const normalizedScanners = readFileSync(scannersPath, 'utf8')
  .replaceAll('| `vulnerabilities` |', '| `supply_chain` |')
  .replaceAll('| `secrets` |', '| `security` |')
  .replaceAll('| `frontend` |', '| `quality` |')
writeFileSync(scannersPath, normalizedScanners)

const after = snapshot()
const changed = [...after.keys()].filter((name) => before.get(name) !== after.get(name))
if (changed.length > 0) {
  console.error(`Generated source-backed references were stale and have been refreshed:\n${changed.map((name) => `- ${name}`).join('\n')}\nReview the generated changes, then run the check again.`)
  process.exit(1)
}

console.log(`Source-backed configuration, rule, and normalized scanner references are current (${after.size} files).`)
