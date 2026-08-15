<template>
  <nav v-if="items.length > 1" class="doc-context" aria-label="Breadcrumb">
    <ol>
      <li v-for="(item, index) in items" :key="item.label">
        <a v-if="item.href" :href="withBase(item.href)">{{ item.label }}</a>
        <span v-else aria-current="page">{{ item.label }}</span>
        <span v-if="index < items.length - 1" class="separator" aria-hidden="true">/</span>
      </li>
    </ol>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useData, withBase } from 'vitepress'

const { page } = useData()

const items = computed(() => {
  const path = page.value.relativePath
  const current = { label: page.value.title, href: '' }
  if (path.startsWith('reference/config/')) {
    return [
      { label: 'Reference', href: '/reference/' },
      { label: 'Configuration', href: '/reference/configuration' },
      current
    ]
  }
  if (path.startsWith('reference/rules/')) {
    return [
      { label: 'Reference', href: '/reference/' },
      { label: 'Rule Catalog', href: '/reference/rule-catalog' },
      current
    ]
  }
  return []
})
</script>

<style scoped>
.doc-context {
  margin-bottom: 18px;
  color: var(--vp-c-text-2);
  font-size: 0.8rem;
  font-weight: 600;
}

ol,
li {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

ol { flex-wrap: wrap; }
a { color: var(--vp-c-brand-1); }
a:hover { text-decoration: underline; }
.separator { color: var(--vp-c-divider); }
</style>
