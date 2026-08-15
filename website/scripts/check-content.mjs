import { readFileSync, readdirSync } from 'node:fs'
import { join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const websiteDir = resolve(fileURLToPath(new URL('..', import.meta.url)))
const docsDir = join(websiteDir, 'docs')

function markdownFiles(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return entry.name === '.vitepress' ? [] : markdownFiles(path)
    return entry.name.endsWith('.md') ? [path] : []
  })
}

function outsideFences(content) {
  let fenced = false
  return content.split('\n').filter((line) => {
    if (line.startsWith('```')) {
      fenced = !fenced
      return false
    }
    return !fenced
  }).join('\n')
}

const errors = []
const titles = new Map()
const files = markdownFiles(docsDir)

for (const file of files) {
  const name = relative(websiteDir, file)
  const content = readFileSync(file, 'utf8')
  const frontmatter = content.match(/^---\n([\s\S]*?)\n---\n/)
  if (!frontmatter) {
    errors.push(`${name}: missing YAML frontmatter`)
    continue
  }
  const rawTitle = frontmatter[1].match(/^title:\s*(.+)$/m)?.[1]?.trim()
  const rawDescription = frontmatter[1].match(/^description:\s*(.+)$/m)?.[1]?.trim()
  const quotedScalar = (value) => (value?.startsWith('"') && value.endsWith('"')) || (value?.startsWith("'") && value.endsWith("'"))
  for (const [field, value] of [['title', rawTitle], ['description', rawDescription]]) {
    if (value && value.includes(': ') && !quotedScalar(value)) {
      errors.push(`${name}: frontmatter ${field} containing a colon must be quoted as valid YAML`)
    }
    if (value && /^["']/.test(value) && !quotedScalar(value)) {
      errors.push(`${name}: frontmatter ${field} has an unclosed quote`)
    }
  }
  const title = rawTitle?.replace(/^['"]|['"]$/g, '').trim()
  const description = rawDescription?.replace(/^['"]|['"]$/g, '').trim()
  if (!title) errors.push(`${name}: frontmatter title is required`)
  if (!description || description.length < 24) errors.push(`${name}: frontmatter description must be useful (at least 24 characters)`)
  if (title) {
    if (titles.has(title)) errors.push(`${name}: duplicate title "${title}" also used by ${titles.get(title)}`)
    else titles.set(title, name)
  }

  const body = content.slice(frontmatter[0].length)
  const h1Count = (outsideFences(body).match(/^#\s+\S.+$/gm) || []).length
  const homeLayout = /^layout:\s*home\s*$/m.test(frontmatter[1])
  if ((!homeLayout && h1Count !== 1) || (homeLayout && h1Count > 0)) {
    errors.push(`${name}: expected ${homeLayout ? 'the home layout hero instead of a Markdown H1' : 'exactly one H1'}, found ${h1Count}`)
  }

  let inFence = false
  body.split('\n').forEach((line, index) => {
    if (!line.startsWith('```')) return
    if (!inFence && !/^```[^\s`]+/.test(line)) errors.push(`${name}:${index + 1}: fenced code block requires a language`)
    inFence = !inFence
  })
  if (inFence) errors.push(`${name}: unclosed fenced code block`)
  if (/file:\/\/\/|\/Users\/|[A-Za-z]:\\Users\\/.test(content)) errors.push(`${name}: contains a workstation-specific absolute path`)
}

if (errors.length > 0) {
  console.error(errors.join('\n'))
  process.exit(1)
}
console.log(`Checked frontmatter, headings, code fences, and portable paths in ${files.length} Markdown pages.`)
