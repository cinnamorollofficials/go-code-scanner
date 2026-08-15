import type { EnhanceAppContext } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import { h } from 'vue'
import './custom.css'
import ConfigBuilder from './components/ConfigBuilder.vue'
import DocContext from './components/DocContext.vue'
import DocFeedback from './components/DocFeedback.vue'
import RuleCatalogBrowser from './components/RuleCatalog.vue'

export default {
  extends: DefaultTheme,
  Layout: () => h(DefaultTheme.Layout, null, {
    'doc-before': () => h(DocContext),
    'doc-after': () => h(DocFeedback)
  }),
  enhanceApp({ app }: EnhanceAppContext) {
    app.component('ConfigBuilder', ConfigBuilder)
    app.component('RuleCatalogBrowser', RuleCatalogBrowser)
  }
}
