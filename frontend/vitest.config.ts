import { mergeConfig, defineConfig } from 'vitest/config'
import viteConfig from './vite.config'

// Reuse the app's Vite pipeline (plugin-vue, TS resolution, import.meta.env) so
// tests transform source exactly like the dev/build does. Test-only settings live
// here, keeping app builds free of test concerns.
export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      globals: true,
      environment: 'happy-dom',
      // Pin the timezone so date formatting/age math is deterministic across
      // developer machines and the (UTC) CI runner.
      env: { TZ: 'UTC' },
      setupFiles: ['./src/test/setup.ts'],
      // Unit/component specs only. Playwright E2E lives in e2e/ as *.spec.ts.
      include: ['src/**/*.test.ts'],
      css: false,
      clearMocks: true,
      coverage: {
        provider: 'v8',
        reporter: ['text', 'text-summary', 'html', 'lcov'],
        reportsDirectory: './coverage',
        // Scope coverage to the modules under test in the current rollout phase so
        // the threshold gate is meaningful (not diluted by hundreds of as-yet
        // untested components). Widen this list as each phase lands; SonarQube's
        // new-code quality gate enforces coverage on everything a PR touches.
        include: [
          'src/lib/units/**',
          'src/lib/format/**',
          'src/lib/text/**',
          'src/lib/navigation/**',
          'src/lib/validation/**',
          'src/lib/storage/**',
          'src/lib/api/request.ts',
          'src/lib/api/errors.ts',
          'src/lib/api/client.ts',
          'src/features/**/utils.ts',
          'src/features/**/utils/**',
          'src/features/**/schema.ts',
          'src/features/trips/api/tripPersistence.ts',
          'src/composables/**',
          'src/stores/auth.ts',
        ],
        exclude: ['src/**/*.d.ts', 'src/test/**', '**/*.config.*'],
        // Ratchet up each phase toward the ≥80% project goal. These mirror the
        // SonarQube gate for fast local/CI feedback; Sonar remains authoritative.
        thresholds: {
          lines: 90,
          functions: 90,
          statements: 90,
          branches: 85,
        },
      },
    },
  }),
)
