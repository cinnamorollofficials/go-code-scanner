import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'
import { presets, validateConfig } from '../docs/.vitepress/theme/components/ConfigBuilderData.js'

const websiteDir = resolve(fileURLToPath(new URL('..', import.meta.url)))
const fixtureDir = mkdtempSync(join(tmpdir(), 'go-code-scanner-presets-'))
const binaryName = process.platform === 'win32' ? 'security-review.exe' : 'security-review'
const binaryPath = join(fixtureDir, binaryName)

try {
  const build = spawnSync('go', ['build', '-o', binaryPath, '../cmd/security-review'], {
    cwd: websiteDir,
    encoding: 'utf8'
  })
  if (build.status !== 0) {
    throw new Error(`Could not build configuration validator:\n${build.stderr || build.stdout}`)
  }

  for (const [name, config] of Object.entries(presets)) {
    const browserErrors = validateConfig(config)
    if (browserErrors.length > 0) {
      throw new Error(`${name}: browser validation failed:\n- ${browserErrors.join('\n- ')}`)
    }
    const configPath = join(fixtureDir, `${name}.json`)
    writeFileSync(configPath, `${JSON.stringify(config, null, 2)}\n`)
    const validation = spawnSync(binaryPath, ['config', 'validate', configPath], {
      cwd: websiteDir,
      encoding: 'utf8'
    })
    if (validation.status !== 0) {
      throw new Error(`${name}: CLI validation failed:\n${validation.stderr || validation.stdout}`)
    }
  }

  console.log(`Validated ${Object.keys(presets).length} Config Builder presets with browser and CLI validators.`)
} finally {
  rmSync(fixtureDir, { recursive: true, force: true })
}
