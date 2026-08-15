import { expect, test } from '@playwright/test'
import axe from 'axe-core'

const criticalPages = [
  ['', 'Go Code Scanner'],
  ['getting-started/first-scan', 'First Scan and Exit Codes'],
  ['reference/cli', 'CLI Command Reference'],
  ['reference/rule-catalog', 'Rule Catalog'],
  ['reference/config-builder', 'Config Builder']
]

async function accessibilityViolations(page) {
  await page.addScriptTag({ content: axe.source })
  const result = await page.evaluate(async () => window.axe.run(document, {
    runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa'] }
  }))
  return result.violations.filter((violation) => ['critical', 'serious'].includes(violation.impact))
}

async function tabTo(page, locator, limit = 60) {
  for (let index = 0; index < limit; index++) {
    await page.keyboard.press('Tab')
    if (await locator.evaluate((element) => element === document.activeElement)) return
  }
  throw new Error(`Could not reach ${await locator.evaluate((element) => element.outerHTML)} with the keyboard after ${limit} Tab presses`)
}

for (const [path, heading] of criticalPages) {
  test(`${heading} has one H1 and no serious accessibility violations`, async ({ page }) => {
    await page.goto(path)
    await expect(page.locator('h1')).toHaveCount(1)
    await expect(page.locator('h1')).toContainText(heading)
    await expect(page.locator('link[rel="canonical"]')).toHaveCount(1)
    await expect(page.locator('meta[property="og:image"]')).toHaveAttribute('content', /\/social-card\.png$/)
    await expect(page.locator('meta[name="twitter:card"]')).toHaveAttribute('content', 'summary_large_image')
    const violations = await accessibilityViolations(page)
    expect(violations.length, violations.map((violation) => `${violation.id}: ${violation.help} (${violation.nodes.length} elements)`).join('\n')).toBe(0)
  })
}

test('Rule Catalog custom controls pass accessibility checks in dark mode', async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem('vitepress-theme-appearance', 'dark'))
  await page.goto('reference/rule-catalog')
  await expect(page.locator('html')).toHaveClass(/dark/)
  const violations = await accessibilityViolations(page)
  expect(violations.length, violations.map((violation) => `${violation.id}: ${violation.help} (${violation.nodes.length} elements)`).join('\n')).toBe(0)
})

test('Rule Catalog filters and links to focused guidance', async ({ page }) => {
  await page.goto('reference/rule-catalog')
  await expect(page.getByRole('status')).toContainText('Showing 71 of 71 rules')
  await page.getByRole('combobox', { name: 'Domain', exact: true }).selectOption('security')
  await expect(page.getByRole('status')).toContainText('Showing 40 of 71 rules')
  await page.getByLabel('Rule ID or description').fill('mock-token')
  await expect(page.locator('tbody tr')).toHaveCount(1)
  await page.getByRole('link', { name: 'mock-token' }).click()
  await expect(page.locator('h1')).toContainText('mock-token')
  await expect(page.getByRole('link', { name: /browser-token-storage/ })).toBeVisible()
})

test('generated reference pages provide breadcrumbs, canonical metadata, and feedback', async ({ page }) => {
  await page.goto('reference/rules/mock-token')
  const breadcrumb = page.getByRole('navigation', { name: 'Breadcrumb' })
  await expect(breadcrumb.getByRole('link', { name: 'Reference', exact: true })).toHaveAttribute('href', '/go-code-scanner/reference/')
  await expect(breadcrumb.getByRole('link', { name: 'Rule Catalog' })).toHaveAttribute('href', '/go-code-scanner/reference/rule-catalog')
  await expect(breadcrumb.getByText('mock-token rule', { exact: true })).toHaveAttribute('aria-current', 'page')
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://cinnamorollofficials.github.io/go-code-scanner/reference/rules/mock-token')
  await expect(page.getByRole('link', { name: 'Report a documentation issue' })).toHaveAttribute('href', /issues\/new\?title=Documentation%3A%20mock-token%20rule/)
})

for (const width of [320, 768, 1440]) {
  test(`Rule Catalog has no page-level horizontal overflow at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('reference/rule-catalog')
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth)
    const offenders = await page.evaluate(() => [...document.querySelectorAll('*')]
      .filter((element) => element.getBoundingClientRect().right > document.documentElement.clientWidth + 1)
      .slice(0, 8)
      .map((element) => {
        const rect = element.getBoundingClientRect()
        return `${element.tagName}.${element.className}: left=${Math.round(rect.left)} right=${Math.round(rect.right)} width=${Math.round(rect.width)} client=${element.clientWidth} scroll=${element.scrollWidth}`
      }))
    expect(overflow, offenders.join('\n')).toBeLessThanOrEqual(1)
    await expect(page.locator('.table-wrap')).toBeVisible()
  })
}

test('Config Builder preserves edits when preset replacement is canceled', async ({ page }) => {
  await page.goto('reference/config-builder')
  await page.getByLabel('Project Name').fill('edited-project')
  page.once('dialog', (dialog) => dialog.dismiss())
  await page.getByLabel('Configuration Preset').selectOption('go-service')
  await expect(page.getByLabel('Configuration Preset')).toHaveValue('minimal')
  await expect(page.getByLabel('Project Name')).toHaveValue('edited-project')
})

test('keyboard focus starts with the skip link', async ({ page }) => {
  await page.goto('getting-started/first-scan')
  await page.keyboard.press('Tab')
  await expect(page.locator(':focus')).toHaveAttribute('href', '#VPContent')
})

test('header search and mobile navigation work from the keyboard', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 900 })
  await page.goto('getting-started/first-scan')

  const search = page.getByRole('button', { name: 'Search' })
  await tabTo(page, search)
  await page.keyboard.press('Enter')
  await expect(page.locator('.VPLocalSearchBox')).toBeVisible()
  await page.keyboard.press('Escape')

  const menu = page.locator('.VPNavBarHamburger')
  await tabTo(page, menu)
  await page.keyboard.press('Enter')
  await expect(page.locator('.VPNavScreen')).toBeVisible()
})

test('sidebar, table, builder, and code-copy controls are keyboard reachable', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 })
  await page.goto('reference/rule-catalog')

  const sidebarLink = page.locator('.VPSidebar a[href$="/reference/cli"]').first()
  await tabTo(page, sidebarLink)
  await expect(sidebarLink).toBeFocused()

  const table = page.locator('.table-wrap')
  await tabTo(page, table)
  await expect(table).toBeFocused()

  await page.goto('reference/config-builder')
  const projectName = page.getByLabel('Project Name')
  await tabTo(page, projectName)
  await page.keyboard.type('-keyboard')
  await expect(projectName).toHaveValue(/-keyboard$/)
  await expect(page.locator('[aria-live="polite"]')).not.toHaveCount(0)

  await page.goto('getting-started/first-scan')
  const copy = page.getByRole('button', { name: 'Copy Code' }).first()
  await tabTo(page, copy)
  await expect(copy).toBeFocused()
})
