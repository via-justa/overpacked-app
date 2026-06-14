import { ensureApiResponse, unwrapApiResponse } from './request'

const ok = <T>(data: T) => Promise.resolve({ data, response: new Response(null, { status: 200 }) })
const fail = (status: number, error?: unknown) =>
  Promise.resolve({ error, response: new Response(null, { status }) })

describe('unwrapApiResponse', () => {
  it('returns the data on a successful response', async () => {
    await expect(unwrapApiResponse(ok([1, 2, 3]), 'fb')).resolves.toEqual([1, 2, 3])
  })

  it('throws the API error message when the response failed', async () => {
    await expect(unwrapApiResponse(fail(500, { error: 'boom' }), 'fb')).rejects.toThrow('boom')
  })

  it('throws the fallback when the body is missing', async () => {
    await expect(unwrapApiResponse(ok(null), 'fallback msg')).rejects.toThrow('fallback msg')
  })
})

describe('ensureApiResponse', () => {
  it('resolves for a no-content success', async () => {
    await expect(
      ensureApiResponse(Promise.resolve({ response: new Response(null, { status: 204 }) }), 'fb'),
    ).resolves.toBeUndefined()
  })

  it('throws for a failed response', async () => {
    await expect(ensureApiResponse(fail(403, { message: 'nope' }), 'fb')).rejects.toThrow('nope')
  })
})
