import { z } from 'zod'
import { buildTypedSchema } from './schema'

describe('buildTypedSchema', () => {
  it('wraps a zod schema into a vee-validate typed schema', async () => {
    const typed = buildTypedSchema(z.object({ name: z.string().min(1, 'Name is required') }))
    expect(typed).toBeDefined()
    expect(typeof typed.parse).toBe('function')

    // Valid input parses; invalid input surfaces the zod message.
    const okResult = await typed.parse({ name: 'Alice' })
    expect(okResult.value).toEqual({ name: 'Alice' })

    const badResult = await typed.parse({ name: '' })
    expect(badResult.errors).toContainEqual(
      expect.objectContaining({ path: 'name', errors: ['Name is required'] }),
    )
  })
})
