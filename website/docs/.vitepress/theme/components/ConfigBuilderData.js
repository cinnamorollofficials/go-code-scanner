export const presetDescriptions = {
  minimal: 'Lightweight starter configuration suitable for standalone Go utilities.',
  'go-service': 'Standard configuration for production Go microservices and APIs.',
  'frontend-app': 'Dedicated scanner configuration for React, Vue, Svelte, and Next.js applications.',
  monorepo: 'Full-featured configuration for large multi-package repositories.',
  'staged-hook': 'Optimized for sub-second pre-commit git hook validation.',
  offline: 'Air-gapped profile containing only the built-in pattern scanner.',
  'strict-ci': 'CI policy that enforces Medium-or-higher findings in core domains.',
  'external-scanner': 'Gitleaks adapter plus the built-in pattern scanner.',
  'gradual-adoption': 'Baseline-backed pre-commit setup that evaluates new findings.'
}

function base(project, overrides = {}) {
  return {
    version: 1,
    project,
    root: '.',
    mode: 'full',
    output: 'security_findings.json',
    fail_on: 'HIGH',
    include_extensions: ['.go'],
    exclude_directories: ['.git', 'vendor'],
    exclude_files: ['security_findings.json'],
    workers: 4,
    scanners: {},
    ...overrides
  }
}

export const presets = {
  minimal: base('minimal-app'),
  'go-service': base('go-service', {
    include_extensions: ['.go', '.yaml', '.yml', '.json'],
    exclude_directories: ['.git', 'vendor', 'bin', 'testdata'],
    workers: 8,
    policy: {
      security: 'HIGH',
      reliability: 'HIGH',
      hardening: 'HIGH',
      quality: 'MEDIUM',
      supply_chain: 'HIGH',
      governance: 'MEDIUM'
    },
    cache: {
      enabled: true,
      directory: '.go-code-scanner-cache',
      max_age: '168h',
      max_bytes: 104857600
    }
  }),
  'frontend-app': base('frontend-app', {
    include_extensions: ['.ts', '.tsx', '.js', '.jsx', '.vue', '.svelte'],
    exclude_directories: ['.git', 'node_modules', 'dist', '.next', '.nuxt'],
    exclude_files: ['package-lock.json', 'yarn.lock'],
    frontend: {
      enabled: true,
      frameworks: ['react', 'vue', 'next'],
      client_roots: ['src/client', 'pages', 'components'],
      server_roots: ['src/server', 'api']
    }
  }),
  monorepo: base('enterprise-monorepo', {
    mode: 'changed',
    include_extensions: ['.go', '.ts', '.tsx', '.js', '.json', '.yaml'],
    exclude_directories: ['.git', 'node_modules', 'vendor', 'dist', 'build'],
    workers: 16,
    policy: {
      security: 'HIGH',
      reliability: 'HIGH',
      hardening: 'HIGH',
      quality: 'HIGH',
      supply_chain: 'CRITICAL',
      governance: 'HIGH'
    },
    cache: {
      enabled: true,
      directory: '.go-code-scanner-cache',
      max_age: '72h',
      max_bytes: 524288000
    }
  }),
  'staged-hook': base('staged-hook-app', {
    mode: 'staged',
    hooks: {
      pre_commit: {
        enabled: true,
        profile: 'fast',
        staged_only: true,
        new_only: true
      },
      commit_msg: {
        enabled: true,
        message_pattern: '^(feat|fix|docs|style|refactor|test|chore)(\\(.+\\))?: .+',
        max_subject_length: 72
      },
      pre_push: { enabled: false }
    }
  }),
  offline: base('offline-workspace', {
    profiles: { offline: ['pattern'] },
    offline_profiles: ['offline']
  }),
  'strict-ci': base('strict-ci-service', {
    fail_on: 'MEDIUM',
    include_extensions: ['.go', '.yaml', '.json'],
    workers: 8,
    policy: {
      security: 'MEDIUM',
      reliability: 'MEDIUM',
      hardening: 'MEDIUM',
      quality: 'MEDIUM',
      supply_chain: 'HIGH',
      governance: 'HIGH'
    }
  }),
  'external-scanner': base('external-scanner-suite', {
    profiles: { external: ['pattern', 'gitleaks'] },
    scanners: {
      gitleaks: {
        enabled: true,
        required: false,
        type: 'adapter',
        adapter: 'gitleaks',
        on_missing: 'skip',
        timeout: '2m'
      }
    }
  }),
  'gradual-adoption': base('legacy-app-adoption', {
    baseline_file: '.security-baseline.json',
    hooks: {
      pre_commit: {
        enabled: true,
        profile: 'fast',
        staged_only: true,
        new_only: true
      },
      commit_msg: { enabled: false },
      pre_push: { enabled: false }
    }
  })
}

const modes = new Set(['full', 'changed', 'staged'])
const severities = new Set(['CRITICAL', 'HIGH', 'MEDIUM', 'LOW'])
const domains = new Set(['quality', 'reliability', 'hardening', 'security', 'supply_chain', 'governance'])
const scannerTypes = new Set(['pattern', 'command', 'adapter'])
const adapters = new Set(['gofmt', 'go-vet', 'go-test', 'govulncheck', 'gosec', 'gitleaks', 'trivy', 'osv-scanner', 'semgrep', 'eslint', 'tsc', 'biome'])
const frameworks = new Set(['vanilla', 'react', 'next', 'nextjs', 'next.js', 'vue', 'nuxt', 'svelte', 'sveltekit'])

function hasTraversal(value) {
  return typeof value === 'string' && /(^|[\\/])\.\.([\\/]|$)/.test(value)
}

function nonEmpty(value) {
  return typeof value === 'string' && value.trim().length > 0
}

