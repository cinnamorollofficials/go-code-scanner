<template>
  <div class="config-builder" role="region" aria-labelledby="builder-title">
    <div class="builder-header">
      <h2 id="builder-title">Interactive Configuration Builder</h2>
      <p>Generate, customize, and export validated <code>security-review.json</code> configuration files for your repository.</p>
      <p class="validation-scope">Browser checks catch common schema and safety errors. The CLI validator remains authoritative for downloaded files.</p>
    </div>

    <!-- Live Announcement for Screen Readers -->
    <div class="sr-only" role="status" aria-live="polite" aria-atomic="true">
      {{ statusAnnouncement }}
    </div>

    <!-- Preset Selection -->
    <div class="control-section">
      <div class="preset-row">
        <label for="preset-select" class="field-label"><strong>Configuration Preset:</strong></label>
        <div class="preset-controls">
          <select id="preset-select" v-model="selectedPreset" @change="onPresetChange" class="input-control select-control" aria-describedby="preset-desc">
            <option value="minimal">Minimal Starter</option>
            <option value="go-service">Go Backend Service</option>
            <option value="frontend-app">Frontend Web Application</option>
            <option value="monorepo">Polyglot Monorepo</option>
            <option value="staged-hook">Staged Git Hook Project</option>
            <option value="offline">Strict Offline Workspace</option>
            <option value="strict-ci">Strict CI Security Gate</option>
            <option value="external-scanner">External Scanner Suite</option>
            <option value="gradual-adoption">Gradual Adoption & Baseline</option>
          </select>
          <button type="button" class="btn-secondary" @click="resetToPreset" :disabled="!isDirty">
            Reset to Preset
          </button>
        </div>
      </div>
      <p id="preset-desc" class="field-help">{{ presetDescriptions[selectedPreset] }}</p>
      <div v-if="isDirty" class="dirty-badge" role="note">
        ✏️ Custom edits applied (differs from base preset)
      </div>
    </div>

    <!-- Form Configuration Grid -->
    <div class="form-grid">
      <!-- General Settings -->
      <div class="form-group">
        <label for="project-name" class="field-label">Project Name <span class="required">*</span></label>
        <input 
          id="project-name" 
          v-model="config.project" 
          type="text" 
          class="input-control" 
          placeholder="my-project" 
          @input="markDirty"
          aria-required="true"
        />
      </div>

      <div class="form-group">
        <label for="project-root" class="field-label">Root Directory</label>
        <input 
          id="project-root" 
          v-model="config.root" 
          type="text" 
          class="input-control" 
          placeholder="." 
          @input="markDirty"
        />
      </div>

      <div class="form-group">
        <label for="scan-mode" class="field-label">Default Scan Mode</label>
        <select id="scan-mode" v-model="config.mode" class="input-control select-control" @change="markDirty">
          <option value="full">full (all recognized files)</option>
          <option value="changed">changed (relative to HEAD)</option>
          <option value="staged">staged (git index snapshot)</option>
        </select>
      </div>

      <div class="form-group">
        <label for="fail-on" class="field-label">Fail On Severity Threshold</label>
        <select id="fail-on" v-model="config.fail_on" class="input-control select-control" @change="markDirty">
          <option value="CRITICAL">CRITICAL</option>
          <option value="HIGH">HIGH</option>
          <option value="MEDIUM">MEDIUM</option>
          <option value="LOW">LOW</option>
        </select>
      </div>

      <div class="form-group">
        <label for="workers" class="field-label">Worker Concurrency (1–64)</label>
        <input 
          id="workers" 
          v-model.number="config.workers" 
          type="number" 
          min="1" 
          max="64" 
          class="input-control" 
          @input="markDirty"
        />
      </div>

      <div class="form-group">
        <label for="output-path" class="field-label">Report Output Path</label>
        <input 
          id="output-path" 
          v-model="config.output" 
          type="text" 
          class="input-control" 
          placeholder="security_findings.json" 
          @input="markDirty"
        />
      </div>
    </div>

    <!-- Validation Error Alert -->
    <div v-if="validationErrors.length > 0" class="validation-box" role="alert">
      <strong>Configuration errors:</strong>
      <ul>
        <li v-for="(err, idx) in validationErrors" :key="idx">{{ err }}</li>
      </ul>
    </div>

    <!-- Output & Actions Section -->
    <div class="output-container">
      <div class="output-header">
        <div class="output-title">
          <h3>Generated <code>security-review.json</code></h3>
          <span class="badge">Schema v1</span>
        </div>
        <div class="actions">
          <button 
            type="button" 
            class="btn-action btn-copy" 
            @click="copyJSON" 
            :disabled="validationErrors.length > 0"
            aria-label="Copy configuration JSON to clipboard"
          >
            {{ copyStatusText }}
          </button>
          <button 
            type="button" 
            class="btn-action btn-download" 
            @click="downloadJSON" 
            :disabled="validationErrors.length > 0"
            aria-label="Download security-review.json file"
          >
            {{ downloadStatusText }}
          </button>
        </div>
      </div>

      <pre class="json-output" tabindex="0" aria-label="Generated JSON configuration code block"><code>{{ jsonOutput }}</code></pre>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { presetDescriptions, presets, validateConfig } from './ConfigBuilderData.js'

const selectedPreset = ref('minimal')
const activePreset = ref('minimal')
const isDirty = ref(false)
const copied = ref(false)
const downloaded = ref(false)
const statusAnnouncement = ref('')

const config = ref(JSON.parse(JSON.stringify(presets.minimal)))

