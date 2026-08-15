import { defineConfig } from 'vitepress'

function sidebarDocs(activeSection: string) {
  const sections = [
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
        { text: 'Rule Catalog', link: '/reference/rule-catalog' },
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
  return [
    sections.find((section) => section.text === activeSection)!,
    {
      text: 'Explore Other Sections',
      items: sections
        .filter((section) => section.text !== activeSection)
        .map((section) => ({ text: section.text, link: section.items[0].link }))
    }
  ]
}

const base = process.env.VITEPRESS_BASE || '/go-code-scanner/'
const basePrefix = base === '/' ? '' : base.replace(/\/$/, '')
const siteOrigin = 'https://cinnamorollofficials.github.io'
const siteBase = new URL(base, siteOrigin)
const socialImage = new URL(`${basePrefix}/social-card.png`, siteOrigin).href

function canonicalURL(relativePath: string) {
  let route = relativePath.replace(/\.md$/, '')
  if (route === 'index') route = ''
  else if (route.endsWith('/index')) route = route.slice(0, -'/index'.length)
  return new URL(route, siteBase).href
}

export default defineConfig({
  title: 'Go Code Scanner',
  description: 'Policy-driven, offline-first security analysis CLI',
  base,
  cleanUrls: true,
  lastUpdated: true,
  sitemap: {
    hostname: 'https://cinnamorollofficials.github.io/go-code-scanner/',
    transformItems: (items) => items.filter((item) => {
      const path = item.url.replace(/\/$/, '')
      return path !== 'reference/rules' && !path.endsWith('/reference/rules')
    })
  },
  transformPageData(pageData) {
    const legacyRules = pageData.relativePath === 'reference/rules.md'
    const canonical = legacyRules
      ? new URL('reference/rule-catalog', siteBase).href
      : canonicalURL(pageData.relativePath)
    const socialTitle = pageData.relativePath === 'index.md'
      ? 'Go Code Scanner'
      : `${pageData.title} | Go Code Scanner`
    const socialDescription = pageData.frontmatter.description || 'Policy-driven, offline-first security analysis CLI'

    pageData.frontmatter.head = [
      ...(pageData.frontmatter.head || []),
      ['link', { rel: 'canonical', href: canonical }],
      ['meta', { property: 'og:url', content: canonical }],
      ['meta', { property: 'og:title', content: socialTitle }],
      ['meta', { property: 'og:description', content: socialDescription }],
      ['meta', { property: 'og:image', content: socialImage }],
      ['meta', { property: 'og:image:width', content: '1200' }],
      ['meta', { property: 'og:image:height', content: '630' }],
      ['meta', { property: 'og:image:alt', content: 'Go Code Scanner documentation' }],
      ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
      ['meta', { name: 'twitter:title', content: socialTitle }],
      ['meta', { name: 'twitter:description', content: socialDescription }],
      ['meta', { name: 'twitter:image', content: socialImage }]
    ]

    if (legacyRules) {
      pageData.frontmatter.search = false
      pageData.frontmatter.head = [
        ...(pageData.frontmatter.head || []),
        ['meta', { name: 'robots', content: 'noindex,follow' }]
      ]
    }
    return pageData
  },
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: `${basePrefix}/favicon.svg` }],
    ['meta', { name: 'theme-color', content: '#10b981' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'Go Code Scanner Documentation' }]
  ],
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      {
        text: 'Learn',
        activeMatch: '^/(getting-started|guides|concepts)/',
        items: [
          { text: 'Get Started', link: '/getting-started/' },
          { text: 'Guides', link: '/guides/' },
          { text: 'Concepts', link: '/concepts/' }
        ]
      },
      {
        text: 'Reference',
        link: '/reference/',
        activeMatch: '^/reference/'
      },
      {
        text: 'Project',
        activeMatch: '^/(security|changelog|contributing|author-guide)',
        items: [
          { text: 'Security Model', link: '/security' },
          { text: 'Changelog', link: '/changelog' },
          { text: 'Contributing', link: '/contributing' },
          { text: 'Documentation Author Guide', link: '/author-guide' }
        ]
      },
      {
        text: 'Development docs',
        link: 'https://github.com/cinnamorollofficials/go-code-scanner/releases'
      }
    ],
    sidebar: {
      '/getting-started/': sidebarDocs('Get Started'),
      '/guides/': sidebarDocs('Guides'),
      '/concepts/': sidebarDocs('Concepts'),
      '/author-guide': sidebarDocs('Project'),
      '/changelog': sidebarDocs('Project'),
      '/contributing': sidebarDocs('Project'),
      '/security': sidebarDocs('Project'),

      '/reference/': [
        {
          text: 'Tools & CLI Reference',
          items: [
            { text: 'Reference Overview', link: '/reference/' },
            { text: 'CLI Reference', link: '/reference/cli' },
            { text: 'Interactive Config Builder', link: '/reference/config-builder' },
            { text: 'Config Builder Contract', link: '/reference/config-builder-contract' },
            { text: 'Rule Catalog', link: '/reference/rule-catalog' },
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
