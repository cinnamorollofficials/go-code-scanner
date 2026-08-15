<template>
  <aside class="doc-feedback" aria-label="Documentation feedback">
    <strong>Was this page unclear?</strong>
    <a :href="issueURL">Report a documentation issue</a>
    <span>with the page URL and the outcome you expected.</span>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useData } from 'vitepress'

const { page } = useData()
const issueURL = computed(() => {
  const title = encodeURIComponent(`Documentation: ${page.value.title}`)
  return `https://github.com/cinnamorollofficials/go-code-scanner/issues/new?title=${title}`
})
</script>

<style scoped>
.doc-feedback {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 8px;
  margin-top: 40px;
  border-top: 1px solid var(--vp-c-divider);
  padding-top: 20px;
  color: var(--vp-c-text-2);
  font-size: 0.9rem;
}

.doc-feedback strong { color: var(--vp-c-text-1); }
.doc-feedback a { color: var(--vp-c-brand-1); font-weight: 600; }
.doc-feedback a:hover { text-decoration: underline; }
</style>
