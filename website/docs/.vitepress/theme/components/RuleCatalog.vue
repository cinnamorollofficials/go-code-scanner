<template>
  <div class="rule-catalog">
    <ul class="domain-summary" aria-label="Rule counts by domain">
      <li v-for="value in domains" :key="value">
        <strong>{{ domainCounts[value] }}</strong>
        <span>{{ label(value) }}</span>
      </li>
    </ul>
    <div class="filters" role="search" aria-label="Filter rule catalog">
      <label class="search-field">
        <span>Rule ID or description</span>
        <input v-model.trim="query" type="search" placeholder="Search rules…" />
      </label>
      <label>
        <span>Domain</span>
        <select v-model="domain">
          <option value="">All domains</option>
          <option v-for="value in domains" :key="value" :value="value">{{ label(value) }}</option>
        </select>
      </label>
      <label>
        <span>Severity</span>
        <select v-model="severity">
          <option value="">All severities</option>
          <option v-for="value in severities" :key="value" :value="value">{{ value }}</option>
        </select>
      </label>
      <label>
        <span>Language / ecosystem</span>
        <select v-model="language">
          <option value="">All languages</option>
          <option v-for="value in languages" :key="value" :value="value">{{ value }}</option>
        </select>
      </label>
      <label>
        <span>Category</span>
        <select v-model="category">
          <option value="">All categories</option>
          <option v-for="value in categories" :key="value" :value="value">{{ label(value) }}</option>
        </select>
      </label>
      <button type="button" :disabled="!hasFilters" @click="clearFilters">Clear filters</button>
    </div>

    <p class="result-count" role="status" aria-live="polite">
      Showing {{ filteredRules.length }} of {{ rules.length }} rules.
    </p>

    <div class="table-wrap" tabindex="0" aria-label="Filtered rule results">
      <table>
        <thead>
          <tr><th>Rule ID</th><th>Domain</th><th>Severity</th><th>Category</th><th>Description</th></tr>
        </thead>
        <tbody>
          <tr v-for="rule in filteredRules" :key="rule.id">
            <td><a :href="withBase(rule.href)"><code>{{ rule.id }}</code></a></td>
            <td>{{ label(rule.domain) }}</td>
            <td><span class="severity" :data-severity="rule.severity">{{ rule.severity }}</span></td>
            <td>{{ label(rule.category) }}</td>
            <td>{{ rule.description }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <p v-if="filteredRules.length === 0" class="empty-state">No rules match the selected filters.</p>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { withBase } from 'vitepress'
import { rules } from './RuleCatalogData.js'

const query = ref('')
const domain = ref('')
const severity = ref('')
const language = ref('')
const category = ref('')

const domains = [...new Set(rules.map((rule) => rule.domain))].sort()
const domainCounts = Object.fromEntries(domains.map((value) => [value, rules.filter((rule) => rule.domain === value).length]))
const severities = ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW'].filter((value) => rules.some((rule) => rule.severity === value))
const languages = [...new Set(rules.flatMap((rule) => rule.languages))].sort()
const categories = [...new Set(rules.map((rule) => rule.category))].sort()
const hasFilters = computed(() => Boolean(query.value || domain.value || severity.value || language.value || category.value))

const filteredRules = computed(() => {
  const needle = query.value.toLowerCase()
  return rules.filter((rule) => {
    const text = `${rule.id} ${rule.description} ${rule.category}`.toLowerCase()
    return (!needle || text.includes(needle)) &&
      (!domain.value || rule.domain === domain.value) &&
      (!severity.value || rule.severity === severity.value) &&
      (!language.value || rule.languages.includes(language.value)) &&
      (!category.value || rule.category === category.value)
  })
})

function clearFilters() {
  query.value = ''
  domain.value = ''
  severity.value = ''
  language.value = ''
  category.value = ''
}

function label(value) {
  return value.replaceAll('_', ' ').replace(/\b\w/g, (letter) => letter.toUpperCase())
}
</script>

<style scoped>
.domain-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
  padding: 0;
  list-style: none;
}

.domain-summary li {
  display: grid;
  padding: 10px 12px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  background: var(--vp-c-bg-soft);
}

.domain-summary strong { font-size: 1.2rem; color: var(--vp-c-brand-1); }
.domain-summary span { font-size: 0.8rem; color: var(--vp-c-text-2); }

.filters {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 10px;
  background: var(--vp-c-bg-soft);
}

.filters label {
  display: grid;
  gap: 5px;
  color: var(--vp-c-text-2);
  font-size: 0.8rem;
  font-weight: 600;
}

.filters input,
.filters select,
.filters button {
  min-height: 40px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  padding: 7px 9px;
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font: inherit;
}

.filters button {
  align-self: end;
  cursor: pointer;
}

.filters button:disabled { cursor: not-allowed; opacity: 0.55; }
.search-field { grid-column: span 2; }
.result-count { color: var(--vp-c-text-2); font-size: 0.9rem; }
.table-wrap { overflow-x: auto; }
.table-wrap:focus { outline: 2px solid var(--vp-c-brand-1); outline-offset: 2px; }
table { min-width: 780px; }
th { white-space: nowrap; }
.severity { font-size: 0.75rem; font-weight: 700; }
.severity[data-severity="CRITICAL"] { color: var(--vp-c-danger-1); }
.severity[data-severity="HIGH"] { color: #c2410c; }
.severity[data-severity="MEDIUM"] { color: #a16207; }
.empty-state { padding: 20px; text-align: center; color: var(--vp-c-text-2); }

@media (max-width: 640px) {
  .search-field { grid-column: span 1; }
}
</style>
