import { http, HttpResponse } from 'msw'
import { useSettings } from './useSettings'
import { withSetup } from '../test/withSetup'
import { server } from '../test/msw/server'
import { settingsFixture } from '../test/fixtures'

describe('useSettings', () => {
  it('exposes defaults before the query resolves', () => {
    const { result, unmount } = withSetup(() => useSettings())
    expect(result.weightUnit.value).toBe('g')
    expect(result.volumeUnit.value).toBe('ml')
    expect(result.currency.value).toBe('usd')
    unmount()
  })

  it('reflects loaded settings (bare object body, no { data } envelope)', async () => {
    server.use(
      http.get('*/api/v1/settings', () =>
        // NOTE: the API returns the Settings object directly. A { data: ... }
        // wrapper here would leave weightUnit on its default and fail the assertion.
        HttpResponse.json(settingsFixture({ weight_unit: 'oz', volume_unit: 'fl_oz', currency: 'eur' })),
      ),
    )

    const { result, unmount } = withSetup(() => useSettings())
    await vi.waitFor(() => {
      expect(result.weightUnit.value).toBe('oz')
    })
    expect(result.volumeUnit.value).toBe('fl_oz')
    expect(result.currency.value).toBe('eur')
    unmount()
  })
})
