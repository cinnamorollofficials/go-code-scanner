import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Go Code Scanner',
  description: 'Policy-driven, offline-first security analysis CLI',
  base: '/go-code-scanner/',
  cleanUrls: true,
  lastUpdated: true,
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/go-code-scanner/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#10b981' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'Go Code Scanner' }],
    ['meta', { property: 'og:description', content: 'Policy-driven, offline-first security analysis CLI' }]
  ],
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Getting Started', link: '/getting-started/' },
      { text: 'Features', link: '/features/' },
      { text: 'Reference', link: '/reference/' },
      { text: 'Guides', link: '/guides/' },
      { text: 'Author Guide', link: '/author-guide' }
    ],
    sidebar: {
      '/getting-started/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Overview', link: '/getting-started/' }
          ]
        }
      ],
      '/features/': [
        {
          text: 'Core Features',
          items: [
            { text: 'Features Overview', link: '/features/' }
          ]
        }
      ],
      '/reference/': [
        {
          text: 'Product References',
          items: [
            { text: 'Reference Overview', link: '/reference/' },
            { text: 'CLI Reference', link: '/reference/cli' },
            { text: 'Configuration Reference', link: '/reference/configuration' }
          ]
        }
      ],
      '/guides/': [
        {
          text: 'Guides & Integration',
          items: [
            { text: 'Guides Overview', link: '/guides/' }
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
