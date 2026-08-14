import { defineConfig } from 'vitepress'

function sidebarDocs() {
  return [
    {
      text: 'Getting Started',
      items: [
        { text: 'Overview', link: '/getting-started/' },
        { text: 'Installation', link: '/getting-started/installation' },
        { text: 'First Scan & Exit Codes', link: '/getting-started/first-scan' }
      ]
    },
    {
      text: 'Core Features',
      items: [
        { text: 'Features Overview', link: '/features/' },
        { text: 'Scan Execution & Policy', link: '/features/scan-execution-and-policy' },
        { text: 'Reports & Finding Lifecycle', link: '/features/reports-and-finding-lifecycle' },
        { text: 'How It Works & Reproduce Findings', link: '/features/analysis-and-reproduction' },
        { text: 'AST & SQL Taint Analysis', link: '/features/sql-taint-analysis' },
        { text: 'Developer Workflow Features', link: '/features/developer-workflow-features' },
        { text: 'Frontend & Client Scanning', link: '/features/client-scanning' }
      ]
    },
    {
      text: 'Guides & Integrations',
      items: [
        { text: 'Guides Overview', link: '/guides/' },
        { text: 'Local & CI Integrations', link: '/guides/ci-integrations' },
        { text: 'Adoption & Troubleshooting', link: '/guides/troubleshooting' }
      ]
    },
    {
      text: 'Development',
      items: [
        { text: 'Security Model', link: '/security' },
        { text: 'Contributing Guide', link: '/contributing' },
        { text: 'Author Guide', link: '/author-guide' },
        { text: 'Changelog', link: '/changelog' }
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
        activeMatch: '^/(getting-started|features|guides|author-guide|changelog|contributing|security)/' 
      },
      { 
        text: 'Reference', 
        link: '/reference/', 
        activeMatch: '^/reference/' 
      },
      { 
        text: 'v1.0.0', 
        link: 'https://github.com/cinnamorollofficials/go-code-scanner/releases' 
      }
    ],
    sidebar: {
      '/getting-started/': sidebarDocs(),
      '/features/': sidebarDocs(),
      '/guides/': sidebarDocs(),
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