function validateStringArray(errors, value, path, { extensions = false, paths = false } = {}) {
  if (!Array.isArray(value)) {
    errors.push(`${path} must be an array.`)
    return
  }
  const seen = new Set()
  value.forEach((item, index) => {
    if (!nonEmpty(item)) errors.push(`${path}[${index}] cannot be empty.`)
    if (extensions && nonEmpty(item) && !item.startsWith('.')) errors.push(`${path}[${index}] must start with ".".`)
    if (paths && hasTraversal(item)) errors.push(`${path}[${index}] cannot contain a parent-directory segment.`)
    const key = String(item).toLowerCase()
    if (seen.has(key)) errors.push(`${path} cannot contain duplicate values.`)
    seen.add(key)
  })
}

function validateScanners(errors, scanners) {
  if (!scanners || typeof scanners !== 'object' || Array.isArray(scanners)) {
    errors.push('scanners must be an object.')
    return
  }
  Object.entries(scanners).forEach(([id, scanner]) => {
    const path = `scanners.${id}`
    if (!nonEmpty(id)) errors.push('Scanner IDs cannot be empty.')
    if (!scanner || typeof scanner !== 'object') {
      errors.push(`${path} must be an object.`)
      return
    }
    if (!scannerTypes.has(scanner.type)) errors.push(`${path}.type must be pattern, command, or adapter.`)
    if (scanner.type === 'pattern' && id !== 'pattern') errors.push(`${path}: pattern type is reserved for the pattern scanner.`)
    if (scanner.type === 'adapter' && !adapters.has(scanner.adapter)) errors.push(`${path}.adapter is not supported.`)
    if (scanner.type === 'command') {
      if (!Array.isArray(scanner.command) || !scanner.command.every(nonEmpty)) errors.push(`${path}.command must contain an executable and optional arguments.`)
      if (!domains.has(scanner.domain)) errors.push(`${path}.domain must be a canonical domain.`)
      if (!severities.has(scanner.severity)) errors.push(`${path}.severity must be a supported severity.`)
      if (!nonEmpty(scanner.category)) errors.push(`${path}.category is required.`)
      if (!nonEmpty(scanner.description)) errors.push(`${path}.description is required.`)
    }
    if (scanner.workspace && !['root', 'staged'].includes(scanner.workspace)) errors.push(`${path}.workspace must be root or staged.`)
    if (scanner.on_missing && !['fail', 'skip'].includes(scanner.on_missing)) errors.push(`${path}.on_missing must be fail or skip.`)
  })
}

export function validateConfig(config) {
  const errors = []
  if (!config || typeof config !== 'object') return ['Configuration must be an object.']
  if (config.version !== 1) errors.push('Schema version must be 1.')
  if (!nonEmpty(config.project)) errors.push('Project name cannot be empty.')
  if (!nonEmpty(config.root)) errors.push('Root directory cannot be empty.')
  else if (hasTraversal(config.root)) errors.push('Root directory cannot contain a parent-directory segment.')
  if (!modes.has(config.mode)) errors.push('Scan mode must be full, changed, or staged.')
  if (!nonEmpty(config.output)) errors.push('Report output path cannot be empty.')
  else if (hasTraversal(config.output)) errors.push('Report output path cannot contain a parent-directory segment.')
  if (!severities.has(config.fail_on)) errors.push('Fail On threshold must be CRITICAL, HIGH, MEDIUM, or LOW.')
  if (!Number.isInteger(config.workers) || config.workers < 1 || config.workers > 64) errors.push('Workers must be an integer between 1 and 64.')

  validateStringArray(errors, config.include_extensions, 'include_extensions', { extensions: true })
  validateStringArray(errors, config.exclude_directories, 'exclude_directories', { paths: true })
  validateStringArray(errors, config.exclude_files, 'exclude_files', { paths: true })
  validateScanners(errors, config.scanners)

  if (config.policy) {
    Object.entries(config.policy).forEach(([domain, severity]) => {
      if (!domains.has(domain)) errors.push(`policy.${domain} is not a canonical domain.`)
      if (!severities.has(severity)) errors.push(`policy.${domain} must be a supported severity.`)
    })
  }
  if (config.cache?.enabled) {
    if (!nonEmpty(config.cache.directory) || hasTraversal(config.cache.directory)) errors.push('cache.directory must be a safe non-empty path.')
    if (!/^([0-9]+(ns|us|µs|ms|s|m|h))+$/.test(config.cache.max_age || '')) errors.push('cache.max_age must be a positive Go duration.')
    if (!Number.isInteger(config.cache.max_bytes) || config.cache.max_bytes < 1) errors.push('cache.max_bytes must be a positive integer.')
  }
  if (config.frontend?.enabled) {
    validateStringArray(errors, config.frontend.frameworks, 'frontend.frameworks')
    config.frontend.frameworks?.forEach((framework, index) => {
      if (!frameworks.has(String(framework).toLowerCase())) errors.push(`frontend.frameworks[${index}] is not supported.`)
    })
    validateStringArray(errors, config.frontend.client_roots, 'frontend.client_roots', { paths: true })
    validateStringArray(errors, config.frontend.server_roots, 'frontend.server_roots', { paths: true })
  }
  if (config.profiles) {
    Object.entries(config.profiles).forEach(([name, scannerIDs]) => validateStringArray(errors, scannerIDs, `profiles.${name}`))
  }
  if (config.offline_profiles) {
    validateStringArray(errors, config.offline_profiles, 'offline_profiles')
    config.offline_profiles.forEach((name) => {
      if (!config.profiles?.[name]) errors.push(`offline profile "${name}" is not defined in profiles.`)
    })
  }
  return errors
}
