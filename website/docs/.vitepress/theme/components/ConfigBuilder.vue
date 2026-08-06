<template>
  <div class="config-builder">
    <h2>Interactive Configuration Builder</h2>
    <p>Select a preset or customize settings below to generate your <code>security-review.json</code> configuration.</p>

    <div class="preset-selector">
      <label for="preset-select"><strong>Load Preset:</strong></label>
      <select id="preset-select" v-model="selectedPreset" @change="applyPreset">
        <option value="minimal">Minimal App</option>
        <option value="go-service">Go Service</option>
        <option value="frontend-app">Frontend App</option>
        <option value="monorepo">Monorepo Suite</option>
        <option value="staged-hook">Staged Hook Project</option>
        <option value="offline">Offline Workspace</option>
        <option value="strict-ci">Strict CI Service</option>
      </select>
    </div>

    <div class="form-grid">
      <div class="form-group">
        <label for="project-name">Project Name</label>
        <input id="project-name" v-model="config.project" type="text" />
      </div>

      <div class="form-group">
        <label for="scan-mode">Scan Mode</label>
        <select id="scan-mode" v-model="config.mode">
          <option value="full">full</option>
          <option value="changed">changed</option>
          <option value="staged">staged</option>
        </select>
      </div>

      <div class="form-group">
        <label for="fail-on">Fail On Threshold</label>
        <select id="fail-on" v-model="config.fail_on">
          <option value="CRITICAL">CRITICAL</option>
          <option value="HIGH">HIGH</option>
          <option value="MEDIUM">MEDIUM</option>
          <option value="LOW">LOW</option>
          <option value="INFO">INFO</option>
        </select>
      </div>

      <div class="form-group">
        <label for="workers">Workers</label>
        <input id="workers" v-model.number="config.workers" type="number" min="1" max="64" />
      </div>
    </div>

    <div class="output-header">
      <h3>Generated <code>security-review.json</code></h3>
      <div class="actions">
        <button class="btn-copy" @click="copyJSON">{{ copied ? 'Copied!' : 'Copy JSON' }}</button>
      </div>
    </div>

    <pre class="json-output"><code>{{ jsonOutput }}</code></pre>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const selectedPreset = ref('minimal')
const copied = ref(false)

const config = ref({
  version: 1,
  project: 'minimal-app',
  root: '.',
  mode: 'full',
  output: 'security_findings.json',
  fail_on: 'HIGH',
  include_extensions: ['.go'],
  exclude_directories: ['.git', 'vendor'],
  exclude_files: ['security_findings.json'],
  workers: 4,
  scanners: {}
})

const presets = {
  minimal: {
    version: 1,
    project: 'minimal-app',
    root: '.',
    mode: 'full',
    output: 'security_findings.json',
    fail_on: 'HIGH',
    include_extensions: ['.go'],
    exclude_directories: ['.git', 'vendor'],
    exclude_files: ['security_findings.json'],
    workers: 4,
    scanners: {}
  },
  'go-service': {
    version: 1,
    project: 'go-service',
    root: '.',
    mode: 'full',
    output: 'security_findings.json',
    fail_on: 'HIGH',
    include_extensions: ['.go', '.yaml', '.yml', '.json'],
    exclude_directories: ['.git', 'vendor', 'bin'],
    exclude_files: ['security_findings.json'],
    workers: 8,
    scanners: {}
  },
  'frontend-app': {
    version: 1,
    project: 'frontend-app',
    root: '.',
    mode: 'full',
    output: 'security_findings.json',
    fail_on: 'HIGH',
    include_extensions: ['.ts', '.tsx', '.js', '.jsx'],
    exclude_directories: ['.git', 'node_modules', 'dist'],
    exclude_files: ['package-lock.json'],
    workers: 4,
    scanners: {}
  }
}

function applyPreset() {
  if (presets[selectedPreset.value]) {
    config.value = JSON.parse(JSON.stringify(presets[selectedPreset.value]))
  }
}

const jsonOutput = computed(() => JSON.stringify(config.value, null, 2))

async function copyJSON() {
  await navigator.clipboard.writeText(jsonOutput.value)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
</script>

<style scoped>
.config-builder {
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  padding: 24px;
  background: var(--vp-c-bg-soft);
  margin-top: 16px;
}
.preset-selector {
  margin-bottom: 20px;
}
.preset-selector select {
  margin-left: 10px;
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}
.form-group label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
}
.form-group input, .form-group select {
  width: 100%;
  padding: 8px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
}
.output-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.btn-copy {
  background: var(--vp-c-brand-1);
  color: white;
  padding: 6px 16px;
  border-radius: 4px;
  cursor: pointer;
  border: none;
  font-weight: 600;
}
.json-output {
  background: var(--vp-c-bg-alt);
  padding: 16px;
  border-radius: 6px;
  overflow-x: auto;
}
</style>
