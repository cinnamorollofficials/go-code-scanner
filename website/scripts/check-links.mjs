import { readFileSync, readdirSync, statSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const websiteDir = resolve(fileURLToPath(new URL('..', import.meta.url)))
const outputDir = join(websiteDir, 'docs/.vitepress/dist')
const base = process.env.VITEPRESS_BASE || '/go-code-scanner/'

function files(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    return entry.isDirectory() ? files(path) : [path]
  })
}

function targetFile(pathname) {
  const clean = pathname.replace(/^\//, '')
  const candidates = clean === ''
    ? [join(outputDir, 'index.html')]
    : [join(outputDir, clean), join(outputDir, `${clean}.html`), join(outputDir, clean, 'index.html')]
  return candidates.find((candidate) => {
    try { return statSync(candidate).isFile() } catch { return false }
  })
}

const htmlFiles = files(outputDir).filter((file) => file.endsWith('.html'))
const anchors = new Map(htmlFiles.map((file) => {
  const html = readFileSync(file, 'utf8')
  return [file, new Set([...html.matchAll(/\sid="([^"]+)"/g)].map((match) => match[1]))]
}))
const errors = []
let checked = 0

for (const file of htmlFiles) {
  // The generated monolith is a no-index compatibility target for historical
  // anchors. Canonical catalog and detail pages are checked below like all
  // other rendered pages.
  if (relative(outputDir, file) === 'reference/rules.html') continue
  const html = readFileSync(file, 'utf8')
  const sourcePath = relative(outputDir, file).replace(/index\.html$/, '').replace(/\.html$/, '')
  const sourceURL = new URL(sourcePath, `https://docs.invalid${base}`)
  for (const match of html.matchAll(/\shref="([^"]+)"/g)) {
    const href = match[1].replaceAll('&amp;', '&')
    if (/^(?:https?:|mailto:|tel:|data:|javascript:|\/\/)/.test(href)) continue
    checked++
    const targetURL = new URL(href, sourceURL)
    if (!targetURL.pathname.startsWith(base)) {
      errors.push(`${relative(outputDir, file)}: link escapes documentation base: ${href}`)
      continue
    }
    const pathname = targetURL.pathname.slice(base.length)
    const target = targetFile(pathname)
    if (!target) {
      errors.push(`${relative(outputDir, file)}: missing target for ${href}`)
      continue
    }
    if (targetURL.hash && target.endsWith('.html')) {
      const anchor = decodeURIComponent(targetURL.hash.slice(1))
      if (!anchors.get(target)?.has(anchor)) errors.push(`${relative(outputDir, file)}: missing anchor ${targetURL.hash} in ${relative(outputDir, target)}`)
    }
  }
}

if (errors.length > 0) {
  console.error(errors.join('\n'))
  process.exit(1)
}
console.log(`Checked ${checked} internal links and anchors across ${htmlFiles.length} rendered pages.`)
