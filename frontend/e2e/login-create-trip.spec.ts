import { test, expect } from '@playwright/test'

// Critical-path smoke against the full stack (nginx-served frontend + /api proxy
// + backend + seeded db). Proves auth round-trips through the proxy and the
// two-step trip planner persists a trip end to end.
//
// Requires E2E_USER / E2E_PASS (the backend's APP_USERNAME / APP_PASSWORD) and a
// running stack reachable at E2E_BASE_URL. Skipped otherwise so local `npm run
// e2e` without a stack is a no-op rather than a failure.
const user = process.env.E2E_USER
const pass = process.env.E2E_PASS

test.skip(!user || !pass, 'Set E2E_USER / E2E_PASS and run against a live stack (see playwright.config.ts)')

test('logs in and creates a trip end to end', async ({ page }) => {
  // 1. Authenticate — proves login → cookie → /api proxy → SPA redirect.
  await page.goto('/login')
  await page.getByLabel(/username/i).fill(user!)
  await page.getByLabel(/password/i).fill(pass!)
  await page.getByRole('button', { name: /sign in/i }).click()
  await expect(page).toHaveURL(/\/trips$/)

  // 2. Open the planner and fill step 1 (details + a seeded person).
  await page.goto('/trips/new')
  await page.getByPlaceholder('Weekend in the Alps').fill('smoke test trip')
  // dev/test-data.sql (loaded by the e2e workflow) provides persons; pick the first.
  await page.locator('[data-element="trip-planner-people"] button').first().click()

  // 3. Advance to step 2 (the "→" button enables once name + person are set).
  await page.getByRole('button', { name: '→' }).click()

  // 4. Save — on success the planner redirects back to the trips list.
  await page.getByRole('button', { name: 'Save trip' }).click()
  await expect(page).toHaveURL(/\/trips$/)
  await expect(page.getByText('Smoke Test Trip')).toBeVisible() // normalizeTitleWords
})
