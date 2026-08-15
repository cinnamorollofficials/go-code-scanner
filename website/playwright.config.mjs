import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  reporter: 'line',
  use: {
    baseURL: 'http://127.0.0.1:4173/go-code-scanner/',
    channel: 'chrome',
    headless: true
  },
  webServer: {
    command: 'npm run docs:preview -- --host 127.0.0.1 --port 4173',
    url: 'http://127.0.0.1:4173/go-code-scanner/',
    reuseExistingServer: true,
    timeout: 30000
  }
})
