import { describe, expect, it, vi } from 'vitest'

const { push } = vi.hoisted(() => ({ push: vi.fn() }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))

import { useActionRouter } from './useActionRouter'

describe('useActionRouter', () => {
  it('invokes the logout callback and does not navigate for the logout target', async () => {
    const onLogout = vi.fn()
    await useActionRouter(onLogout).selectAction('logout')

    expect(onLogout).toHaveBeenCalledOnce()
    expect(push).not.toHaveBeenCalled()
  })

  it('navigates to the mapped route for a navigable target', async () => {
    const onLogout = vi.fn()
    await useActionRouter(onLogout).selectAction('add-item')

    expect(push).toHaveBeenCalledWith({ path: '/gear', query: { action: 'create-item' } })
    expect(onLogout).not.toHaveBeenCalled()
  })
})
