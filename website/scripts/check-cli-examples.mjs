import { readFileSync, readdirSync } from 'node:fs'
import { join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const websiteDir = resolve(fileURLToPath(new URL('..', import.meta.url)))
const docsDir = join(websiteDir, 'docs')
const cliSource = readFileSync(resolve(websiteDir, '../cmd/security-review/main.go'), 'utf8')
const sourceFlags = new Set([...cliSource.matchAll(/\.\s*(?:String|Bool|Int|Int64|Duration)\("([^"]+)"/g)].map((match) => match[1]))
const topLevel = new Set((cliSource.match(/usage: security-review <([^>]+)>/)?.[1] || '').split('|'))

const commandFlags = new Map([
  ['scan', ['config', 'root', 'output', 'changed', 'staged', 'ci', 'fail-on', 'quiet', 'verbose', 'color', 'explain', 'profile', 'baseline', 'new-only', 'format', 'fix', 'dry-run', 'scope']],
  ['config validate', []],
  ['baseline create', ['report', 'baseline', 'dry-run']],
  ['baseline update', ['report', 'baseline', 'dry-run', 'accept-resolved']],
  ['baseline status', ['report', 'baseline']],
  ['suppress add', ['suppression-file', 'file', 'line', 'rule', 'fingerprint', 'reason', 'expires', 'ticket', 'approved-by', 'dry-run']],
  ['cache stats', ['dir']],
  ['cache clean', ['dir']],
  ['release archive', ['binary', 'output', 'timestamp']],
  ['release checksums verify', ['manifest', 'directory']],
  ['release provenance generate', ['directory', 'output', 'version', 'commit', 'build-date', 'builder', 'private-key', 'signature']],
  ['release provenance sign', ['provenance', 'private-key', 'output']],
  ['release verify', ['provenance', 'signature', 'public-key', 'directory']],
  ['release changelog validate', ['file']],
  ['upgrade check', ['contract']],
  ['hook install', ['root']],
  ['hook uninstall', ['root']],
  ['hook status', ['root']],
  ['hook run', ['root', 'file']],
  ['version', []]
])

for (const flags of commandFlags.values()) {
  for (const flag of flags) if (!sourceFlags.has(flag)) throw new Error(`Checker contains flag --${flag}, but current CLI source does not`)
}

function markdownFiles(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return entry.name === '.vitepress' ? [] : markdownFiles(path)
    return entry.name.endsWith('.md') ? [path] : []
  })
}

function shellBlocks(content) {
  return [...content.matchAll(/```(?:sh|bash|shell)(?:[ \t]+[^\n]*)?\n([\s\S]*?)```/g)].map((match) => match[1])
}

function logicalCommands(block) {
  const commands = []
  let current = ''
  for (const raw of block.split('\n')) {
    const line = raw.trim()
    if (!line || line.startsWith('#')) continue
    current += `${current ? ' ' : ''}${line.replace(/\\$/, '').trim()}`
    if (!line.endsWith('\\')) {
      commands.push(current)
      current = ''
    }
  }
  if (current) commands.push(current)
  return commands
}

const errors = []
let checked = 0
for (const file of markdownFiles(docsDir)) {
  const name = relative(websiteDir, file)
  for (const block of shellBlocks(readFileSync(file, 'utf8'))) {
    for (const line of logicalCommands(block)) {
      const start = line.match(/^security-review\s+(.+)$/)
      if (!start) continue
      checked++
      const args = start[1].split(/\s+/)
      if (!topLevel.has(args[0])) {
        errors.push(`${name}: unsupported top-level command in "${line}"`)
        continue
      }
      const candidates = [...commandFlags.keys()].filter((prefix) => start[1] === prefix || start[1].startsWith(`${prefix} `)).sort((a, b) => b.length - a.length)
      const prefix = candidates[0]
      if (!prefix) {
        errors.push(`${name}: unsupported command shape in "${line}"`)
        continue
      }
      const allowed = new Set(commandFlags.get(prefix))
      for (const match of line.matchAll(/--([a-z0-9-]+)(?:=|\s|$)/g)) {
        if (!allowed.has(match[1])) errors.push(`${name}: --${match[1]} is not valid for "${prefix}"`)
      }
    }
  }
}

if (errors.length > 0) {
  console.error(errors.join('\n'))
  process.exit(1)
}
console.log(`Checked ${checked} documented security-review commands against current command and flag definitions.`)
