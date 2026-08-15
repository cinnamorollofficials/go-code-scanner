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
  .replace('title: Scanner & Adapter Compatibility Reference', 'title: Scanner and Adapter Compatibility')
  .replace('description: Complete list of built-in scanners, external adapters, network requirements, and parser formats.', 'description: "For configuration authors: compare built-in scanners and external adapters by domain, network use, and parser format."')
  .replace('# Scanner & Adapter Compatibility Reference', '# Scanner and Adapter Compatibility')
  .replace('| Adapter ID | Executable | Domain | Offline Compatible | Parser Format | Description |', '| Adapter ID | Executable | Domain | Offline compatible | Parser format | Description |')
  .replaceAll('Yes 🔒', 'Yes')
  .replaceAll('No 🌐', 'No (network required)')
  .replace('Comprehensive vulnerability scanner for containers and dependencies.', 'Vulnerability scanner for containers and dependencies.')
  .replaceAll('| `vulnerabilities` |', '| `supply_chain` |')
  .replaceAll('| `secrets` |', '| `security` |')
  .replaceAll('| `frontend` |', '| `quality` |')
writeFileSync(scannersPath, normalizedScanners)

const configPath = join(websiteDir, 'docs/reference/config/generated-reference.md')
const normalizedConfig = readFileSync(configPath, 'utf8')
  .replace('title: Complete Field Reference', 'title: Configuration Field Reference')
  .replace('description: Automatically generated configuration reference tables from Go struct definitions.', 'description: "For configuration authors: look up schema fields, types, defaults, and requirements generated from Go definitions."')
  .replace('# Complete Configuration Field Reference', '# Configuration Field Reference')
writeFileSync(configPath, normalizedConfig)

const rulesPath = join(websiteDir, 'docs/reference/rules.md')
const normalizedRules = readFileSync(rulesPath, 'utf8')
  .replace('title: Rule Catalog', 'title: Legacy Rule Reference')
  .replace("description: Complete catalog of default built-in security, secret, governance, and quality rules with Do's and Don'ts code examples.", 'description: "For maintainers preserving legacy anchors: inspect the generated monolithic rule reference and its remediation examples."')
  .replace('# Built-In Rule Catalog', '# Legacy Rule Reference')
  .replace("Below is the complete catalog of built-in detection rules provided by `security-review`. This catalog includes detailed guidance, recommendations, and **Do's and Don'ts** code examples for each rule.", 'This generated compatibility page preserves historical rule anchors. Use the [Rule Catalog](/reference/rule-catalog) to search rules and open focused remediation pages.')
  .replaceAll('### Details & Guidance', '### Details and Guidance')
  .replaceAll("##### Code Examples (Don't vs Do)", '##### Unsafe and Safer Examples')
  .replaceAll("##### Code Example (Don't vs Do)", '##### Unsafe and Safer Example')
  .replaceAll("// ❌ Don't (Unsafe)", '// Unsafe example')
  .replaceAll("// ✅ Do (Recommended)", '// Safer example')
  .replaceAll("# ❌ Don't (Unsafe)", '# Unsafe example')
  .replaceAll("# ✅ Do (Recommended)", '# Safer example')
writeFileSync(rulesPath, normalizedRules)

const after = snapshot()
const changed = [...after.keys()].filter((name) => before.get(name) !== after.get(name))
if (changed.length > 0) {
  console.error(`Generated source-backed references were stale and have been refreshed:\n${changed.map((name) => `- ${name}`).join('\n')}\nReview the generated changes, then run the check again.`)
  process.exit(1)
}

console.log(`Source-backed configuration, rule, and normalized scanner references are current (${after.size} files).`)
