import { fireEvent } from '@testing-library/vue'
import { MOBILE_BREAKPOINT, useIsMobile } from './useIsMobile'
import { withSetup } from '../test/withSetup'

const setWidth = (width: number) => {
  Object.defineProperty(globalThis.window, 'innerWidth', { configurable: true, value: width })
}

describe('useIsMobile', () => {
  it('is true below the md breakpoint and tracks resize events', async () => {
    setWidth(MOBILE_BREAKPOINT - 1)
    const { result, unmount } = withSetup(() => ({ isMobile: useIsMobile() }))
    expect(result.isMobile.value).toBe(true)

    setWidth(MOBILE_BREAKPOINT + 200)
    await fireEvent(globalThis.window, new Event('resize'))
    expect(result.isMobile.value).toBe(false)

    unmount()
  })
})
