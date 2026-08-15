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

function value(frontmatter, key) {
  return frontmatter.match(new RegExp(`^${key}:\\s*(.+)$`, 'm'))?.[1]
    ?.replace(/^["']|["']$/g, '')
    .trim()
}

function plainHeading(heading) {
  return heading.replaceAll('`', '').replace(/\s+\{#[^}]+\}$/, '').trim()
}

const errors = []
const files = markdownFiles(docsDir)
const bannedPhrases = [
  ['seamlessly', 'replace marketing language with the concrete outcome'],
  ['100% locally', 'state the browser data boundary precisely'],
  ['never impeded', 'replace an absolute performance claim with a measured budget'],
  ['fast track guide', 'describe the audience and outcome instead'],
  ['deep dive', 'describe the actual scope instead'],
  ["do's and don'ts", 'use "unsafe and safer examples"']
]

for (const file of files) {
  const name = relative(websiteDir, file)
  const content = readFileSync(file, 'utf8')
  const frontmatterMatch = content.match(/^---\n([\s\S]*?)\n---\n/)
  if (!frontmatterMatch) continue
  const frontmatter = frontmatterMatch[1]
  const title = value(frontmatter, 'title') || ''
  const description = value(frontmatter, 'description') || ''
  const body = content.slice(frontmatterMatch[0].length)
  const prose = outsideFences(body)
  const homeLayout = /^layout:\s*home\s*$/m.test(frontmatter)

  if (!/^For [^:]+:\s+\S/.test(description)) {
    errors.push(`${name}: description must use "For <audience>: <outcome and scope>"`)
  }
  if (description.length < 50 || description.length > 160) {
    errors.push(`${name}: description must be 50-160 characters, found ${description.length}`)
  }
  if (title.includes('&')) errors.push(`${name}: use "and" instead of "&" in the title`)

  if (!homeLayout) {
    const h1 = prose.match(/^#\s+(.+)$/m)?.[1] || ''
    if (plainHeading(h1) !== title) {
      errors.push(`${name}: frontmatter title "${title}" must match H1 "${plainHeading(h1)}"`)
    }
  }

  for (const heading of prose.matchAll(/^#{1,6}\s+(.+)$/gm)) {
    if (heading[1].includes('&')) errors.push(`${name}: use "and" instead of "&" in heading "${heading[1]}"`)
  }

  const lower = prose.toLowerCase()
  for (const [phrase, guidance] of bannedPhrases) {
    if (lower.includes(phrase)) errors.push(`${name}: ${guidance} (found "${phrase}")`)
  }
  if (/^(?:example|expected|sample) output:/mi.test(prose)) {
    errors.push(`${name}: classify output as captured, illustrative, abbreviated, or complete`)
  }
  if (/\b(?:Go code scanner|Go-Code-Scanner)\b/.test(prose)) {
    errors.push(`${name}: use the product name "Go Code Scanner"`)
  }
  if (name.startsWith('docs/reference/rules/') && /##### Unsafe and Safer Examples?/.test(content) && !content.includes('examples below are illustrative')) {
    errors.push(`${name}: generated rule examples must be labelled illustrative`)
  }
}

if (errors.length > 0) {
  console.error(errors.join('\n'))
  process.exit(1)
}

console.log(`Checked audience-focused summaries, title consistency, terminology, headings, output labels, and generated example labels in ${files.length} pages.`)