function markDirty() {
  isDirty.value = true
}

function onPresetChange() {
  const nextPreset = selectedPreset.value
  if (!presets[nextPreset]) return
  if (isDirty.value && !window.confirm('Changing presets will replace your unsaved edits. Continue?')) {
    selectedPreset.value = activePreset.value
    announce('Preset change canceled; custom edits were preserved.')
    return
  }
  activePreset.value = nextPreset
  config.value = JSON.parse(JSON.stringify(presets[nextPreset]))
  isDirty.value = false
  announce(`Loaded ${nextPreset} preset`)
}

function resetToPreset() {
  if (presets[selectedPreset.value]) {
    config.value = JSON.parse(JSON.stringify(presets[selectedPreset.value]))
    isDirty.value = false
    announce(`Reset configuration to ${selectedPreset.value} preset`)
  }
}

const validationErrors = computed(() => validateConfig(config.value))

const jsonOutput = computed(() => JSON.stringify(config.value, null, 2))

const copyStatusText = computed(() => (copied.value ? 'Copied to Clipboard! ✓' : 'Copy JSON'))
const downloadStatusText = computed(() => (downloaded.value ? 'Downloaded! ✓' : 'Download JSON'))

function announce(message) {
  statusAnnouncement.value = message
  setTimeout(() => {
    if (statusAnnouncement.value === message) {
      statusAnnouncement.value = ''
    }
  }, 4000)
}

async function copyJSON() {
  if (validationErrors.value.length > 0) return
  try {
    await navigator.clipboard.writeText(jsonOutput.value)
    copied.value = true
    announce('Configuration JSON copied to clipboard successfully.')
    setTimeout(() => {
      copied.value = false
    }, 2500)
  } catch (err) {
    announce('Failed to copy to clipboard.')
  }
}

function downloadJSON() {
  if (validationErrors.value.length > 0) return
  try {
    const blob = new Blob([jsonOutput.value], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'security-review.json'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    downloaded.value = true
    announce('security-review.json file downloaded successfully.')
    setTimeout(() => {
      downloaded.value = false
    }, 2500)
  } catch (err) {
    announce('Failed to download configuration file.')
  }
}
</script>

<style scoped>
.config-builder {
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  padding: 24px;
  background: var(--vp-c-bg-soft);
  margin-top: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.builder-header h2 {
  margin-top: 0;
  margin-bottom: 8px;
  font-size: 1.4rem;
}

.builder-header p {
  color: var(--vp-c-text-2);
  margin-bottom: 20px;
}

.builder-header .validation-scope {
  font-size: 0.9rem;
  margin-top: -12px;
}

.control-section {
  background: var(--vp-c-bg-alt);
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 24px;
  border: 1px solid var(--vp-c-divider);
}

.preset-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.preset-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.field-label {
  font-weight: 600;
  color: var(--vp-c-text-1);
}

.field-label .required {
  color: var(--vp-c-danger-1, #e11d48);
}

.field-help {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  margin-top: 8px;
  margin-bottom: 0;
}

.dirty-badge {
  font-size: 0.8rem;
  color: var(--vp-c-brand-1);
  margin-top: 8px;
  font-weight: 500;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  font-size: 0.9rem;
  margin-bottom: 6px;
}

.input-control {
  padding: 8px 12px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-size: 0.9rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.input-control:focus {
  outline: none;
  border-color: var(--vp-c-brand-1);
  box-shadow: 0 0 0 3px var(--vp-c-brand-soft, rgba(16, 185, 129, 0.15));
}

.select-control {
  cursor: pointer;
}

.validation-box {
  background: var(--vp-custom-block-danger-bg, #fef2f2);
  border: 1px solid var(--vp-custom-block-danger-border, #fecaca);
  color: var(--vp-custom-block-danger-text, #991b1b);
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 0.9rem;
}

.validation-box ul {
  margin: 6px 0 0 16px;
  padding: 0;
}

.output-container {
  border-top: 1px solid var(--vp-c-divider);
  padding-top: 20px;
}

.output-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 12px;
}

.output-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.output-title h3 {
  margin: 0;
  font-size: 1.1rem;
}

.badge {
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 12px;
  background: var(--vp-c-brand-soft, rgba(16, 185, 129, 0.15));
  color: var(--vp-c-brand-1);
  font-weight: 600;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn-action {
  padding: 8px 16px;
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  border: none;
  transition: opacity 0.2s, background-color 0.2s;
}

.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-copy {
  background: var(--vp-c-brand-1);
  color: #ffffff;
}

.btn-copy:hover:not(:disabled) {
  background: var(--vp-c-brand-2);
}

.btn-download {
  background: var(--vp-c-bg-alt);
  border: 1px solid var(--vp-c-divider);
  color: var(--vp-c-text-1);
}

.btn-download:hover:not(:disabled) {
  background: var(--vp-c-bg-soft);
}

.btn-secondary {
  padding: 6px 12px;
  font-size: 0.85rem;
  border-radius: 6px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-2);
  cursor: pointer;
}

.btn-secondary:hover:not(:disabled) {
  color: var(--vp-c-text-1);
  border-color: var(--vp-c-text-3);
}

.btn-secondary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.json-output {
  background: var(--vp-c-bg-alt);
  padding: 16px;
  border-radius: 8px;
  border: 1px solid var(--vp-c-divider);
  overflow-x: auto;
  font-family: var(--vp-font-family-mono);
  font-size: 0.85rem;
  line-height: 1.5;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
