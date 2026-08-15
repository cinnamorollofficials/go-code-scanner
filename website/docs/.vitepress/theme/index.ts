import type { EnhanceAppContext } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import './custom.css'
import ConfigBuilder from './components/ConfigBuilder.vue'
import RuleCatalogBrowser from './components/RuleCatalog.vue'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }: EnhanceAppContext) {
    app.component('ConfigBuilder', ConfigBuilder)
    app.component('RuleCatalogBrowser', RuleCatalogBrowser)
  }
}
