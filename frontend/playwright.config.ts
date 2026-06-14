import { defineConfig, devices } from '@playwright/test'

// E2E runs against a real, same-origin stack (the deployment docker-compose:
// nginx-served frontend that proxies /api to the backend). Point at it with
// E2E_BASE_URL; it defaults to the compose frontend port (8080:80).
//
// There is intentionally no auto-started webServer: `vite preview` does NOT
// proxy /api (that proxy is dev-only in vite.config.ts), so E2E needs the full
// stack. Bring it up with `docker compose -f deployment/docker-compose.yaml up`
// (locally or in CI) before running `npm run e2e`.
const baseURL = process.env.E2E_BASE_URL ?? 'http://localhost:8080'

export default defineConfig({
  testDir: './e2e',
  testMatch: '**/*.spec.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI ? [['github'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
