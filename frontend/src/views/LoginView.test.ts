import { fireEvent, waitFor } from '@testing-library/vue'
import { http, HttpResponse } from 'msw'
import LoginView from './LoginView.vue'
import { renderWithProviders } from '../test/renderWithProviders'
import { server } from '../test/msw/server'

describe('LoginView', () => {
  it('signs in with valid credentials and redirects to the target route', async () => {
    server.use(
      http.post('*/api/v1/auth/login', () =>
        HttpResponse.json({ access_token: 't', token_type: 'Bearer', expires_in: 900 }),
      ),
    )

    const { getByLabelText, getByRole, router } = renderWithProviders(LoginView, {
      initialRoute: '/login',
    })
    await router.isReady()

    await fireEvent.update(getByLabelText(/username/i), 'alice')
    await fireEvent.update(getByLabelText(/password/i), 'secret')

    const submit = getByRole('button', { name: /sign in/i })
    await waitFor(() => expect(submit).not.toBeDisabled())
    await fireEvent.click(submit)

    await waitFor(() => expect(router.currentRoute.value.path).toBe('/trips'))
  })

  it('surfaces a server error and stays on the login page', async () => {
    server.use(
      http.post('*/api/v1/auth/login', () =>
        HttpResponse.json({ error: 'Invalid credentials' }, { status: 401 }),
      ),
    )

    const { getByLabelText, getByRole, findByText, router } = renderWithProviders(LoginView, {
      initialRoute: '/login',
    })
    await router.isReady()

    await fireEvent.update(getByLabelText(/username/i), 'alice')
    await fireEvent.update(getByLabelText(/password/i), 'wrong')

    const submit = getByRole('button', { name: /sign in/i })
    await waitFor(() => expect(submit).not.toBeDisabled())
    await fireEvent.click(submit)

    expect(await findByText(/invalid credentials/i)).toBeInTheDocument()
    expect(router.currentRoute.value.path).toBe('/login')
  })
})
