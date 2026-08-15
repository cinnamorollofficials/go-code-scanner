import { mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'

const websiteDir = resolve(fileURLToPath(new URL('..', import.meta.url)))
const docsDir = join(websiteDir, 'docs')
const fixtureDir = mkdtempSync(join(tmpdir(), 'go-code-scanner-config-examples-'))
const binaryPath = join(fixtureDir, process.platform === 'win32' ? 'security-review.exe' : 'security-review')
const configKeys = new Set(['project', 'root', 'mode', 'output', 'fail_on', 'scanners', 'profiles', 'offline_profiles', 'policy', 'hooks', 'frontend', 'cache', 'supply_chain', 'governance', 'architecture'])

function markdownFiles(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return entry.name === '.vitepress' ? [] : markdownFiles(path)
    return entry.name.endsWith('.md') ? [path] : []
  })
}

try {
  const build = spawnSync('go', ['build', '-o', binaryPath, '../cmd/security-review'], { cwd: websiteDir, encoding: 'utf8' })
  if (build.status !== 0) throw new Error(build.stderr || build.stdout)
  const errors = []
  let checked = 0
  for (const file of markdownFiles(docsDir)) {
    const name = relative(websiteDir, file)
    for (const [index, match] of [...readFileSync(file, 'utf8').matchAll(/```json(?:[ \t]+[^\n]*)?\n([\s\S]*?)```/g)].entries()) {
      let value
      try { value = JSON.parse(match[1]) } catch { continue }
      if (!value || Array.isArray(value) || !Object.keys(value).some((key) => configKeys.has(key))) continue
      checked++
      const path = join(fixtureDir, `example-${checked}.json`)
      writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`)
      const validation = spawnSync(binaryPath, ['config', 'validate', path], { cwd: websiteDir, encoding: 'utf8' })
      if (validation.status !== 0) errors.push(`${name}: JSON block ${index + 1}: ${validation.stderr.trim() || validation.stdout.trim()}`)
    }
  }
  if (errors.length > 0) {
    console.error(errors.join('\n'))
    process.exitCode = 1
  } else {
    console.log(`Validated ${checked} copyable configuration JSON examples with the current CLI.`)
  }
} finally {
  rmSync(fixtureDir, { recursive: true, force: true })
}
