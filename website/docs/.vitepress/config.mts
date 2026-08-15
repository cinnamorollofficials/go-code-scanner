import { defineConfig } from 'vitepress'

function sidebarDocs() {
  return [
    {
      text: 'Get Started',
      items: [
        { text: 'Overview', link: '/getting-started/' },
        { text: 'Installation', link: '/getting-started/installation' },
        { text: 'First Scan & Exit Codes', link: '/getting-started/first-scan' },
        { text: 'Five-Minute CI Setup', link: '/getting-started/ci-setup' }
      ]
    },
    {
      text: 'Guides',
      items: [
        { text: 'Guides Overview', link: '/guides/' },
        { text: 'Pre-Commit Hooks', link: '/guides/pre-commit-hooks' },
        { text: 'GitHub Actions / GitLab CI', link: '/guides/ci-integrations' },
        { text: 'Gradual Adoption with Baselines', link: '/guides/baselines' },
        { text: 'Managing Suppressions', link: '/guides/suppressions' },
        { text: 'Reproducing a Finding', link: '/guides/reproducing-findings' },
        { text: 'Troubleshooting', link: '/guides/troubleshooting' }
      ]
    },
    {
      text: 'Concepts',
      items: [
        { text: 'Concepts Overview', link: '/concepts/' },
        { text: 'Scan Modes and Isolation', link: '/concepts/scan-modes' },
        { text: 'Profiles and Policy', link: '/concepts/profiles-and-policy' },
        { text: 'Reports and Finding Lifecycle', link: '/concepts/reports-and-finding-lifecycle' },
        { text: 'Frontend Scanning', link: '/concepts/frontend-scanning' },
        { text: 'SQL Taint Analysis', link: '/concepts/sql-taint-analysis' }
      ]
    },
    {
      text: 'Reference',
      items: [
        { text: 'CLI Reference', link: '/reference/cli' },
        { text: 'Configuration Reference', link: '/reference/configuration' },
        { text: 'Scanner Compatibility', link: '/reference/scanners' },
        { text: 'Rule Catalog', link: '/reference/rules' },
        { text: 'Config Builder', link: '/reference/config-builder' }
      ]
    },
    {
      text: 'Project',
      items: [
        { text: 'Security Model', link: '/security' },
        { text: 'Changelog', link: '/changelog' },
        { text: 'Contributing Guide', link: '/contributing' },
        { text: 'Documentation Author Guide', link: '/author-guide' }
      ]
    }
  ]
}

const base = process.env.VITEPRESS_BASE || '/go-code-scanner/'
const basePrefix = base === '/' ? '' : base.replace(/\/$/, '')

export default defineConfig({
  title: 'Go Code Scanner',
  description: 'Policy-driven, offline-first security analysis CLI',
  base,
  cleanUrls: true,
  lastUpdated: true,
  sitemap: {
    hostname: 'https://cinnamorollofficials.github.io/go-code-scanner/'
  },
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: `${basePrefix}/favicon.svg` }],
    ['meta', { name: 'theme-color', content: '#10b981' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'Go Code Scanner' }],
    ['meta', { property: 'og:description', content: 'Policy-driven, offline-first security analysis CLI' }]
  ],
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { 
        text: 'Documentation', 
        link: '/getting-started/', 
        activeMatch: '^/(getting-started|guides|concepts|features|author-guide|changelog|contributing|security)/' 
      },
      { 
        text: 'Reference', 
        link: '/reference/', 
        activeMatch: '^/reference/' 
      },
      { 
        text: 'Unreleased', 
        link: 'https://github.com/cinnamorollofficials/go-code-scanner/releases' 
      }
    ],
    sidebar: {
      '/getting-started/': sidebarDocs(),
      '/guides/': sidebarDocs(),
      '/concepts/': sidebarDocs(),
      '/features/': sidebarDocs(),
      '/author-guide': sidebarDocs(),
      '/changelog': sidebarDocs(),
      '/contributing': sidebarDocs(),
      '/security': sidebarDocs(),

      '/reference/': [
        {
          text: 'Tools & CLI Reference',
          items: [
            { text: 'Reference Overview', link: '/reference/' },
            { text: 'CLI Reference', link: '/reference/cli' },
            { text: 'Interactive Config Builder', link: '/reference/config-builder' },
            { text: 'Config Builder Contract', link: '/reference/config-builder-contract' },
            { text: 'Rule Catalog', link: '/reference/rules' },
            { text: 'Scanner Compatibility', link: '/reference/scanners' }
          ]
        },
        {
          text: 'Configuration Specs',
          items: [
            { text: 'Configuration Overview', link: '/reference/configuration' },
            { text: 'Generated Field Reference', link: '/reference/config/generated-reference' },
            { text: 'Input & Paths', link: '/reference/config/input-and-paths' },
            { text: 'Profiles & Policy', link: '/reference/config/profiles-and-policy' },
            { text: 'Scanner Definitions', link: '/reference/config/scanners' },
            { text: 'Git Hooks', link: '/reference/config/hooks' },
            { text: 'Frontend Policy', link: '/reference/config/frontend' },
            { text: 'Supply Chain Policy', link: '/reference/config/supply-chain' },
            { text: 'Governance Policy', link: '/reference/config/governance' },
            { text: 'Architecture Policy', link: '/reference/config/architecture' },
            { text: 'Cache Policy', link: '/reference/config/cache' }
          ]
        }
      ]
    },
    editLink: {
      pattern: 'https://github.com/cinnamorollofficials/go-code-scanner/edit/main/website/docs/:path',
      text: 'Edit this page on GitHub'
    },
    docFooter: {
      prev: 'Previous page',
      next: 'Next page'
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/cinnamorollofficials/go-code-scanner' }
    ],
    search: {
      provider: 'local'
    },
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2026 Security Review Team'
    }
  }
})
