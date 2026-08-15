import { existsSync, readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const websiteDir = resolve(fileURLToPath(new URL('..', import.meta.url)))
const read = (name) => readFileSync(join(websiteDir, name), 'utf8')
const deployment = read('DEPLOYMENT.md')
const dockerfile = read('Dockerfile')
const compose = read('docker-compose.yml')
const nginx = read('nginx.conf')
const scripts = JSON.parse(read('package.json')).scripts
const errors = []

if (/file:\/\/\/|\/Users\/|[A-Za-z]:\\Users\\/.test(deployment)) {
  errors.push('DEPLOYMENT.md contains a workstation-specific path')
}

for (const match of deployment.matchAll(/\[[^\]]+\]\((\.\/[^)#]+)(?:#[^)]+)?\)/g)) {
  if (!existsSync(join(websiteDir, match[1]))) errors.push(`DEPLOYMENT.md links to missing file ${match[1]}`)
}

if (!scripts['docs:build-site']) errors.push('package.json is missing docs:build-site for the website-only container build')
if (!dockerfile.includes('RUN npm run docs:build-site')) errors.push('Dockerfile must use the website-only docs:build-site command')
if (!dockerfile.includes('ARG VITEPRESS_BASE=/')) errors.push('Dockerfile must default VITEPRESS_BASE to the root path')
if (!/VITEPRESS_BASE:\s*["']\/["']/.test(compose)) errors.push('docker-compose.yml must explicitly build for the root path')
if (!/ports:\s*\n\s*-\s*["']80:80["']/.test(compose)) errors.push('docker-compose.yml must publish the documented port 80')
if (!nginx.includes('try_files $uri $uri.html $uri/index.html $uri/ /404.html =404;')) {
  errors.push('nginx.conf must preserve clean URL and nested index routing')
}
if (!deployment.includes('VITEPRESS_BASE=/docs/ npm run docs:build-site')) {
  errors.push('DEPLOYMENT.md must show the verified custom-base static build command')
}
if (!deployment.includes('tidak langsung melayani subpath')) {
  errors.push('DEPLOYMENT.md must state the shipped Nginx root-path limitation')
}

if (errors.length > 0) {
  console.error(errors.join('\n'))
  process.exit(1)
}

console.log('Checked portable deployment links, Docker build contract, Compose root path, Nginx routing, and base-path guidance.')
