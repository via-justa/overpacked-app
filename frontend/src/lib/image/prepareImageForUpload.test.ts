import { afterEach, describe, expect, it, vi } from 'vitest'
import { prepareImageForUpload } from './prepareImageForUpload'

const makeFile = (bytes: string, type: string) => new File([bytes], 'photo', { type })

// Stub a canvas whose getContext/toBlob behaviour the test controls.
const stubCanvas = (overrides: Partial<Record<'getContext' | 'toBlob', unknown>>) => {
  vi.spyOn(document, 'createElement').mockReturnValue({
    width: 0,
    height: 0,
    getContext: () => ({ drawImage: vi.fn() }),
    toBlob: (cb: (b: Blob | null) => void) => cb(new Blob(['jpeg-bytes'], { type: 'image/jpeg' })),
    ...overrides,
  } as unknown as HTMLCanvasElement)
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('prepareImageForUpload', () => {
  it('downscales large images to JPEG and strips the data-URL prefix', async () => {
    const close = vi.fn()
    vi.stubGlobal('createImageBitmap', vi.fn().mockResolvedValue({ width: 2000, height: 1000, close }))
    stubCanvas({})

    const result = await prepareImageForUpload(makeFile('orig', 'image/png'), 1024)

    expect(result.mimeType).toBe('image/jpeg')
    expect(result.sizeBytes).toBeGreaterThan(0)
    expect(result.base64.length).toBeGreaterThan(0)
    expect(result.base64).not.toContain(',') // prefix stripped
    expect(close).toHaveBeenCalledOnce()
  })

  it('falls back to the original bytes when the browser cannot decode the image', async () => {
    vi.stubGlobal('createImageBitmap', vi.fn().mockRejectedValue(new Error('decode unsupported')))

    const file = makeFile('rawbytes', 'image/webp')
    const result = await prepareImageForUpload(file)

    expect(result.mimeType).toBe('image/webp')
    expect(result.sizeBytes).toBe(file.size)
    expect(result.base64.length).toBeGreaterThan(0)
  })

  it('falls back when no 2D canvas context is available', async () => {
    vi.stubGlobal('createImageBitmap', vi.fn().mockResolvedValue({ width: 10, height: 10, close: vi.fn() }))
    stubCanvas({ getContext: () => null })

    const result = await prepareImageForUpload(makeFile('x', 'image/png'))

    expect(result.mimeType).toBe('image/png') // original, not re-encoded
  })

  it('falls back when the canvas fails to encode', async () => {
    vi.stubGlobal('createImageBitmap', vi.fn().mockResolvedValue({ width: 10, height: 10, close: vi.fn() }))
    stubCanvas({ toBlob: (cb: (b: Blob | null) => void) => cb(null) })

    const result = await prepareImageForUpload(makeFile('y', 'image/gif'))

    expect(result.mimeType).toBe('image/gif') // original, not re-encoded
  })
})
