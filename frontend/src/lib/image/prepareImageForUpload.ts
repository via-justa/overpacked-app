// Downscales and re-encodes a picked image entirely in the browser so we never
// store oversized originals in the database. Best-effort: if the browser can't
// decode/encode the file, the original bytes are returned unchanged.

export type PreparedImage = {
  // base64-encoded bytes WITHOUT the `data:...;base64,` prefix.
  base64: string
  mimeType: string
  sizeBytes: number
}

// Longest edge (px) the stored image is scaled down to; smaller images are kept
// as-is (never upscaled). Item photos are only shown in small cards/previews.
const DEFAULT_MAX_DIMENSION = 1024
const OUTPUT_MIME = 'image/jpeg'
const OUTPUT_QUALITY = 0.82

export async function prepareImageForUpload(
  file: File,
  maxDimension = DEFAULT_MAX_DIMENSION,
): Promise<PreparedImage> {
  try {
    return await resizeToJpeg(file, maxDimension)
  } catch {
    // Decoding/encoding unsupported for this file — fall back to the original.
    return { base64: await blobToBase64(file), mimeType: file.type, sizeBytes: file.size }
  }
}

async function resizeToJpeg(file: File, maxDimension: number): Promise<PreparedImage> {
  // `from-image` honours EXIF orientation so portrait photos aren't rotated.
  const bitmap = await createImageBitmap(file, { imageOrientation: 'from-image' })
  try {
    const scale = Math.min(1, maxDimension / Math.max(bitmap.width, bitmap.height))
    const width = Math.max(1, Math.round(bitmap.width * scale))
    const height = Math.max(1, Math.round(bitmap.height * scale))

    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext('2d')
    if (!ctx) {
      throw new Error('canvas 2d context unavailable')
    }
    ctx.drawImage(bitmap, 0, 0, width, height)

    const blob = await canvasToBlob(canvas, OUTPUT_MIME, OUTPUT_QUALITY)
    return { base64: await blobToBase64(blob), mimeType: OUTPUT_MIME, sizeBytes: blob.size }
  } finally {
    bitmap.close()
  }
}

function canvasToBlob(canvas: HTMLCanvasElement, type: string, quality: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => (blob ? resolve(blob) : reject(new Error('failed to encode image'))),
      type,
      quality,
    )
  })
}

function blobToBase64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      resolve(result.includes(',') ? result.split(',')[1] ?? '' : '')
    }
    reader.onerror = () => reject(reader.error ?? new Error('failed to read image'))
    reader.readAsDataURL(blob)
  })
}
